package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ihsan-ramadhan/tuckify/internal/config"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:     "tuckify",
	Short:   "Automatic file organizer",
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", config.DefaultConfigPath(), "path to rules.toml")
}
