package cmd

import (
	"fmt"
	"os"

	"github.com/ihsan-ramadhan/tuckify/internal/ansi"
	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var validateSchedule string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check rules.toml (and optionally a saved schedule) for errors without running anything",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if validateSchedule != "" {
			return validateNamedSchedule(validateSchedule)
		}
		return validateConfigFile(configPath)
	},
}

// validateConfigFile loads and parses rules.toml, reporting syntax and
// semantic errors (invalid action, missing destination, bad conflict
// strategy, invalid size/age/regex, etc.) without organizing any files.
func validateConfigFile(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		ansi.PrintRed("invalid config: %v\n", err)
		return err
	}

	if len(cfg.Rules) == 0 {
		ansi.PrintYellow("warning: no rules defined in %s\n", path)
		return nil
	}

	fmt.Printf("%s is valid — %d rule(s), conflict_strategy: %s\n", path, len(cfg.Rules), cfg.Settings.ConflictStrategy)
	for _, r := range cfg.Rules {
		name := r.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Printf("  ✓ %s\n", name)
	}
	return nil
}

// validateNamedSchedule cross-checks a saved schedule: its config file,
// cron expression, and that all its folders still exist on disk.
func validateNamedSchedule(name string) error {
	sched, err := store.Find(name)
	if err != nil {
		return fmt.Errorf("load schedules: %w", err)
	}
	if sched == nil {
		return fmt.Errorf("schedule %q not found", name)
	}

	ok := true

	if _, err := cron.ParseStandard(sched.Cron); err != nil {
		ansi.PrintRed("invalid cron expression %q: %v\n", sched.Cron, err)
		ok = false
	} else {
		fmt.Printf("  ✓ cron: %s\n", sched.Cron)
	}

	for _, f := range sched.GetFolders() {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			ansi.PrintRed("folder not found: %s\n", f)
			ok = false
			continue
		}
		fmt.Printf("  ✓ folder: %s\n", f)
	}

	cfgPath := sched.Config
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	if err := validateConfigFile(cfgPath); err != nil {
		ok = false
	}

	if !ok {
		return fmt.Errorf("schedule %q has validation errors", name)
	}
	fmt.Printf("schedule %q is valid\n", name)
	return nil
}

func init() {
	validateCmd.Flags().StringVar(&validateSchedule, "schedule", "", "also validate a saved schedule's cron expression and folders")
	rootCmd.AddCommand(validateCmd)
}
