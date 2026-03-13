package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/script"
)

// scriptsCmd represents the generate scripts command.
var scriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Generate complementary scripts",
}

func init() {
	rootCmd.AddCommand(scriptsCmd)
	scriptsCmd.AddCommand(scriptsReconnectCmd)
}

// scriptsReconnectCmd represents the start command.
var scriptsReconnectCmd = &cobra.Command{
	Use:   "win-reconnect-ssh",
	Short: "Generate Windows ssh reconnect script",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagContext == "" && len(args) == 1 {
			flagContext = args[0]
		}

		cfg, err := readConfig()
		if err != nil {
			return err
		}

		s, err := script.WindowsReconnectSSH(cfg)
		if err != nil {
			return err
		}

		cmd.Println(strings.Repeat("-", 80))
		cmd.Println()
		cmd.Printf("SSH reconnect script for ges context '%s' on Windows\n", cfg.CurrentContextName)
		cmd.Println(strings.Repeat("-", 80))
		cmd.Println(s)
		cmd.Println(strings.Repeat("-", 80))

		return nil
	},
}
