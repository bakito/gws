package gcloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/gorilla/websocket"

	"github.com/bakito/gws/pkg/types"
)

type tunnel struct {
	headers http.Header
	wsName  string
	wsHost  string
	client  *workstations.Client
}

func TCPTunnel(cfg *types.Config, port int) error {
	sshContext, ctx, c, ws, err := setup(cfg)
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

	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p)))
	if err != nil {
		_, _ = fmt.Printf("Failed to start TCP listener: %v\n", err)
		return err
	}
	defer closeIt(listener)

	_, _ = fmt.Printf("Listening on local ssh port %d ...\n", p)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			_, _ = fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		_, _ = fmt.Println("Accepted TCP connection")

		// Handle the connection in a separate goroutine
		go t.handleConnection(clientConn)
	}
}

func (t *tunnel) connectWebsocket() (*websocket.Conn, error) {
	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", t.wsHost, 22)
	// Establish persistent WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, t.headers)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			defer closeIt(resp.Body)
			_, _ = fmt.Println(string(body))
		}
		_, _ = fmt.Printf("Failed to connect to WebSocket %q: %v\n", wsURL, err)
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
					_, _ = fmt.Printf("Error reading from TCP connection: %v\n", err)
				}
				return
			}

			// Send TCP data over WebSocket
			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				_, _ = fmt.Printf("Error sending data over WebSocket: %v\n", err)
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
				_, _ = fmt.Println("Connection closed")
			} else {
				_, _ = fmt.Printf("Error reading from WebSocket: %v\n", err)
			}
			return
		}

		// Send WebSocket data to the TCP client
		_, err = clientConn.Write(msg)
		if err != nil {
			_, _ = fmt.Printf("Error writing to TCP connection: %v\n", err)
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
		_, _ = fmt.Printf("Error generating token: %v\n", err)
		return
	}
	t.headers["Authorization"] = []string{"Bearer " + tr.GetAccessToken()}
	_, _ = fmt.Println("Got new Token")
}

func closeIt(cl io.Closer) {
	_ = cl.Close()
}
