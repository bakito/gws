package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/script"
)

// scriptsCmd represents the generate scripts command.
var (
	scriptsCmd = &cobra.Command{
		Use:   "scripts",
		Short: "Generate complementary scripts",
	}
	flagOutput string
)

func init() {
	rootCmd.AddCommand(scriptsCmd)
	scriptsCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output file (default stdout)")
	scriptsCmd.AddCommand(scriptsReconnectCmd)
}

// scriptsReconnectCmd represents the start command.
var scriptsReconnectCmd = &cobra.Command{
	Use:   "win-reconnect-ssh",
	Short: "Generate Windows ssh reconnect script",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagOutput == "" {
			log.SetLogger(log.Null)
		}
		cfg, err := readConfig(args...)
		if err != nil {
			return err
		}

		s, err := script.WindowsReconnectSSH(cfg)
		if err != nil {
			return err
		}

		if flagOutput != "" {
			cmd.Printf("💾 Writing SSH reconnect script for context %q on Windows to %q\n", cfg.CurrentContextName, flagOutput)
			return os.WriteFile(flagOutput, []byte(s), 0o644)
		}
		cmd.Println(s)

		return nil
	},
}
