package gcloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/gorilla/websocket"

	"github.com/bakito/gws/internal/ssh"
	"github.com/bakito/gws/internal/types"
)

type tunnel struct {
	headers  http.Header
	wsName   string
	wsHost   string
	client   *workstations.Client
	reporter func(string)
}

func defaultReporter(s string) {
	fmt.Println(s)
}

func TCPTunnel(ctx context.Context, cfg *types.Config, port int, reporter func(string)) error {
	sshContext, c, ws, err := setup(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeIt(c)

	if reporter == nil {
		reporter = defaultReporter
	}

	t := &tunnel{
		headers:  http.Header{},
		wsHost:   ws.GetHost(),
		wsName:   ws.GetName(),
		client:   c,
		reporter: reporter,
	}
	go t.refreshAuthToken(ctx)
	t.setAuthToken(ctx)

	p := sshContext.Port
	if port != 0 {
		p = port
	}

	lc := net.ListenConfig{}
	sshAddress := net.JoinHostPort("127.0.0.1", strconv.Itoa(p))
	listener, err := lc.Listen(ctx, "tcp", sshAddress)
	if err != nil {
		reporter(fmt.Sprintf("🚨 Failed to start TCP listener: %v", err))
		return err
	}
	defer closeIt(listener)

	reporter(fmt.Sprintf("🕳️ Opening tunnel to %s and listening on local ssh port %d ...", sshContext.GCloud.Name, p))

	// Create an error channel to handle errors from goroutines
	errChan := make(chan error, 1)

	// Start accepting connections in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				clientConn, err := listener.Accept()
				if err != nil {
					if !errors.Is(err, net.ErrClosed) {
						reporter(fmt.Sprintf("🚨 Failed to accept connection: %v", err))
					}
					continue
				}
				reporter("🤝 Accepted TCP connection")
				go t.handleConnection(clientConn)
			}
		}
	}()

	if sshContext.KnownHostsFile != "" {
		go updateKnownHosts(sshContext, sshAddress, p, cfg.SSHTimeout(), reporter)
	}

	// Wait for either context cancellation or error
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

func updateKnownHosts(sshContext *types.Context, address string, port int, timeout time.Duration, reporter func(string)) {
	if sshContext.KnownHostsFile == "" {
		return
	}
	c, err := ssh.NewClient(address, sshContext.User, sshContext.PrivateKeyFile, timeout)
	if err != nil {
		reporter(fmt.Sprintf("🚨 Error creating ssh client: %v", err))
		return
	}
	defer c.Close()

	f, err := os.ReadFile(sshContext.KnownHostsFile)
	if err != nil {
		reporter(fmt.Sprintf("🚨 Error reading known_hosts %s file: %v", sshContext.KnownHostsFile, err))
		return
	}

	lines := strings.Split(string(f), "\n")
	found := false
	changed := false
	linePrefix := fmt.Sprintf("[127.0.0.1]:%d", port)
	for i, line := range lines {
		if strings.HasPrefix(line, linePrefix) {
			if line != c.KnownHostsEntry() {
				lines[i] = c.KnownHostsEntry()
				changed = true
			}
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, c.KnownHostsEntry())
		changed = true
	}

	if changed {
		err = os.WriteFile(sshContext.KnownHostsFile, []byte(strings.Join(lines, "\n")), 0o644)
		if err != nil {
			reporter(fmt.Sprintf("🚨 Error writing known_hosts file: %v", err))
			return
		}
		reporter(fmt.Sprintf("📝 KnownHosts file %s updated for %s", sshContext.KnownHostsFile, linePrefix))
	}
}

func (t *tunnel) connectWebsocket() (*websocket.Conn, error) {
	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", t.wsHost, 22)
	// Establish a persistent WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, t.headers)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			defer closeIt(resp.Body)
			t.reporter(string(body))
		}
		t.reporter(fmt.Sprintf("🚨 Failed to connect to WebSocket %q: %v\n", wsURL, err))
		return nil, err
	}
	return conn, nil
}

// handleConnection forwards data between the TCP client and the WebSocket connection.
func (t *tunnel) handleConnection(clientConn net.Conn) {
	wsConn, err := t.connectWebsocket()
	if err != nil {
		return
	}

	// Create a local context to coordinate the shutdown of both goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer closeIt(clientConn)
	defer closeIt(wsConn)

	// Create a goroutine to send data from TCP client to WebSocket
	go func() {
		defer cancel() // Trigger cancel if the TCP client disconnects
		buf := make([]byte, 32*1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := clientConn.Read(buf)
				if err != nil {
					return // EOF or closed connection is handled by the defer cancel()
				}

				// Send TCP data over WebSocket
				if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					return
				}
			}
		}
	}()

	// Read data from WebSocket and send to the TCP client
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				var ce *websocket.CloseError
				if !errors.As(err, &ce) && !errors.Is(err, net.ErrClosed) {
					t.reporter(fmt.Sprintf("🚨 Error reading from WebSocket: %v\n", err))
				}
				return
			}

			// Send WebSocket data to the TCP client
			_, err = clientConn.Write(msg)
			if err != nil {
				// Prevent logging expected errors when the connection is closed or aborted by the host
				if !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "wsasend") {
					t.reporter(fmt.Sprintf("🚨 Error writing to TCP connection: %v\n", err))
				}
				return
			}
		}
	}
}

func (t *tunnel) refreshAuthToken(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.setAuthToken(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (t *tunnel) setAuthToken(ctx context.Context) {
	tr, err := t.client.GenerateAccessToken(ctx, &workstationspb.GenerateAccessTokenRequest{Workstation: t.wsName})
	if err != nil {
		t.reporter(fmt.Sprintf("🚨 Error generating token: %v\n", err))
		return
	}
	t.headers["Authorization"] = []string{"Bearer " + tr.GetAccessToken()}
	t.reporter("🎫 Got new Tunnel Auth Token")
}

func closeIt(cl io.Closer) {
	_ = cl.Close()
}
