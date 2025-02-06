package cmd

import (
	"github.com/spf13/cobra"
)

// ctxCmd represents the ctx command
var ctxCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Set the active context",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			return cfg.SwitchContext(args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ctxCmd)
}
