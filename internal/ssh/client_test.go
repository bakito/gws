package ssh

import (
	"encoding/base64"
	"fmt"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

type mockPublicKey struct {
	ssh.PublicKey
	t string
	m []byte
}

func (m *mockPublicKey) Type() string    { return m.t }
func (m *mockPublicKey) Marshal() []byte { return m.m }

func Test_formatHostKey(t *testing.T) {
	key := &mockPublicKey{t: "ssh-rsa", m: []byte("keydata")}
	got := formatHostKey("host", 22, key)
	want := "[host]:22 ssh-rsa " + base64.StdEncoding.EncodeToString([]byte("keydata"))
	if got != want {
		t.Errorf("formatHostKey() = %v, want %v", got, want)
	}
}

func Test_HostKeyCallback_Localhost(t *testing.T) {
	key := &mockPublicKey{t: "ssh-rsa", m: []byte("keydata")}
	keyStr := base64.StdEncoding.EncodeToString([]byte("keydata"))

	tests := []struct {
		name     string
		hostname string
		remote   net.Addr
		want     string
	}{
		{
			name:     "Not localhost",
			hostname: "example.com:22",
			remote:   &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 22},
			want:     "[1.2.3.4]:22 ssh-rsa " + keyStr,
		},
		{
			name:     "Localhost",
			hostname: "localhost:22",
			remote:   &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22},
			want:     fmt.Sprintf("[127.0.0.1]:22 ssh-rsa %s\n[localhost]:22 ssh-rsa %s", keyStr, keyStr),
		},
		{
			name:     "Localhost already IP",
			hostname: "127.0.0.1:22",
			remote:   &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22},
			want:     "[127.0.0.1]:22 ssh-rsa " + keyStr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var knownHostsEntry string
			callback := func(hostname string, remote net.Addr, key ssh.PublicKey) {
				if tcpAddr, ok := remote.(*net.TCPAddr); ok {
					knownHostsEntry = formatHostKey(tcpAddr.IP.String(), tcpAddr.Port, key)
					host, _, err := net.SplitHostPort(hostname)
					if err != nil {
						host = hostname
					}
					if host == "localhost" {
						knownHostsEntry += "\n" + formatHostKey(host, tcpAddr.Port, key)
					}
				}
			}

			callback(tt.hostname, tt.remote, key)
			if knownHostsEntry != tt.want {
				t.Errorf("HostKeyCallback got = %q, want %q", knownHostsEntry, tt.want)
			}
		})
	}
}
