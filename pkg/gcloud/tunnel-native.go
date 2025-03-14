package gcloud

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/bakito/gws/pkg/types"
	"golang.org/x/net/websocket"
)

func TcpTunnelNative(cfg *types.Config, port int) {
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
	wsConfig, err := websocket.NewConfig(wsURL, "https://workstation.google.com")
	if err != nil {
		fmt.Printf("Failed to create WebSocket config: %v\n", err)
		os.Exit(1)
	}

	wsConfig.Header.Add("Authorization", "Bearer "+tr.AccessToken)

	// Connect to the WebSocket
	wsConn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		log.Fatalf("Failed to dial WebSocket: %v\n", err)
	}
	defer wsConn.Close()
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
		go handleConnectionNative(clientConn, wsConn)
	}
}

// handleConnection forwards data between the TCP client and the WebSocket connection
func handleConnectionNative(clientConn net.Conn, wsConn *websocket.Conn) {
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
			if _, err := wsConn.Write(buf[:n]); err != nil {
				fmt.Printf("Error sending data over WebSocket: %v\n", err)
				return
			}
		}
	}()

	// Read data from WebSocket and send to the TCP client
	for {
		var msg []byte
		_, err := wsConn.Read(msg)
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
