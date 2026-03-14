package script

import (
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/bakito/gws/internal/types"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	orange = "\033[33m"
	reset  = "\033[0m"
)

func TestWindowsReconnectSSH(t *testing.T) {
	tests := []struct {
		name string
		cfg  *types.Config
		want string
	}{
		{
			name: "Basic Configuration",
			cfg: &types.Config{
				CurrentContextName: "test-context",
				Contexts: map[string]*types.Context{
					"test-context": {
						Host: "example.com",
						Port: 22,
						User: "user",
					},
				},
			},
			want: `REM gws-ssh-with-reconnect-test-context.cmd
@echo off
setlocal

REM This script automatically reconnects to an SSH server when the connection drops.
REM It continuously attempts to establish an SSH connection and retries after a specified delay if disconnected.

REM The number of seconds to wait before retrying the connection.
set "RETRY_SECONDS=3"

:reconnect_loop
ssh user@example.com -p 22
if %errorlevel% == 0 goto end
cls
echo Disconnected. Reconnecting in %RETRY_SECONDS% seconds...
timeout /t %RETRY_SECONDS% /nobreak
cls
goto reconnect_loop

:end`,
		},
		{
			name: "With KnownHostsFile",
			cfg: &types.Config{
				CurrentContextName: "test-context-known-hosts",
				Contexts: map[string]*types.Context{
					"test-context-known-hosts": {
						Host:           "example.com",
						Port:           22,
						User:           "user",
						KnownHostsFile: "known_hosts_path",
					},
				},
			},
			want: `REM gws-ssh-with-reconnect-test-context-known-hosts.cmd
@echo off
setlocal

REM This script automatically reconnects to an SSH server when the connection drops.
REM It continuously attempts to establish an SSH connection and retries after a specified delay if disconnected.

REM The number of seconds to wait before retrying the connection.
set "RETRY_SECONDS=3"

:reconnect_loop
ssh user@example.com -p 22 -o UserKnownHostsFile=known_hosts_path
if %errorlevel% == 0 goto end
cls
echo Disconnected. Reconnecting in %RETRY_SECONDS% seconds...
timeout /t %RETRY_SECONDS% /nobreak
cls
goto reconnect_loop

:end`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.cfg.SwitchContext(tt.cfg.CurrentContextName, false)

			got, err := WindowsReconnectSSH(tt.cfg)
			if err != nil {
				t.Fatalf("WindowsReconnectSSH() error = %v", err)
			}

			normalizedGot := strings.ReplaceAll(got, "\r\n", "\n")
			normalizedWant := strings.ReplaceAll(tt.want, "\r\n", "\n")

			if normalizedGot != normalizedWant {
				diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(normalizedWant),
					B:        difflib.SplitLines(normalizedGot),
					FromFile: "want",
					ToFile:   "got",
					Context:  3,
				})
				t.Errorf("WindowsReconnectSSH() mismatch:\n%s", Colorize(diff))
			}
		})
	}
}

// Colorize returns a string with colorized diff.
func Colorize(diff string) string {
	var colored strings.Builder
	for line := range strings.SplitSeq(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "-"):
			colored.WriteString(red + line + reset + "\n")
		case strings.HasPrefix(line, "+"):
			colored.WriteString(green + line + reset + "\n")
		default:
			colored.WriteString(line + "\n")
		}
	}
	return colored.String()
}
