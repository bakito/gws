package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/pkg/gcloud"
)

// stopCmd represents the stop command.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		return gcloud.StopWorkstation(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
