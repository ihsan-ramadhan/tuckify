package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "Restart a schedule's system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		found, err := store.Find(name)
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if found == nil {
			return fmt.Errorf("schedule %q not found", name)
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		exists, err := srv.Exists(name)
		if err != nil {
			return err
		}
		if exists {
			if err := srv.Uninstall(name); err != nil {
				return fmt.Errorf("stop service: %w", err)
			}
		}

		if err := srv.Install(found.Name, found.Folder, found.Cron, found.Config); err != nil {
			return fmt.Errorf("start service: %w", err)
		}

		fmt.Printf("restarted %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
