package gcloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/bakito/gws/pkg/types"
	"github.com/gorilla/websocket"
)

type tunnel struct {
	headers http.Header
	wsName  string
	wsHost  string
	client  *workstations.Client
}

func TcpTunnel(cfg *types.Config, port int) {
	_, ctx, c, err, wsName, ws := setup(cfg)
	if err != nil {
		return
	}
	defer c.Close()

	t := &tunnel{
		headers: http.Header{},
		wsHost:  ws.GetHost(),
		wsName:  wsName,
		client:  c,
	}
	go t.refreshAuthToken(ctx)
	t.getAuthToken(ctx)

	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", ws.Host, 22)

	// Establish persistent WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, t.headers)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			fmt.Println(string(body))
		}
		fmt.Printf("Failed to connect to WebSocket %q: %v\n", wsURL, err)
		os.Exit(1)
	}
	defer conn.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Printf("Failed to start TCP listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d ...\n", port)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		fmt.Println("Accepted TCP connection")

		// Handle the connection in a separate goroutine
		go t.handleConnection(clientConn, conn)
	}
}

// handleConnection forwards data between the TCP client and the WebSocket connection
func (t *tunnel) handleConnection(clientConn net.Conn, wsConn *websocket.Conn) {
	defer clientConn.Close()
	errChan := make(chan error, 2)

	// Goroutine to send data from TCP client to WebSocket
	go func() {
		defer func() { errChan <- nil }()
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
				errChan <- err
				return
			}
		}
	}()

	// Goroutine to read data from WebSocket and send to TCP client
	go func() {
		defer func() { errChan <- nil }()
		for {
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Printf("Error reading from WebSocket: %v\n", err)
				}
				errChan <- err
				return
			}

			// Send WebSocket data to the TCP client
			_, err = clientConn.Write(msg)
			if err != nil {
				fmt.Printf("Error writing to TCP connection: %v\n", err)
				errChan <- err
				return
			}
		}
	}()

	<-errChan // Wait for any error
}

func (t *tunnel) refreshAuthToken(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.getAuthToken(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (t *tunnel) getAuthToken(ctx context.Context) {
	tr, err := t.client.GenerateAccessToken(ctx, &workstationspb.GenerateAccessTokenRequest{Workstation: t.wsName})
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}
	t.headers["Authorization"] = []string{"Bearer " + tr.AccessToken}
}
