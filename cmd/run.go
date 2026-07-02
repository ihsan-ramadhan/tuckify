package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/history"
	"github.com/ihsan-ramadhan/tuckify/internal/organizer"
	"github.com/spf13/cobra"
)

var (
	dryRun    bool
	recursive bool
	yesFlag   bool
)

var runCmd = &cobra.Command{
	Use:   "run <folder>",
	Short: "Organize files in a folder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		folder := args[0]

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			return fmt.Errorf("folder not found: %s", folder)
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		warnNoRules(cfg, configPath)

		if !dryRun && !yesFlag {
			previewResults, err := organizer.Organize(folder, cfg, true, recursive)
			if err != nil {
				return err
			}
			deletions := 0
			for _, r := range previewResults {
				if !r.Skipped && r.Action == "delete" {
					deletions++
				}
			}
			if deletions > 0 {
				color.Red("warning: this operation will delete %d file(s).", deletions)
				fmt.Print("confirm deletion? [y/N]: ")
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					return fmt.Errorf("operation cancelled")
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					return fmt.Errorf("operation cancelled")
				}
			}
		}

		results, err := organizer.Organize(folder, cfg, dryRun, recursive)
		if err != nil {
			return err
		}

		// save history for undo (only on real run)
		if !dryRun {
			var histEntries []history.Entry
			for _, r := range results {
				if !r.Skipped && (r.Action == "" || r.Action == "move") {
					histEntries = append(histEntries, history.Entry{
						Src:    r.Source,
						Dest:   r.Destination,
						Action: "move",
					})
				}
			}
			if len(histEntries) > 0 {
				_ = history.Save(histEntries)
			}
		}

		moved := 0
		copied := 0
		deleted := 0
		for _, r := range results {
			if r.Skipped {
				color.Yellow("skipped %s: %s", r.Source, r.SkipReason)
				continue
			}

			actionVerb := "moved"
			switch r.Action {
			case "copy":
				actionVerb = "copied"
			case "delete":
				actionVerb = "deleted"
			}

			if dryRun {
				if r.Action == "delete" {
					fmt.Printf("[dry-run] delete %q\n", r.Source)
				} else {
					act := r.Action
					if act == "" {
						act = "move"
					}
					fmt.Printf("[dry-run] %s %q → %s\n", act, r.Source, r.Destination)
				}
			} else {
				if r.Action == "delete" {
					fmt.Printf("deleted %q\n", r.Source)
					deleted++
				} else {
					fmt.Printf("%s %q → %s\n", actionVerb, r.Source, r.Destination)
					if r.Action == "copy" {
						copied++
					} else {
						moved++
					}
				}
			}
		}

		if !dryRun {
			summary := ""
			if moved > 0 {
				summary += fmt.Sprintf("%d file(s) moved", moved)
			}
			if copied > 0 {
				if summary != "" {
					summary += ", "
				}
				summary += fmt.Sprintf("%d file(s) copied", copied)
			}
			if deleted > 0 {
				if summary != "" {
					summary += ", "
				}
				summary += fmt.Sprintf("%d file(s) deleted", deleted)
			}
			if summary == "" {
				summary = "0 file(s) processed"
			}
			fmt.Printf("\n%s\n", summary)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview without moving files")
	runCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "organize subfolders recursively")
	runCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "bypass deletion confirmation prompt")
	rootCmd.AddCommand(runCmd)
}
