package ssh

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/bakito/gws/internal/env"
	"github.com/bakito/gws/internal/passwd"
)

func NewClient(addr, user, privateKeyFile string, timeout time.Duration) (Client, error) {
	privateKey, err := os.ReadFile(env.ExpandEnv(privateKeyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the private key
	auth, err := evaluateAuthMethod(privateKey, privateKeyFile)
	if err != nil {
		return nil, err
	}

	// Define SSH connection details
	var knownHostsEntry string
	clientConfig := &ssh.ClientConfig{
		User: user,                   // Remote SSH username
		Auth: []ssh.AuthMethod{auth}, // Auth method
		HostKeyCallback: func(_ string, remote net.Addr, key ssh.PublicKey) error {
			// #nosec G106: Insecure, as we always get a new cert with gcloud
			if tcpAddr, ok := remote.(*net.TCPAddr); ok {
				knownHostsEntry = fmt.Sprintf(
					"[%s]:%d %s %s",
					tcpAddr.IP,
					tcpAddr.Port,
					key.Type(),
					base64.StdEncoding.EncodeToString(key.Marshal()),
				)
			}
			return nil
		},
	}

	var sshClient *ssh.Client
	if timeout > 0 {
		fmt.Printf("⏲  Using ssh client with timeout %s\n", timeout.String())
		clientConfig.Timeout = timeout
		sshClient, err = clientWithTimeout(addr, timeout, clientConfig)
	} else {
		sshClient, err = defaultClient(addr, clientConfig)
	}
	if err != nil {
		return nil, err
	}

	// For other authentication methods see ssh.ClientConfig and ssh.AuthMethod

	// Create a new SCP client
	scpClient := scp.NewClient(addr, clientConfig)

	// Connect to the remote server
	err = scpClient.Connect()
	if err != nil {
		return nil, fmt.Errorf("couldn't establish a connection to the remote server: %w", err)
	}

	return &client{
		sshClient:       sshClient,
		scpClient:       scpClient,
		knownHostsEntry: knownHostsEntry,
	}, nil
}

func clientWithTimeout(addr string, timeout time.Duration, clientConfig *ssh.ClientConfig) (*ssh.Client, error) {
	// Use a dialer with TCP KeepAlive enabled to prevent connection drops
	dialer := net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	// Connect to the SSH server using the existing TCP connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH connection: %w", err)
	}
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	return sshClient, nil
}

func defaultClient(addr string, clientConfig *ssh.ClientConfig) (*ssh.Client, error) {
	// Connect to the SSH server
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	return sshClient, nil
}

type Client interface {
	Close()
	Execute(command string) (output string, err error)
	CopyFile(from, to, permissions string) (err error)
	KnownHostsEntry() string
}

type client struct {
	sshClient       *ssh.Client
	scpClient       scp.Client
	knownHostsEntry string
}

func (c *client) Close() {
	if c.sshClient != nil {
		_ = c.sshClient.Close()
	}

	c.scpClient.Close()
}

func (c *client) KnownHostsEntry() string {
	return c.knownHostsEntry
}

func (c *client) Execute(command string) (string, error) {
	fmt.Printf("Executing ssh %q\n", command)

	// Start a new SSH session
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Execute the command
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}
	return string(output), nil
}

func (c *client) CopyFile(from, to, permissions string) error {
	fmt.Printf("Copy file form %q to %q with perissions %s\n", from, to, permissions)
	// Open a file
	f, _ := os.Open(env.ExpandEnv(from))
	// Close the file after it has been copied
	defer f.Close()

	err := c.scpClient.CopyFromFile(context.Background(), *f, to, permissions)
	if err != nil {
		return fmt.Errorf("error while copying file: %w", err)
	}
	_, err = c.Execute(fmt.Sprintf("chmod %s %s", permissions, to))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func evaluateAuthMethod(privateKey []byte, privateKeyFile string) (ssh.AuthMethod, error) {
	auth, err := getSSHAgentAuthMethod()
	if err != nil {
		return nil, err
	}
	if auth != nil {
		return auth, nil
	}

	// try private key
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		var missingPassphraseErr *ssh.PassphraseMissingError
		if !errors.As(err, &missingPassphraseErr) {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		pass, err := passwd.Prompt(fmt.Sprintf("🔐 Please enter the passphrase for private key (%s):", privateKeyFile))
		if err != nil {
			return nil, err
		}

		passBytes := []byte(pass)
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, passBytes)
		// Zero out the passphrase immediately after use
		for i := range passBytes {
			passBytes[i] = 0
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse private key with passphrase: %w", err)
		}
	}
	return ssh.PublicKeys(signer), nil
}

func getSSHAgentAuthMethod() (ssh.AuthMethod, error) {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return nil, nil
	}

	conn, err := (&net.Dialer{}).DialContext(context.Background(), "unix", sshAuthSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers), nil
}
