package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/ssh"
	"github.com/bakito/gws/internal/types"
)

// downCmd represents the down command.
var downCmd = &cobra.Command{
	Use:   types.DirectionDown,
	Short: "Download files and dirs",
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, err := readConfig(args...)
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
				if file.Direction != types.DirectionDown {
					continue
				}

				log.Logf(
					"Downloading file from %q to %q with permissions %s [%s]",
					file.Path,
					file.SourcePath,
					file.Permissions,
					file.Direction,
				)
				err = cl.DownloadFile(file.Path, file.SourcePath, file.Permissions)
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
