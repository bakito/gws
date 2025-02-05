package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2/google"
)

func main() {
	// Replace these values with your actual details
	privateKeyPath := "/home/foo/.ssh/id_rsa"
	username := "user"

	// Read the private key
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}

	// Set up OAuth2 authentication
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		log.Fatalf("Failed to find default credentials: %v", err)
	}
	tokenSource := creds.TokenSource

	// Create a WebSocket connection to the IAP TCP forwarding endpoint
	url := "workstation.google.com"
	wsURL := fmt.Sprintf("wss://%s/_workstation/tcp/%d", url, 22)
	wsConfig, err := websocket.NewConfig(wsURL, "https://workstation.google.com")
	if err != nil {
		log.Fatalf("Failed to create WebSocket config: %v", err)
	}

	// Add OAuth2 token to the WebSocket request
	token, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("Failed to get OAuth2 token: %v", err)
	}
	wsConfig.Header.Add("Authorization", "Bearer "+token.AccessToken)

	// Connect to the WebSocket
	wsConn, err := websocket.DialConfig(wsConfig)
	if err != nil {
		log.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer wsConn.Close()

	// Create an SSH client over the WebSocket connection
	sshConn, chans, reqs, err := ssh.NewClientConn(&wsNetConn{wsConn}, "localhost", &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use a more secure method in production
	})
	if err != nil {
		log.Fatalf("Failed to create SSH client connection: %v", err)
	}
	defer sshConn.Close()

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Run a command
	output, err := session.Output("ls -la")
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}

	// Print the output
	fmt.Println(string(output))
}

// wsNetConn wraps a WebSocket connection to implement net.Conn
type wsNetConn struct {
	*websocket.Conn
}

func (c *wsNetConn) Read(b []byte) (int, error) {
	return c.Conn.Read(b)
}

func (c *wsNetConn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c *wsNetConn) Close() error {
	return c.Conn.Close()
}

func (c *wsNetConn) LocalAddr() net.Addr {
	return nil
}

func (c *wsNetConn) RemoteAddr() net.Addr {
	return nil
}

func (c *wsNetConn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *wsNetConn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *wsNetConn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}
