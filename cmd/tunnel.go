package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/pkg/gcloud"
)

var flagLocalPort *int

// tunnelCmd represents the tunnel command.
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "tunnel a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}
		return gcloud.TCPTunnel(cfg, flagLocalPort)
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)
	tunnelCmd.PersistentFlags().
		IntVarP(flagLocalPort, "local-host-port", "p", 0, "The local host port to open (default ist the port from the config)")
}
