package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/patch"
)

// patchCmd represents the patch command.
var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "patch local files",
	RunE: func(*cobra.Command, []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}

		for id, filePatch := range cfg.FilePatches {
			if err := patch.Patch(id, filePatch); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
}
