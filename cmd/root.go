package cmd

import (
	"fmt"
	"os"

	"github.com/ihsan-ramadhan/tuckify/internal/ansi"
	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/spf13/cobra"
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

func warnNoRules(cfg *config.Config, path string) {
		if len(cfg.Rules) == 0 {
			ansi.PrintYellow("warning: no rules defined in %s — nothing to organize\n", path)
			fmt.Printf("hint: create a config at %s (see rules.example.toml)\n", config.DefaultConfigPath())
		}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", config.DefaultConfigPath(), "path to rules.toml")
}
