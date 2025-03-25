//go:build windows
// +build windows

package gcloud

import "os/exec"

func openBrowser(authURL string) {
	_ = exec.Command("cmd", "/c", "start", authURL).Start()
}
