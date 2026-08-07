package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
)

// restartCmd represents the restart command.
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		if err := gcloud.StopWorkstation(cmd.Context(), cfg); err != nil {
			return err
		}
		return gcloud.StartWorkstation(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
