package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "Install all saved schedules as system service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		schedules, err := store.Load()
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if len(schedules) == 0 {
			fmt.Println("No saved schedules. Run 'tuckify schedule <name> <folder> --cron ...' first.")
			return nil
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		for _, s := range schedules {
			if err := srv.Install(s.Name, s.GetFolders(), s.Cron, s.Config); err != nil {
				fmt.Printf("failed to install %q: %v\n", s.Name, err)
				continue
			}
			fmt.Printf("installed %q (%s)\n", s.Name, s.Cron)
		}

		statusMsg, err := srv.CheckStatus()
		if err == nil && statusMsg != "" {
			fmt.Println(statusMsg)
		}
		return nil
	},
}

var unstartupCmd = &cobra.Command{
	Use:   "unstartup",
	Short: "Remove all tuckify system services",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := service.NewService()
		if err != nil {
			return err
		}
		if err := srv.Uninstall(""); err != nil {
			return fmt.Errorf("uninstall services: %w", err)
		}
		fmt.Println("All tuckify services removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startupCmd)
	rootCmd.AddCommand(unstartupCmd)
}
