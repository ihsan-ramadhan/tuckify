package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Update an existing schedule's cron, folder, or config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		
		name := args[0]

		found, err := store.Find(name)
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if found == nil {
			return fmt.Errorf("schedule %q not found", name)
		}

		updated := *found
		changed := map[string][2]string{}

		if cmd.Flags().Changed("cron") {
			newCron, _ := cmd.Flags().GetString("cron")
			if _, err := cron.ParseStandard(newCron); err != nil {
				return fmt.Errorf("invalid cron expression %q: %w", newCron, err)
			}
			changed["cron"] = [2]string{updated.Cron, newCron}
			updated.Cron = newCron
		}

		if cmd.Flags().Changed("folder") {
			newFolder, _ := cmd.Flags().GetString("folder")
			if newFolder == "" {
				return fmt.Errorf("folder path cannot be empty")
			}
			abs, err := filepath.Abs(newFolder)
			if err != nil {
				return fmt.Errorf("resolve folder path: %w", err)
			}
			if _, err := os.Stat(abs); os.IsNotExist(err) {
				return fmt.Errorf("folder not found: %s", abs)
			}
			changed["folder"] = [2]string{updated.Folder, abs}
			updated.Folder = abs
		}

		if cmd.Flags().Changed("config") {
			newConfig, _ := cmd.Flags().GetString("config")
			abs, err := filepath.Abs(newConfig)
			if err != nil {
				return fmt.Errorf("resolve config path: %w", err)
			}
			changed["config"] = [2]string{updated.Config, abs}
			updated.Config = abs
		}

		if cmd.Flags().Changed("recursive") {
			newRec, _ := cmd.Flags().GetBool("recursive")
			changed["recursive"] = [2]string{fmt.Sprintf("%t", updated.Recursive), fmt.Sprintf("%t", newRec)}
			updated.Recursive = newRec
		}

		if len(changed) == 0 {
			fmt.Printf("nothing to update for %q — use --cron, --folder, --config, or --recursive\n", name)
			return nil
		}

		if err := store.Upsert(updated); err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}

		fmt.Printf("updated %q\n", name)
		for field, vals := range changed {
			fmt.Printf("  %s: %s → %s\n", field, vals[0], vals[1])
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}
		online, _ := srv.Exists(name)
		if online {
			if err := srv.Uninstall(name); err != nil {
				return fmt.Errorf("stop service: %w", err)
			}
			if err := srv.Install(updated.Name, updated.GetFolders(), updated.Cron, updated.Config); err != nil {
				return fmt.Errorf("restart service: %w", err)
			}
			fmt.Println("  service restarted")
		}

		return nil
	},
}

func init() {
	editCmd.Flags().String("cron", "", `new cron expression, e.g. "0 9 * * *"`)
	editCmd.Flags().String("folder", "", "new folder path")
	editCmd.Flags().String("config", "", "new config path")
	editCmd.Flags().BoolP("recursive", "r", false, "toggle recursive mode (true/false)")
	rootCmd.AddCommand(editCmd)
}
