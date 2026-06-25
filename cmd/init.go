package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
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
