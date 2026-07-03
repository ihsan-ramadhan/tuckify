package scheduler

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

func runTick(name string, folders []string, configPath string) ([]organizer.Result, error) {
	fmt.Printf("[%s] running organizer on folders: %s\n", time.Now().Format("2006-01-02 15:04:05"), strings.Join(folders, ", "))
	actualPath := ResolveConfigPath(name, configPath)
	cfg, err := config.Load(actualPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	recursive := false
	if s, err := store.Find(name); err == nil && s != nil {
		recursive = s.Recursive
	}

	var allResults []organizer.Result
	for _, folder := range folders {
		results, err := organizer.Organize(folder, cfg, false, recursive)
		if err != nil {
			return nil, fmt.Errorf("organize %q: %w", folder, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// resultActionVerb maps a Result.Action to its past-tense display verb.
func resultActionVerb(action string) string {
	switch action {
	case "copy":
		return "copied"
	case "delete":
		return "deleted"
	default:
		return "moved"
	}
}

// printTickResults prints one line per non-skipped/skipped result.
// ponytail: extracted from Run's closure to reduce cognitive complexity (SonarQube go:S3776).
func printTickResults(results []organizer.Result) {
	for _, r := range results {
		if r.Skipped {
			fmt.Fprintf(os.Stderr, "skipped %s: %s\n", r.Source, r.SkipReason)
			continue
		}
		if r.Action == "delete" {
			fmt.Printf("deleted %q\n", r.Source)
			continue
		}
		fmt.Printf("%s %q → %s\n", resultActionVerb(r.Action), r.Source, r.Destination)
	}
}

// summarizeTickResults builds the "N file(s) moved, M file(s) copied, ..." summary line.
func summarizeTickResults(results []organizer.Result) string {
	moved, copied, deleted := 0, 0, 0
	for _, r := range results {
		if r.Skipped {
			continue
		}
		switch r.Action {
		case "copy":
			copied++
		case "delete":
			deleted++
		default:
			moved++
		}
	}

	var parts []string
	if moved > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) moved", moved))
	}
	if copied > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) copied", copied))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) deleted", deleted))
	}
	if len(parts) == 0 {
		return "0 file(s) processed"
	}
	return strings.Join(parts, ", ")
}

func Run(name string, folders []string, expr, configPath string) error {
	c := cron.New()

	_, err := c.AddFunc(expr, func() {
		results, err := runTick(name, folders, configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		// ponytail: print moved to cmd layer (avoid duplication)
		printTickResults(results)
		fmt.Println(summarizeTickResults(results))
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
