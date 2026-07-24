package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall tuckify service, binary, and optionally configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := service.NewService()
		if err == nil {
			if err := srv.Uninstall(""); err == nil {
				fmt.Println("Service successfully uninstalled.")
			} else {
				fmt.Fprintf(os.Stderr, "Warning: failed to uninstall service: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize service manager: %v\n", err)
		}

		binaryPath, err := os.Executable()
		if err == nil {
			if err := os.Remove(binaryPath); err == nil {
				fmt.Printf("Binary successfully deleted from %s.\n", binaryPath)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete binary: %v\n", err)
			}

			dir := filepath.Dir(binaryPath)
			for _, partner := range []string{"tuckify", "tuckify-gui"} {
				partnerPath := filepath.Join(dir, partner)
				if partnerPath != binaryPath {
					_ = os.Remove(partnerPath)
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, "Warning: failed to find running binary path.")
		}

		homeDir, err := os.UserHomeDir()
		if err == nil {
			desktopFile := filepath.Join(homeDir, ".local", "share", "applications", "tuckify.desktop")
			_ = os.Remove(desktopFile)
			iconFile := filepath.Join(homeDir, ".local", "share", "icons", "hicolor", "512x512", "apps", "tuckify.png")
			_ = os.Remove(iconFile)
		}

		configDir := filepath.Dir(config.DefaultConfigPath())
		if _, err := os.Stat(configDir); err == nil {
			fd := os.Stdout.Fd()
			deleteConfig := false

			if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
				fmt.Print("Do you also want to delete the configuration directory ~/.tuckify? [y/N]: ")
				var response string
				_, _ = fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "y" || response == "yes" {
					deleteConfig = true
				}
			}

			if deleteConfig {
				if err := os.RemoveAll(configDir); err == nil {
					fmt.Println("Configuration directory successfully deleted.")
				} else {
					fmt.Fprintf(os.Stderr, "Warning: failed to delete configuration directory: %v\n", err)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
