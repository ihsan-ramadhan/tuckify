package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/scheduler"
	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var (
	cronExpr          string
	scheduleRun       bool
	scheduleStart     bool
	scheduleForce     bool
	scheduleYes       bool
	scheduleRecursive bool
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule <name> [folders...]",
	Short: "Save a named schedule (use --run to test interactively, --start to activate as service)",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		
		name := args[0]
		var folders []string
		for _, f := range args[1:] {
			abs, err := filepath.Abs(f)
			if err != nil {
				return fmt.Errorf("resolve folder path %q: %w", f, err)
			}
			if _, err := os.Stat(abs); os.IsNotExist(err) {
				return fmt.Errorf("folder not found: %s", abs)
			}
			folders = append(folders, abs)
		}

		if _, err := cron.ParseStandard(cronExpr); err != nil {
			return fmt.Errorf("invalid cron expression %q: %w", cronExpr, err)
		}

		var absConfig string
		if cmd.Flags().Changed("config") {
			var err error
			absConfig, err = filepath.Abs(configPath)
			if err != nil {
				return fmt.Errorf("resolve config path: %w", err)
			}
		}

		// Check if schedule already exists
		existing, err := store.Find(name)
		if err != nil {
			return fmt.Errorf("load schedules: %w", err)
		}
		if existing != nil && !scheduleForce && !scheduleRun {
			var response string
			fmt.Printf("schedule %q already exists — overwrite? [y/N]: ", name)
			_, err := fmt.Scanln(&response)
			if err != nil || (strings.ToLower(strings.TrimSpace(response)) != "y" && strings.ToLower(strings.TrimSpace(response)) != "yes") {
				return fmt.Errorf("operation cancelled")
			}
		}

		if err := store.Upsert(store.Schedule{
			Name:      name,
			Folders:   folders,
			Cron:      cronExpr,
			Config:    absConfig,
			Recursive: scheduleRecursive,
			Yes:       scheduleYes,
		}); err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}

		fmt.Printf("saved schedule %q\n", name)

		actualConfigPath := scheduler.ResolveConfigPath(name, absConfig)
		cfg, err := config.Load(actualConfigPath)
		if err != nil {
			return err
		}
		warnNoRules(cfg, actualConfigPath)

		if scheduleRun {
			// Run interactive scheduler (blocking)
			return scheduler.Run(name, folders, cronExpr, absConfig)
		}
		
		if scheduleStart {
			// Start as background service
			srv, err := service.NewService()
			if err != nil {
				return err
			}
			if err := srv.Install(name, folders, cronExpr, absConfig); err != nil {
				return fmt.Errorf("start service: %w", err)
			}
			fmt.Printf("started %q\n", name)
			return nil
		}
		
		// Only show hint if neither --run nor --start was used
		fmt.Printf("  run 'tuckify start %s' to activate as a background service\n", name)
		return nil
	},
}

func init() {
	scheduleCmd.Flags().StringVar(&cronExpr, "cron", "", `cron expression, e.g. "0 9 * * *"`)
	scheduleCmd.Flags().BoolVar(&scheduleRun, "run", false, "also start interactive scheduler after saving")
	scheduleCmd.Flags().BoolVar(&scheduleStart, "start", false, "also start as background service after saving")
	scheduleCmd.Flags().BoolVar(&scheduleForce, "force", false, "overwrite existing schedule without prompting")
	scheduleCmd.Flags().BoolVarP(&scheduleYes, "yes", "y", false, "skip deletion confirmation prompt")
	scheduleCmd.Flags().BoolVarP(&scheduleRecursive, "recursive", "r", false, "organize subfolders recursively")
	_ = scheduleCmd.MarkFlagRequired("cron")
	rootCmd.AddCommand(scheduleCmd)
}
