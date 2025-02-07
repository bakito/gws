package cmd

import (
	"github.com/bakito/gws/pkg/patch"
	"github.com/spf13/cobra"
)

// patchCmd represents the patch command
var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "patch local files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}

		for _, filePatch := range cfg.FilePatches {
			if err := patch.Patch(filePatch); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
}
