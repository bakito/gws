package cmd

import (
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}
		gcloud.DeleteWorkstation(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
