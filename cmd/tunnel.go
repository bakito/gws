package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/tunnel"
)

var (
	flagLocalPort  int
	flagTokenCheck bool
	flagRestart    bool
)

// tunnelCmd represents the tunnel command.
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "tunnel a workstation",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		if flagRestart {
			if err := restartWorkstation(cfg); err != nil {
				return err
			}
		} else {
			if err := startWorkstation(cfg); err != nil {
				return err
			}
		}

		m := tunnel.NewModel(cmd.Context(), cfg, flagLocalPort)
		p := tea.NewProgram(m)
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)
	tunnelCmd.PersistentFlags().
		IntVarP(&flagLocalPort, "local-host-port", "p", 0, "The local host port to open (default ist the port from the config)")
	tunnelCmd.PersistentFlags().
		BoolVar(&flagTokenCheck, "check-token", true, "Enable periodic token check")
	tunnelCmd.PersistentFlags().
		BoolVar(&flagRestart, "restart", false, "Restart the workstation before opening the tunnel")
}
