package cmd

import (
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/spf13/cobra"
)

// restartCmd represents the start command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		gcloud.StopWorkstation(cfg)
		gcloud.StartWorkstation(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
