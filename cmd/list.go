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

		fmt.Printf("%-20s %-10s %-8s %-15s %s\n", "NAME", "STATUS", "SAVED", "CRON", "FOLDER")

		var unsaved []string
		for _, s := range schedules {
			online, _ := srv.Exists(s.Name)

			status := "offline"
			saved := "no"

			if online {
				status = "online"
				saved = "yes"
			}

			fmt.Printf("%-20s %-10s %-8s %-15s %s\n", s.Name, status, saved, s.Cron, s.Folder)

			if !online {
				unsaved = append(unsaved, s.Name)
			}
		}

		if len(unsaved) > 0 {
			fmt.Println()
			for _, name := range unsaved {
				fmt.Printf("  ! %q not active — run 'tuckify start %s'\n", name, name)
			}
			fmt.Println("  To activate all at once: tuckify startup")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
