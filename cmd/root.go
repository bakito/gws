package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bakito/gws/pkg/types"
	"github.com/bakito/gws/version"
)

// rootCmd represents the base command when called without any subcommands.
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
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", types.ConfigFileName, "The config file to be used")
}

func readConfig() (*types.Config, error) {
	var err error
	config := &types.Config{}

	if err := config.Load(flagConfig); err != nil {
		return nil, err
	}

	if flagContext != "" {
		err = config.SwitchContext(flagContext)
	}

	return config, err
}
