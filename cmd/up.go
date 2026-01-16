package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/ssh"
)

// upCmd represents the up command.
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Upload files and dirs",
	RunE: func(*cobra.Command, []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		log.Logf("Running context %s\n", cfg.CurrentContextName)

		sshCtx := cfg.CurrentContext()
		cl, err := ssh.NewClient(sshCtx.HostAddr(), sshCtx.User, sshCtx.PrivateKeyFile, cfg.SSHTimeout())
		if err != nil {
			return err
		}

		defer cl.Close()

		if len(sshCtx.Dirs) > 0 {
			log.Log("Creating directories")
			for _, dir := range sshCtx.Dirs {
				if dir.Permissions != "" {
					log.Logf("Creating directory %q with permissions %s\n", dir.Path, dir.Permissions)
					_, err = cl.Execute(fmt.Sprintf("mkdir -p %s; chmod %s /home/user/.ssh", dir.Path, dir.Permissions))
				} else {
					log.Logf("Creating directory %q\n", dir.Path)
					_, err = cl.Execute("mkdir -p " + dir.Path)
				}
				if err != nil {
					return err
				}
			}
		}

		if len(sshCtx.Files) > 0 {
			log.Log("Uploading files")
			for _, file := range sshCtx.Files {
				if file.Permissions == "0400" {
					log.Logf(
						"Add writable file permission for upload %q with permissions %s\n",
						file.Path,
						file.Permissions,
					)
					_, err := cl.Execute(fmt.Sprintf("if [ -f %s ]; then chmod u+w %s; fi", file.Path, file.Path))
					if err != nil {
						return err
					}
				}
				log.Logf(
					"Uploading file for %q to %q with permissions %s\n",
					file.SourcePath,
					file.Path,
					file.Permissions,
				)
				err = cl.CopyFile(file.SourcePath, file.Path, file.Permissions)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
