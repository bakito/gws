//go:build darwin

package gcloud

import (
	"context"
	"os/exec"
)

func openBrowser(authURL string) {
	_ = exec.CommandContext(context.Background(), "open", authURL).Start()
}
