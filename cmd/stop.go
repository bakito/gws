package cmd

import (
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, ssh, err := readConfig()
		if err != nil {
			return err
		}
		gcloud.StopWorkstation(*ssh)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
