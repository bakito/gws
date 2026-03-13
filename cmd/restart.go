package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
	"github.com/bakito/gws/internal/types"
)

// restartCmd represents the restart command.
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a workstation",
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		return restartWorkstation(cfg)
	},
}

func restartWorkstation(cfg *types.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := gcloud.StopWorkstation(ctx, cfg); err != nil {
		return err
	}
	return gcloud.StartWorkstation(ctx, cfg)
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
