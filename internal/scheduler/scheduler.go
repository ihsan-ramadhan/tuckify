package scheduler

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/organizer"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/robfig/cron/v3"
)

func ResolveConfigPath(name, configPath string) string {
	if configPath != "" {
		return configPath
	}
	home, err := os.UserHomeDir()
	if err == nil {
		custom := filepath.Join(home, ".tuckify", name+".toml")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	return config.DefaultConfigPath()
}

func runTick(name, folder, configPath string) ([]organizer.Result, error) {
	fmt.Printf("[%s] running organizer on %s\n", time.Now().Format("2006-01-02 15:04:05"), folder)
	actualPath := ResolveConfigPath(name, configPath)
	cfg, err := config.Load(actualPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	recursive := false
	if s, err := store.Find(name); err == nil && s != nil {
		recursive = s.Recursive
	}

	return organizer.Organize(folder, cfg, false, recursive)
}

func Run(name, folder, expr, configPath string) error {
	c := cron.New()

	_, err := c.AddFunc(expr, func() {
		results, err := runTick(name, folder, configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		// ponytail: print moved to cmd layer (avoid duplication)
		for _, r := range results {
			if r.Skipped {
				fmt.Fprintf(os.Stderr, "skipped %s: %s\n", r.Source, r.SkipReason)
				continue
			}
			actionVerb := "moved"
			switch r.Action {
			case "copy":
				actionVerb = "copied"
			case "delete":
				actionVerb = "deleted"
			}
			if r.Action == "delete" {
				fmt.Printf("deleted %q\n", r.Source)
			} else {
				fmt.Printf("%s %q → %s\n", actionVerb, r.Source, r.Destination)
			}
			}
			moved, copied, deleted := 0, 0, 0
			for _, r := range results {
				if !r.Skipped {
					switch r.Action {
					case "copy":
						copied++
					case "delete":
						deleted++
					default:
						moved++
					}
				}
			}
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
			fmt.Println(summary)
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	c.Start()
	fmt.Printf("scheduler started — press Ctrl+C to stop\n")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	c.Stop()
	fmt.Println("scheduler stopped")
	return nil
}
