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
	headers http.Header
	wsName  string
	wsHost  string
	client  *workstations.Client
}

func TCPTunnel(ctx context.Context, cfg *types.Config, port int) error {
	sshContext, c, ws, err := setup(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeIt(c)

	t := &tunnel{
		headers: http.Header{},
		wsHost:  ws.GetHost(),
		wsName:  ws.GetName(),
		client:  c,
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
		fmt.Printf("Failed to start TCP listener: %v\n", err)
		return err
	}
	defer closeIt(listener)

	fmt.Printf("🕳️ Opening tunnel to %s and listening on local ssh port %d ...\n", sshContext.GCloud.Name, p)

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
						fmt.Printf("Failed to accept connection: %v\n", err)
					}
					continue
				}
				fmt.Println("🤝 Accepted TCP connection")
				go t.handleConnection(clientConn)
			}
		}
	}()

	if sshContext.KnownHostsFile != "" {
		go updateKnownHosts(sshContext, sshAddress, p)
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

func updateKnownHosts(sshContext *types.Context, address string, port int) {
	if sshContext.KnownHostsFile == "" {
		return
	}
	c, err := ssh.NewClient(address, sshContext.User, sshContext.PrivateKeyFile)
	if err != nil {
		fmt.Println("Error creating ssh client")
		return
	}
	defer c.Close()

	f, err := os.ReadFile(sshContext.KnownHostsFile)
	if err != nil {
		fmt.Printf("Error reading known_hosts %s file: %v\n", sshContext.KnownHostsFile, err)
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
			fmt.Printf("Error writing known_hosts file: %v\n", err)
			return
		}
		fmt.Printf("📝 KnownHosts file %s updated for %s\n", sshContext.KnownHostsFile, linePrefix)
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
			fmt.Println(string(body))
		}
		fmt.Printf("Failed to connect to WebSocket %q: %v\n", wsURL, err)
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

	defer closeIt(clientConn)
	defer closeIt(wsConn)

	// Create a goroutine to send data from TCP client to WebSocket
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := clientConn.Read(buf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Printf("Error reading from TCP connection: %v\n", err)
				}
				return
			}

			// Send TCP data over WebSocket
			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				fmt.Printf("Error sending data over WebSocket: %v\n", err)
				return
			}
		}
	}()

	// Read data from WebSocket and send to the TCP client
	for {
		_, msg, err := wsConn.ReadMessage()
		var ce *websocket.CloseError
		if err != nil {
			if errors.As(err, &ce) {
				fmt.Println("👋 Connection closed")
			} else {
				fmt.Printf("Error reading from WebSocket: %v\n", err)
			}
			return
		}

		// Send WebSocket data to the TCP client
		_, err = clientConn.Write(msg)
		if err != nil {
			fmt.Printf("Error writing to TCP connection: %v\n", err)
			return
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
		fmt.Printf("Error generating token: %v\n", err)
		return
	}
	t.headers["Authorization"] = []string{"Bearer " + tr.GetAccessToken()}
	fmt.Println("🎫 Got new Tunnel Auth Token")
}

func closeIt(cl io.Closer) {
	_ = cl.Close()
}
