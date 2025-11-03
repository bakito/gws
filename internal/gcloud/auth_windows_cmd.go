package gcloud

import (
	"context"
	"os/exec"
)

func windowsCmd(authURL string) *exec.Cmd {
	return exec.CommandContext(context.Background(), "rundll32.exe", "url.dll,FileProtocolHandler", authURL)
}
