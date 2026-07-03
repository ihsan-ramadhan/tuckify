package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ihsan-ramadhan/tuckify/internal/history"
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

var (
	listJSON   bool
	listStatus bool
)

// scheduleView is the machine-readable representation of a schedule row,
// used by `tuckify list --json` and shared with the human-readable renderer.
type scheduleView struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`  // "online" or "offline"
	Service   bool      `json:"service"` // true if installed as a system service
	Cron      string    `json:"cron"`
	Folders   []string  `json:"folders"`
	Config    string    `json:"config,omitempty"`
	LastRun   time.Time `json:"last_run,omitempty"`
	LastFiles int       `json:"last_files,omitempty"`
}

func findLastRunForSchedule(folders []string, runs []history.Run) (time.Time, int) {
	// Look through runs from newest to oldest
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		// Check if folders overlap
		for _, sf := range folders {
			for _, rf := range r.Folders {
				if sf == rf {
					moves := 0
					for _, e := range r.Entries {
						if e.Action == "move" || e.Action == "" {
							moves++
						}
					}
					return r.Timestamp, moves
				}
			}
		}
	}
	return time.Time{}, 0
}

func buildScheduleViews(schedules []store.Schedule, srv service.Service, runs []history.Run) []scheduleView {
	views := make([]scheduleView, 0, len(schedules))
	for _, s := range schedules {
		online, _ := srv.Exists(s.Name)
		status := "offline"
		if online {
			status = "online"
		}
		folders := s.GetFolders()
		lastRun, lastFiles := findLastRunForSchedule(folders, runs)

		views = append(views, scheduleView{
			Name:      s.Name,
			Status:    status,
			Service:   online,
			Cron:      s.Cron,
			Folders:   folders,
			Config:    s.Config,
			LastRun:   lastRun,
			LastFiles: lastFiles,
		})
	}
	return views
}

func printSchedulesJSON(views []scheduleView) error {
	if views == nil {
		views = []scheduleView{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(views)
}

func printSchedulesTable(views []scheduleView, showStatus bool) {
	const (
		wName   = 20
		wStatus = 10
		wSaved  = 8
		wCron   = 16
		wLast   = 20
		wFiles  = 8
	)

	wFolder := 6 // min "FOLDER"
	for _, v := range views {
		fStr := strings.Join(v.Folders, ", ")
		if len(fStr)+2 > wFolder {
			wFolder = len(fStr) + 2
		}
	}

	var sep string
	if showStatus {
		sep = strings.Repeat("─", wName) + "┼" +
			strings.Repeat("─", wStatus) + "┼" +
			strings.Repeat("─", wSaved) + "┼" +
			strings.Repeat("─", wCron) + "┼" +
			strings.Repeat("─", wLast) + "┼" +
			strings.Repeat("─", wFiles) + "┼" +
			strings.Repeat("─", wFolder)

		_, _ = colHeader.Printf(" %-*s│ %-*s│ %-*s│ %-*s│ %-*s│ %-*s│ %s\n",
			wName-1, "NAME",
			wStatus-1, "STATUS",
			wSaved-1, "SERVICE",
			wCron-1, "CRON",
			wLast-1, "LAST RUN",
			wFiles-1, "FILES",
			"FOLDER")
	} else {
		sep = strings.Repeat("─", wName) + "┼" +
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
	}
	fmt.Println(sep)

	var unsaved []string
	for _, v := range views {
		statusText := v.Status
		savedText := "no"
		col := colOffline
		if v.Service {
			savedText = "yes"
			col = colOnline
		} else {
			unsaved = append(unsaved, v.Name)
		}

		if showStatus {
			lastRunText := "never"
			filesText := "-"
			if !v.LastRun.IsZero() {
				lastRunText = v.LastRun.Format("02 Jan 06 15:04")
				filesText = fmt.Sprintf("%d", v.LastFiles)
			}

			fmt.Printf(" %-*s│ %s%s│ %s%s│ %-*s│ %-*s│ %-*s│ %s\n",
				wName-1, v.Name,
				col.Sprint(statusText), strings.Repeat(" ", wStatus-1-len(statusText)),
				col.Sprint(savedText), strings.Repeat(" ", wSaved-1-len(savedText)),
				wCron-1, v.Cron,
				wLast-1, lastRunText,
				wFiles-1, filesText,
				strings.Join(v.Folders, ", "))
		} else {
			fmt.Printf(" %-*s│ %s%s│ %s%s│ %-*s│ %s\n",
				wName-1, v.Name,
				col.Sprint(statusText), strings.Repeat(" ", wStatus-1-len(statusText)),
				col.Sprint(savedText), strings.Repeat(" ", wSaved-1-len(savedText)),
				wCron-1, v.Cron,
				strings.Join(v.Folders, ", "))
		}
	}

	if len(unsaved) > 0 {
		fmt.Println()
		for _, name := range unsaved {
			_, _ = colHint.Printf("  ! %q not active — run 'tuckify start %s'\n", name, name)
		}
		_, _ = colHint.Println("  To activate all at once: tuckify startup")
	}
}

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
			if listJSON {
				return printSchedulesJSON(nil)
			}
			fmt.Println("No saved schedules.")
			return nil
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		runs, _ := history.LoadAll()
		views := buildScheduleViews(schedules, srv, runs)

		if listJSON {
			return printSchedulesJSON(views)
		}

		printSchedulesTable(views, listStatus)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON for scripting/integrations")
	listCmd.Flags().BoolVar(&listStatus, "status", false, "show last run status and processed files count")
	rootCmd.AddCommand(listCmd)
}
