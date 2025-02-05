package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/oauth2/google"
)

type Workstations struct {
	apiVersion    string
	client        *http.Client
	host          string
	port          int
	tcpTunnelOpen bool
}

func NewWorkstations() *Workstations {
	return &Workstations{
		client:        &http.Client{},
		tcpTunnelOpen: false,
	}
}

func (ws *Workstations) StartWebSocketTunnel(workstationHost string, workstationPort int) {
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
	url := fmt.Sprintf("wss://%s/_workstation/tcp/%d", workstationHost, workstationPort)
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.AccessToken)
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connection established")

	// Handle WebSocket data transfer
	go ws.forwardData(conn)
}

func (ws *Workstations) forwardData(wsConn *websocket.Conn) {
	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Fatal(err)
			break
		}
		// Handle the message (e.g., send to a TCP client connection)
		log.Printf("Received message: %s", message)
	}
}

func main() {
	// Example of using the Workstations struct to start and stop workstations
	ws := NewWorkstations()

	// Start a WebSocket tunnel
	ws.StartWebSocketTunnel("workstation.google.com", 22)
}
