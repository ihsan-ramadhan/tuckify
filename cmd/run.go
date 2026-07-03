package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ihsan-ramadhan/tuckify/internal/ansi"
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
	Use:   "run [folders...]",
	Short: "Organize files in folders",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		warnNoRules(cfg, configPath)

		var folders []string
		if len(args) > 0 {
			folders = args
		} else {
			seen := make(map[string]bool)
			for _, r := range cfg.Rules {
				for _, loc := range r.LocationsExpanded() {
					if !seen[loc] {
						seen[loc] = true
						folders = append(folders, loc)
					}
				}
			}
			if len(folders) == 0 {
				return fmt.Errorf("no folders specified and no locations defined in config rules")
			}
		}

		for _, f := range folders {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				return fmt.Errorf("folder not found: %s", f)
			}
		}

		if !dryRun && !yesFlag {
			deletions := 0
			for _, f := range folders {
				previewResults, err := organizer.Organize(f, cfg, true, recursive)
				if err != nil {
					return err
				}
				for _, r := range previewResults {
					if !r.Skipped && r.Action == "delete" {
						deletions++
					}
				}
			}
			if deletions > 0 {
				ansi.PrintRed("warning: this operation will delete %d file(s).\n", deletions)
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

		var allResults []organizer.Result
		for _, f := range folders {
			results, err := organizer.Organize(f, cfg, dryRun, recursive)
			if err != nil {
				return err
			}
			allResults = append(allResults, results...)
		}

		// save history for undo (only on real run)
		if !dryRun {
			var histEntries []history.Entry
			for _, r := range allResults {
				if !r.Skipped && (r.Action == "" || r.Action == "move") {
					histEntries = append(histEntries, history.Entry{
						Src:    r.Source,
						Dest:   r.Destination,
						Action: "move",
					})
				}
			}
			if len(histEntries) > 0 {
				_ = history.Save(folders, histEntries)
			}
		}

		moved := 0
		copied := 0
		deleted := 0
		for _, r := range allResults {
			if r.Skipped {
				ansi.PrintYellow("skipped %s: %s\n", r.Source, r.SkipReason)
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
					deleted++
				} else {
					act := r.Action
					if act == "" {
						act = "move"
					}
					fmt.Printf("[dry-run] %s %q → %s\n", act, r.Source, r.Destination)
					if r.Action == "copy" {
						copied++
					} else {
						moved++
					}
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

		// Print summary for both dry-run and actual run
		summary := ""
		if dryRun {
			if moved > 0 {
				summary += fmt.Sprintf("%d file(s) would be moved", moved)
			}
			if copied > 0 {
				if summary != "" {
					summary += ", "
				}
				summary += fmt.Sprintf("%d file(s) would be copied", copied)
			}
			if deleted > 0 {
				if summary != "" {
					summary += ", "
				}
				summary += fmt.Sprintf("%d file(s) would be deleted", deleted)
			}
			if summary == "" {
				summary = "0 file(s) would be processed"
			}
			fmt.Printf("\n[dry-run] %s\n", summary)
		} else {
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
