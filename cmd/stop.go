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
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}
		gcloud.StopWorkstation(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
