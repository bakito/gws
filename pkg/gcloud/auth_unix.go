//go:build aix || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris

package gcloud

import (
	"os"
	"os/exec"
	"strings"
)

func openBrowser(authURL string) {
	var cmd *exec.Cmd
	if IsWSL() {
		cmd = windowsCmd(authURL)
	} else {
		cmd = exec.Command("xdg-open", authURL)
	}
	_ = cmd.Start()
}

func IsWSL() bool {
	b, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(b)), "microsoft") || strings.Contains(strings.ToLower(string(b)), "wsl")
}
