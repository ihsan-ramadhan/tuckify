package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var initCronExpr string

var initCmd = &cobra.Command{
	Use:   "init <folder>",
	Short: "Register tuckify as a startup service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		folder, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve absolute folder path: %w", err)
		}

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			return fmt.Errorf("folder not found: %s", folder)
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		exists, err := srv.Exists()
		if err != nil {
			return fmt.Errorf("check service existence: %w", err)
		}

		if exists {
			fd := os.Stdout.Fd()
			if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
				fmt.Print("Service 'tuckify' already exists. Overwrite? [y/N]: ")
				var response string
				_, _ = fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Installation aborted.")
					return nil
				}
			}
		}

		var customConfigPath string
		if cmd.Flags().Changed("config") {
			customConfigPath, err = filepath.Abs(configPath)
			if err != nil {
				return fmt.Errorf("resolve absolute config path: %w", err)
			}
		}

		if err := srv.Install(folder, initCronExpr, customConfigPath); err != nil {
			return fmt.Errorf("install service: %w", err)
		}

		fmt.Println("Service 'tuckify' successfully installed.")
		statusMsg, err := srv.CheckStatus()
		if err == nil && statusMsg != "" {
			fmt.Println(statusMsg)
		}

		return nil
	},
}

var uninitCmd = &cobra.Command{
	Use:   "uninit",
	Short: "Remove tuckify from startup service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := service.NewService()
		if err != nil {
			return err
		}

		if err := srv.Uninstall(); err != nil {
			return fmt.Errorf("uninstall service: %w", err)
		}

		fmt.Println("Service 'tuckify' successfully removed from startup.")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initCronExpr, "cron", "", `cron expression, e.g. "0 9 * * *"`)
	_ = initCmd.MarkFlagRequired("cron")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(uninitCmd)
}
