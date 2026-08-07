package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
)

// stopCmd represents the stop command.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		return gcloud.StopWorkstation(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
