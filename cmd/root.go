package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ihsan-ramadhan/tuck/internal/config"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "tuck",
	Short: "Automatic file organizer",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", config.DefaultConfigPath(), "path to rules.toml")
}
