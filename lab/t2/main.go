package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/google"
)

func main() {
	// Set up OAuth2 authentication
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		log.Fatalf("Failed to find default credentials: %v", err)
	}
	tokenSource := creds.TokenSource
	// Add OAuth2 token to the WebSocket request
	token, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Failed to get OAuth2 token: %v", err)
	}

	url := "workstation.google.com"
	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", url, 22)

	// Establish WebSocket connection
	println(token.AccessToken)
	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL,
		http.Header{"Authorization": []string{"Bearer " + token.AccessToken}},
	)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		println(string(body))
		log.Fatalf("Failed to connect to WebSocket %q: %v", resp.Status, err)
	}
	defer conn.Close()

	// Start the local TCP server (e.g., on port 22)
	listener, err := net.Listen("tcp", ":22222")
	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	log.Println("Listening on port 22...")

	for {
		// Accept incoming TCP connections
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Println("Accepted TCP connection")

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
				if err != io.EOF {
					log.Printf("Error reading from TCP connection: %v", err)
				}
				return
			}

			// Send TCP data over WebSocket
			if err := wsConn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("Error sending data over WebSocket: %v", err)
				return
			}
		}
	}()

	// Read data from WebSocket and send to the TCP client
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from WebSocket: %v", err)
			return
		}

		// Send WebSocket data to the TCP client
		_, err = clientConn.Write(msg)
		if err != nil {
			log.Printf("Error writing to TCP connection: %v", err)
			return
		}
	}
}
