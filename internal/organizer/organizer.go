package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
)

type Result struct {
	Source      string
	Destination string
	Skipped     bool
	SkipReason  string
}

func MatchRule(filename string, rules []config.Rule) *config.Rule {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil
	}
	for i := range rules {
		for _, e := range rules[i].Extensions {
			if strings.ToLower(e) == ext {
				return &rules[i]
			}
		}
	}
	return nil
}

func resolveDest(destDir, filename, strategy string) (string, error) {
	dest := filepath.Join(destDir, filename)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest, nil
	}
	switch strategy {
	case "skip":
		return "", nil
	case "overwrite":
		return dest, nil
	default:
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		for i := 1; ; i++ {
			candidate := filepath.Join(destDir, fmt.Sprintf("%s_%d%s", base, i, ext))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				return candidate, nil
			}
		}
	}
}

func MoveFile(src, destDir, conflictStrategy string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}
	dest, err := resolveDest(destDir, filepath.Base(src), conflictStrategy)
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil
	}
	if err := os.Rename(src, dest); err != nil {
		return "", fmt.Errorf("move file: %w", err)
	}
	return dest, nil
}

func Organize(folder string, cfg *config.Config, dryRun bool) ([]Result, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("read folder: %w", err)
	}

	var results []Result
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		rule := MatchRule(name, cfg.Rules)
		if rule == nil {
			continue
		}

		src := filepath.Join(folder, name)

		if f, err := os.Open(src); err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: fmt.Sprintf("cannot open: %v", err),
			})
			continue
		} else {
			_ = f.Close()
		}

		if dryRun {
			results = append(results, Result{
				Source:      src,
				Destination: filepath.Join(rule.Destination, name),
			})
			continue
		}

		dest, err := MoveFile(src, rule.Destination, cfg.Settings.ConflictStrategy)
		if err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: err.Error(),
			})
			continue
		}
		results = append(results, Result{Source: src, Destination: dest})
	}
	return results, nil
}
