package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/scheduler"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var cronExpr string
var scheduleRun bool

var scheduleCmd = &cobra.Command{
	Use:   "schedule <name> <folder>",
	Short: "Save a named schedule (use --run to also start interactively)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		folder, err := filepath.Abs(args[1])
		if err != nil {
			return fmt.Errorf("resolve folder path: %w", err)
		}

		if _, err := os.Stat(folder); os.IsNotExist(err) {
			return fmt.Errorf("folder not found: %s", folder)
		}

		if _, err := cron.ParseStandard(cronExpr); err != nil {
			return fmt.Errorf("invalid cron expression %q: %w", cronExpr, err)
		}

		var absConfig string
		if cmd.Flags().Changed("config") {
			absConfig, err = filepath.Abs(configPath)
			if err != nil {
				return fmt.Errorf("resolve config path: %w", err)
			}
		}

		if err := store.Upsert(store.Schedule{
			Name:   name,
			Folder: folder,
			Cron:   cronExpr,
			Config: absConfig,
		}); err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}

		fmt.Printf("saved schedule %q\n", name)
		fmt.Printf("  run 'tuckify start %s' to activate as a background service\n", name)

		actualConfigPath := scheduler.ResolveConfigPath(name, absConfig)
		cfg, err := config.Load(actualConfigPath)
		if err != nil {
			return err
		}
		warnNoRules(cfg, actualConfigPath)

		if !scheduleRun {
			return nil
		}
		return scheduler.Run(name, folder, cronExpr, absConfig)
	},
}

func init() {
	scheduleCmd.Flags().StringVar(&cronExpr, "cron", "", `cron expression, e.g. "0 9 * * *"`)
	scheduleCmd.Flags().BoolVar(&scheduleRun, "run", false, "also start interactive scheduler after saving")
	_ = scheduleCmd.MarkFlagRequired("cron")
	rootCmd.AddCommand(scheduleCmd)
}
