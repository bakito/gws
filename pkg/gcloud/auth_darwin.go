//go:build darwin

package gcloud

import "os/exec"

func openBrowser(authURL string) {
	_ = exec.Command("open", authURL).Start()
}
