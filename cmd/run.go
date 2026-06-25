package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/organizer"
	"github.com/spf13/cobra"
)

var dryRun bool

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

		results, err := organizer.Organize(folder, cfg, dryRun)
		if err != nil {
			return err
		}

		moved := 0
		for _, r := range results {
			if r.Skipped {
				color.Yellow("skipped %s: %s", r.Source, r.SkipReason)
				continue
			}
			if dryRun {
				fmt.Printf("[dry-run] %q → %s\n", r.Source, r.Destination)
			} else {
				fmt.Printf("moved %q → %s\n", r.Source, r.Destination)
				moved++
			}
		}

		if !dryRun {
			fmt.Printf("\n%d file(s) moved\n", moved)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview without moving files")
	rootCmd.AddCommand(runCmd)
}
