//go:build darwin

package gcloud

import (
	"context"
	"os/exec"

	"github.com/bakito/gws/internal/types"
)

func openBrowser(ctx context.Context, _ *types.Config, authURL string) {
	_ = exec.CommandContext(ctx, "open", authURL).Start()
}
