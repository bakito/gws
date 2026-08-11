//go:build darwin

package gcloud

import (
	"context"
	"os/exec"
)

func openBrowser(ctx context.Context, _ *types.Config, authURL string) {
	_ = exec.CommandContext(ctx, "open", authURL).Start()
}
