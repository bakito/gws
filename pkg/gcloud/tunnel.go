package gcloud

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

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

	tr, err := c.GenerateAccessToken(ctx, &workstationspb.GenerateAccessTokenRequest{Workstation: wsName})
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", ws.Host, 22)

	// Establish WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL,
		http.Header{"Authorization": []string{"Bearer " + tr.AccessToken}},
	)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err == nil {
			println(string(body))
		}
		fmt.Printf("Failed to connect to WebSocket %q: %v", resp.Status, err)
		os.Exit(1)
	}
	defer conn.Close()

	// Start the local TCP server (e.g., on port 22)
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Printf("Failed to start TCP listener: %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d ...\n", port)

	for {
		// Accept incoming TCP connections
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v", err)
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
				if errors.Is(err, io.EOF) {
					fmt.Printf("Error reading from TCP connection: %v", err)
				}
				return
			}

			// Send TCP data over WebSocket
			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				fmt.Printf("Error sending data over WebSocket: %v", err)
				return
			}
		}
	}()

	// Read data from WebSocket and send to the TCP client
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading from WebSocket: %v", err)
			return
		}

		// Send WebSocket data to the TCP client
		_, err = clientConn.Write(msg)
		if err != nil {
			fmt.Printf("Error writing to TCP connection: %v", err)
			return
		}
	}
}
