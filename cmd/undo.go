package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ihsan-ramadhan/tuckify/internal/history"
	"github.com/spf13/cobra"
)

var (
	undoHistory bool
	undoID      int
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last tuckify run (or a specific run by ID)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if undoHistory {
			return printUndoHistory()
		}

		n, err := history.Undo(undoID)
		if err != nil {
			return err
		}
		if n == 0 {
			if undoID > 0 {
				fmt.Fprintf(os.Stderr, "run %d has no undoable file operations\n", undoID)
			} else {
				fmt.Fprintln(os.Stderr, "nothing to undo")
			}
			return nil
		}
		fmt.Printf("reverted %d file(s)\n", n)
		return nil
	},
}

func printUndoHistory() error {
	runs, err := history.LoadAll()
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	if len(runs) == 0 {
		fmt.Println("No undo history.")
		return nil
	}

	// Display newest first
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		ts := r.Timestamp.Format(time.RFC822)
		folders := "no folders"
		if len(r.Folders) > 0 {
			folders = fmt.Sprintf("%d folder(s)", len(r.Folders))
		}
		moves := 0
		for _, e := range r.Entries {
			if e.Action == "move" || e.Action == "" {
				moves++
			}
		}
		fmt.Printf("  %d  %s  %s, %d file(s) moved\n", r.ID, ts, folders, moves)
	}
	fmt.Println("\nUsage: tuckify undo --id <n>  (or just 'tuckify undo' for latest)")
	return nil
}

func init() {
	undoCmd.Flags().BoolVar(&undoHistory, "history", false, "view past runs that can be undone")
	undoCmd.Flags().IntVar(&undoID, "id", 0, "undo a specific run by ID (0 = latest)")
	rootCmd.AddCommand(undoCmd)
}
