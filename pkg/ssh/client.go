package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bakito/gws/pkg/env"
	"github.com/bakito/gws/pkg/passwd"
	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

func Client(addr string, user string, privateKeyFile string) (*client, error) {
	privateKey, err := os.ReadFile(env.ExpandEnv(privateKeyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the private key
	signer, err := loadPrivateKey(privateKey, privateKeyFile)
	if err != nil {
		return nil, err
	}

	// Define SSH connection details
	clientConfig := &ssh.ClientConfig{
		User: user, // Remote SSH username
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer), // Use the private key for authentication
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // #nosec G106: Insecure, as we always get a new cert with gcloud
	}

	// Connect to the SSH server
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
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
		sshClient: sshClient,
		scpClient: scpClient,
	}, nil
}

func loadPrivateKey(privateKey []byte, privateKeyFile string) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		// Check if the error is due to a missing passphrase
		if errors.Is(err, &ssh.PassphraseMissingError{}) {
			pass, err := passwd.Prompt(fmt.Sprintf("Please enter the passphrase for private key (%s):", privateKeyFile))
			if err != nil {
				return nil, err
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(pass))
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}
	return signer, nil
}

type client struct {
	sshClient *ssh.Client
	scpClient scp.Client
}

func (c *client) Close() {
	if c.sshClient != nil {
		_ = c.sshClient.Close()
	}

	c.scpClient.Close()
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

func (c *client) CopyFile(from string, to string, permissions string) error {
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
