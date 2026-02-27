package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/ssh"
)

// downCmd represents the down command.
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Download files and dirs",
	RunE: func(*cobra.Command, []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		log.Logf("Running context %s", cfg.CurrentContextName)

		sshCtx := cfg.CurrentContext()
		cl, err := ssh.NewClient(sshCtx.HostAddr(), sshCtx.User, sshCtx.PrivateKeyFile, cfg.SSHTimeout())
		if err != nil {
			return err
		}

		defer cl.Close()

		if len(sshCtx.Files) > 0 {
			log.Log("Downloading files")
			for _, file := range sshCtx.Files {
				log.Logf(
					"Downloading file for %q to %q",
					file.SourcePath,
					file.Path,
				)
				err = cl.CopyFile(file.Path, file.SourcePath, file.Permissions)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
