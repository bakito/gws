package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		return gcloud.StartWorkstation(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
