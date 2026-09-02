package gcloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bakito/gws/internal/types"
)

func Test_updateKnownHostsWithGetHostKey(t *testing.T) {
	tempDir := t.TempDir()
	knownHostsFile := filepath.Join(tempDir, "known_hosts")

	type args struct {
		sshContext *types.Context
		port       int
		timeout    time.Duration
		hostKey    []string
	}
	tests := []struct {
		name           string
		args           args
		initialContent string
		wantContent    string
	}{
		{
			name: "New host added to empty file",
			args: args{
				sshContext: &types.Context{Host: "host1", KnownHostsFile: knownHostsFile},
				port:       2222,
				timeout:    time.Second,
				hostKey:    []string{"[host1]:2222 ssh-rsa AAAAB3Nza...key1"},
			},
			initialContent: "",
			wantContent:    "[host1]:2222 ssh-rsa AAAAB3Nza...key1",
		},
		{
			name: "New host added to existing file",
			args: args{
				sshContext: &types.Context{Host: "host2", KnownHostsFile: knownHostsFile},
				port:       2223,
				timeout:    time.Second,
				hostKey:    []string{"[host2]:2223 ssh-rsa AAAAB3Nza...key2"},
			},
			initialContent: "[host1]:2222 ssh-rsa AAAAB3Nza...key1",
			wantContent:    "[host1]:2222 ssh-rsa AAAAB3Nza...key1\n[host2]:2223 ssh-rsa AAAAB3Nza...key2",
		},
		{
			name: "Existing host key updated",
			args: args{
				sshContext: &types.Context{Host: "host1", KnownHostsFile: knownHostsFile},
				port:       2222,
				timeout:    time.Second,
				hostKey:    []string{"[host1]:2222 ssh-rsa AAAAB3Nza...newkey1"},
			},
			initialContent: "[host1]:2222 ssh-rsa AAAAB3Nza...key1\n[host2]:2223 ssh-rsa AAAAB3Nza...key2",
			wantContent:    "[host1]:2222 ssh-rsa AAAAB3Nza...newkey1\n[host2]:2223 ssh-rsa AAAAB3Nza...key2",
		},
		{
			name: "Existing host key same, no change",
			args: args{
				sshContext: &types.Context{Host: "host1", KnownHostsFile: knownHostsFile},
				port:       2222,
				timeout:    time.Second,
				hostKey:    []string{"[host1]:2222 ssh-rsa AAAAB3Nza...key1"},
			},
			initialContent: "[host1]:2222 ssh-rsa AAAAB3Nza...key1\n[host2]:2223 ssh-rsa AAAAB3Nza...key2",
			wantContent:    "[host1]:2222 ssh-rsa AAAAB3Nza...key1\n[host2]:2223 ssh-rsa AAAAB3Nza...key2",
		},
		{
			name: "Multiple host keys (localhost case)",
			args: args{
				sshContext: &types.Context{Host: "localhost", KnownHostsFile: knownHostsFile},
				port:       2222,
				timeout:    time.Second,
				hostKey:    []string{"[127.0.0.1]:2222 ssh-rsa AAAAB3Nza...key1", "[localhost]:2222 ssh-rsa AAAAB3Nza...key1"},
			},
			initialContent: "",
			wantContent:    "[127.0.0.1]:2222 ssh-rsa AAAAB3Nza...key1\n[localhost]:2222 ssh-rsa AAAAB3Nza...key1",
		},
		{
			name: "Update one of multiple host keys",
			args: args{
				sshContext: &types.Context{Host: "localhost", KnownHostsFile: knownHostsFile},
				port:       2222,
				timeout:    time.Second,
				hostKey: []string{
					"[127.0.0.1]:2222 ssh-rsa AAAAB3Nza...newkey1",
					"[localhost]:2222 ssh-rsa AAAAB3Nza...newkey1",
				},
			},
			initialContent: "[127.0.0.1]:2222 ssh-rsa AAAAB3Nza...key1\n[localhost]:2222 ssh-rsa AAAAB3Nza...key1",
			wantContent:    "[127.0.0.1]:2222 ssh-rsa AAAAB3Nza...newkey1\n[localhost]:2222 ssh-rsa AAAAB3Nza...newkey1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.WriteFile(knownHostsFile, []byte(tt.initialContent), 0o644)
			updateKnownHosts(
				tt.args.sshContext,
				tt.args.hostKey,
			)

			content, _ := os.ReadFile(knownHostsFile)
			got := strings.TrimSpace(string(content))
			if got != tt.wantContent {
				t.Errorf("updateKnownHostsWithGetHostKey() gotContent = %v, want %v", got, tt.wantContent)
			}
		})
	}
}
