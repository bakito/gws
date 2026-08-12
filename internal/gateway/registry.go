//go:build !windows

package gateway

import (
	"context"

	"github.com/bakito/gws/internal/types"
)

func UpdateDownloadLocation(_ context.Context, _ *types.Config) error {
	return nil
}
