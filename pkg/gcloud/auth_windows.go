//go:build windows
// +build windows

package gcloud

import (
	"os/exec"
)

func openBrowser(authURL string) {
	_ = exec.Command("cmd", "/C", "start", "", authURL).Start()
}
