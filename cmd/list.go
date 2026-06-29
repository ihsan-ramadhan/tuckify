package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved schedules",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		schedules, err := store.Load()
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if len(schedules) == 0 {
			fmt.Println("No saved schedules.")
			return nil
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		fmt.Printf("%-20s %-10s %-15s %s\n", "NAME", "STATUS", "CRON", "FOLDER")
		for _, s := range schedules {
			status := "offline"
			if online, _ := srv.Exists(s.Name); online {
				status = "online"
			}
			fmt.Printf("%-20s %-10s %-15s %s\n", s.Name, status, s.Cron, s.Folder)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
