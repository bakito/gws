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

func TcpTunnel(cfg *types.Config, port int) {
	_, ctx, c, err, wsName, ws := setup(cfg)
	if err != nil {
		return
	}
	defer c.Close()

	headers := http.Header{}

	go refreshAuthToken(ctx, c, wsName, headers)
	getAuthToken(ctx, c, wsName, headers)

	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", ws.Host, 22)

	// Establish WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL,
		headers,
	)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err == nil {
			println(string(body))
		}
		fmt.Printf("Failed to connect to WebSocket %q: %v\n", resp.Status, err)
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
		// Accept incoming TCP connections
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		fmt.Println("Accepted TCP connection")

		// Handle the TCP connection with WebSocket forwarding
		go handleConnection(clientConn, conn)
	}
}

// handleConnection forwards data between the TCP client and the WebSocket connection
func handleConnection(clientConn net.Conn, wsConn *websocket.Conn) {
	defer clientConn.Close()

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
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Error reading from WebSocket: %v\n", err)
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

func refreshAuthToken(ctx context.Context, c *workstations.Client, wsName string, headers http.Header) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	// Start an infinite loop to handle the task periodically
	for {
		select {
		case <-ticker.C:
			getAuthToken(ctx, c, wsName, headers)
		case <-ctx.Done():
			// Context is done (cancelled), stop the task execution
			return
		}
	}
}

func getAuthToken(ctx context.Context, c *workstations.Client, wsName string, headers http.Header) {
	tr, err := c.GenerateAccessToken(ctx, &workstationspb.GenerateAccessTokenRequest{Workstation: wsName})
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}
	headers["Authorization"] = []string{"Bearer " + tr.AccessToken}
}
