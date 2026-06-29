package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/scheduler"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var cronExpr string

var scheduleCmd = &cobra.Command{
	Use:   "schedule <name> <folder>",
	Short: "Run organizer on a cron schedule and save to list",
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

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
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

		return scheduler.Run(folder, cronExpr, cfg)
	},
}

func init() {
	scheduleCmd.Flags().StringVar(&cronExpr, "cron", "", `cron expression, e.g. "0 9 * * *"`)
	scheduleCmd.MarkFlagRequired("cron")
	rootCmd.AddCommand(scheduleCmd)
}
