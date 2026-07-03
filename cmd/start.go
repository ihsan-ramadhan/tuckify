package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Activate a saved schedule as a system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		name := args[0]

		found, err := store.Find(name)
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if found == nil {
			return fmt.Errorf("schedule %q not found, run 'tuckify schedule %s <folder> --cron ...' first", name, name)
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		if err := srv.Install(found.Name, found.GetFolders(), found.Cron, found.Config); err != nil {
			return fmt.Errorf("start service: %w", err)
		}

		fmt.Printf("started %q\n", name)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Deactivate a schedule's system service (keeps in list)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		name := args[0]

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		exists, err := srv.Exists(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("schedule %q is already offline", name)
		}

		if err := srv.Uninstall(name); err != nil {
			return fmt.Errorf("stop service: %w", err)
		}

		fmt.Printf("stopped %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}
