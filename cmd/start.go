package cmd

import (
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, ssh, err := readConfig()
		if err != nil {
			return err
		}
		gcloud.StartWorkstation(*ssh)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
