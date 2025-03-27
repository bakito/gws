package gcloud

import (
	"os/exec"
)

func windowsCmd(authURL string) *exec.Cmd {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", authURL)
}
