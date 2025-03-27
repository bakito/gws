package cmd

import (
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/spf13/cobra"
)

// restartCmd represents the start command.
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}
		if err := gcloud.StopWorkstation(cfg); err != nil {
			return err
		}
		return gcloud.StartWorkstation(cfg)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
