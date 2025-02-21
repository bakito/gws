package cmd

import (
	"fmt"
	"log/slog"

	"github.com/bakito/gws/pkg/client"
	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Upload files and dirs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		slog.Info("Running context", "name", cfg.CurrentContextName)

		ssh := cfg.CurrentContext()
		cl, err := client.New(ssh.HostAddr(), ssh.User, ssh.PrivateKeyFile)
		if err != nil {
			return err
		}

		defer cl.Close()

		if len(ssh.Dirs) > 0 {
			slog.Info("Creating directories")
			for _, dir := range ssh.Dirs {
				if dir.Permissions != "" {
					slog.Info("Creating directory", "path", dir.Path, "permissions", dir.Permissions)
					_, err = cl.Execute(fmt.Sprintf("mkdir -p %s; chmod %s /home/user/.ssh", dir.Path, dir.Permissions))
				} else {
					slog.Info("Creating directory", "path", dir.Path)
					_, err = cl.Execute(fmt.Sprintf("mkdir -p %s", dir.Path))
				}
				if err != nil {
					return err
				}
			}
		}

		if len(ssh.Files) > 0 {
			slog.Info("Uploading files")
			for _, file := range ssh.Files {
				if file.Permissions == "0400" {
					slog.Info("Add writable file permission for upload", "file", file.Path, "permissions", file.Permissions)
					_, err := cl.Execute(fmt.Sprintf("if [ -f %s ]; then chmod u+w %s; fi", file.Path, file.Path))
					if err != nil {
						return err
					}
				}
				slog.Info("Uploading file", "from", file.SourcePath, "to", file.Path, "permissions", file.Permissions)
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
