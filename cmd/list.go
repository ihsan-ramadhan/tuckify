package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var (
	colOnline  = color.New(color.FgGreen, color.Bold)
	colOffline = color.New(color.FgRed)
	colHint    = color.New(color.FgYellow)
	colHeader  = color.New(color.Bold)
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

		const (
			wName   = 20
			wStatus = 10
			wSaved  = 8
			wCron   = 16
		)

		wFolder := 6 // min "FOLDER"
		for _, s := range schedules {
			fStr := strings.Join(s.GetFolders(), ", ")
			if len(fStr)+2 > wFolder {
				wFolder = len(fStr) + 2
			}
		}

		sep := strings.Repeat("─", wName) + "┼" +
			strings.Repeat("─", wStatus) + "┼" +
			strings.Repeat("─", wSaved) + "┼" +
			strings.Repeat("─", wCron) + "┼" +
			strings.Repeat("─", wFolder)

		_, _ = colHeader.Printf(" %-*s│ %-*s│ %-*s│ %-*s│ %s\n",
			wName-1, "NAME",
			wStatus-1, "STATUS",
			wSaved-1, "SERVICE",
			wCron-1, "CRON",
			"FOLDER")
		fmt.Println(sep)

		var unsaved []string
		for _, s := range schedules {
			online, _ := srv.Exists(s.Name)

			var statusText, savedText string
			col := colOnline
			if !online {
				statusText, savedText = "offline", "no"
				col = colOffline
				unsaved = append(unsaved, s.Name)
			} else {
				statusText, savedText = "online", "yes"
			}

			fmt.Printf(" %-*s│ %s%s│ %s%s│ %-*s│ %s\n",
				wName-1, s.Name,
				col.Sprint(statusText), strings.Repeat(" ", wStatus-1-len(statusText)),
				col.Sprint(savedText), strings.Repeat(" ", wSaved-1-len(savedText)),
				wCron-1, s.Cron,
				strings.Join(s.GetFolders(), ", "))
		}

		if len(unsaved) > 0 {
			fmt.Println()
			for _, name := range unsaved {
				_, _ = colHint.Printf("  ! %q not active — run 'tuckify start %s'\n", name, name)
			}
			_, _ = colHint.Println("  To activate all at once: tuckify startup")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
