//go:build windows

package gcloud

import (
	"context"

	"github.com/bakito/gws/internal/types"
)

func openBrowser(ctx context.Context, cfg *types.Config, authURL string) {
	_ = windowsCmd(ctx, cfg, authURL).Start()
}
