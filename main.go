package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/bakito/gws/pkg/client"
	"github.com/bakito/gws/pkg/gcloud"
	"github.com/bakito/gws/pkg/types"
	"gopkg.in/yaml.v3"
)

func main() {
	cfg := flag.String("config", ".gws.yaml", "Path to config file")
	ctxName := flag.String("context", "", "The context to use")
	startOnly := flag.Bool("start", false, "Start a workstation")
	stopOnly := flag.Bool("stop", false, "Stop a workstation")
	flag.Parse()

	data, err := os.ReadFile(*cfg)
	if err != nil {
		slog.Error("Error reading config file", "file", *cfg, "error", err)
		os.Exit(1)
	}

	config := &types.Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		slog.Error("Error parsing config", "error", err)
		os.Exit(1)
	}

	selectedContext, sshContext, err := config.Get(*ctxName)
	if err != nil {
		slog.Error("Error getting context", "error", err)
		os.Exit(1)
	}
	if *stopOnly {
		gcloud.StopWorkstation(sshContext)
		return
	}
	if *startOnly {
		gcloud.StartWorkstation(sshContext)
		return
	}

	slog.Info("Running context", "name", selectedContext)

	cl, err := client.New(sshContext.HostAddr(), sshContext.User, sshContext.PrivateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	defer cl.Close()

	if len(sshContext.Dirs) > 0 {
		slog.Info("Creating directories")
		for _, dir := range sshContext.Dirs {
			if dir.Permissions != "" {
				slog.Info("Creating directory", "path", dir.Path, "permissions", dir.Permissions)
				_, err = cl.Execute(fmt.Sprintf("mkdir -p %s; chmod %s /home/user/.ssh", dir.Path, dir.Permissions))
			} else {
				slog.Info("Creating directory", "path", dir.Path)
				_, err = cl.Execute(fmt.Sprintf("mkdir -p %s", dir.Path))
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if len(sshContext.Files) > 0 {
		slog.Info("Uploading files")
		for _, file := range sshContext.Files {
			slog.Info("Uploading file", "from", file.SourcePath, "to", file.Path, "permissions", file.Permissions)
			if file.Permissions == "0400" {
				_, err = cl.Execute(fmt.Sprintf("if [ -f %s ]; then chmod u+w %s; fi", file.Path, file.Path))
				if err != nil {
					log.Fatal(err)
				}
				err = cl.CopyFile(file.SourcePath, file.Path, file.Permissions)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
