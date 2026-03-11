package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
	"github.com/bakito/gws/internal/types"
)

// restartCmd represents the start command.
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a workstation",
	RunE: func(_ *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
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
