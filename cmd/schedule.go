package cmd

import (
	"fmt"
	"os"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/scheduler"
	"github.com/spf13/cobra"
)

var cronExpr string

var scheduleCmd = &cobra.Command{
	Use:   "schedule <folder>",
	Short: "Run organizer on a cron schedule",
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

		return scheduler.Run(folder, cronExpr, cfg)
	},
}

func init() {
	scheduleCmd.Flags().StringVar(&cronExpr, "cron", "", `cron expression, e.g. "0 9 * * *"`)
	scheduleCmd.MarkFlagRequired("cron")
	rootCmd.AddCommand(scheduleCmd)
}
