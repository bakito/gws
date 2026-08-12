package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gateway"
)

// gatewayCmd represents the tunnel command.
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "JetBrains Gateway tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}
		return gateway.UpdateDownloadLocation(cmd.Context(), cfg)
	},
}

func init() {
	rootCmd.AddCommand(gatewayCmd)
}
