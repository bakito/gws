package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/google"
)

func main() {
	url := "workstation.google.com"
	websocketURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", url, 22)

	localPort := "localhost:33333" // Local port to bind to
	remotePort := 22               // Remote port on the workstation

	// Start the TCP tunnel
	err := startTCPTunnel(websocketURL, localPort, remotePort)
	if err != nil {
		log.Fatalf("Failed to start TCP tunnel: %v", err)
	}
}

func startTCPTunnel(websocketURL, localPort string, remotePort int) error {
	// Create a local listener
	listener, err := net.Listen("tcp", localPort)
	if err != nil {
		return fmt.Errorf("failed to listen on local port: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on %s\n", localPort)

	for {
		// Accept local connections
		localConn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept local connection: %w", err)
		}
		defer localConn.Close()

		// Start a goroutine to handle the connection
		go func() {
			creds, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
			if err != nil {
				log.Fatalf("Failed to find default credentials: %v", err)
			}
			tokenSource := creds.TokenSource
			// Add OAuth2 token to the WebSocket request
			token, err := tokenSource.Token()
			if err != nil {
				log.Fatalf("Failed to get OAuth2 token: %v", err)
			}

			// Start WebSocket connection to the workstation
			header := http.Header{}
			header.Add("Authorization", "Bearer "+token.AccessToken)

			// Establish a WebSocket connection
			wsConn, resp, err := websocket.DefaultDialer.Dial(websocketURL, header)
			if err != nil {
				log.Printf("Failed to dial WebSocket %q: %v", resp.Status, err)
				return
			}
			defer wsConn.Close()

			// Forward traffic between local and WebSocket connections
			go func() {
				for {
					// Read from the WebSocket and write to the local connection
					_, message, err := wsConn.ReadMessage()
					if err != nil {
						log.Printf("Error reading from WebSocket: %v", err)
						return
					}
					_, err = localConn.Write(message)
					if err != nil {
						log.Printf("Error writing to local connection: %v", err)
						return
					}
				}
			}()

			for {
				// Read from the local connection and write to the WebSocket
				buf := make([]byte, 1024)
				n, err := localConn.Read(buf)
				if err != nil {
					log.Printf("Error reading from local connection: %v", err)
					return
				}
				err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
				if err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					return
				}
			}
		}()
	}
}
