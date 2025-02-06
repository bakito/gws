package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bakito/gws/pkg/types"
	"github.com/bakito/gws/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".gws.yaml"

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:     "gws",
		Short:   "Google Cloud Workstation Utils",
		Version: version.Version,
	}
	flagConfig  string
	flagContext string
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagContext, "ctx", "", "The context to be used")
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", configFileName, "The config file to be used")
}

func readConfig() (string, *types.Context, error) {
	var file string
	if flagConfig != "" {
		if _, err := os.Stat(flagConfig); err == nil {
			file = flagConfig
		}
	}

	if file == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", nil, err
		}

		homePath := filepath.Join(userHomeDir, configFileName)
		if _, err := os.Stat(homePath); err == nil {
			file = homePath
		} else {
			return "", nil, errors.New("config file not found")
		}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Error reading config file", "file", file, "error", err)
		return "", nil, err
	}

	config := &types.Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		slog.Error("Error parsing config", "error", err)
		return "", nil, err
	}

	return config.Get(flagContext)
}
