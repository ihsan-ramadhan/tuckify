package organizer

import (
	"fmt"
	"io"
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
	Action      string
}

func MatchRule(filename string, rules []config.Rule) *config.Rule {
	ext := strings.ToLower(filepath.Ext(filename))
	lower := strings.ToLower(filename)
	for i := range rules {
		for _, e := range rules[i].Extensions {
			if ext != "" && strings.ToLower(e) == ext {
				return &rules[i]
			}
		}
		for _, pattern := range rules[i].FilenamePatterns {
			if matched, _ := filepath.Match(strings.ToLower(pattern), lower); matched {
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

func copyFile(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func parseTemplates(pattern string, info os.FileInfo) string {
	if pattern == "" {
		return ""
	}
	ext := filepath.Ext(info.Name())
	base := strings.TrimSuffix(info.Name(), ext)
	t := info.ModTime()

	r := strings.NewReplacer(
		"{year}", t.Format("2006"),
		"{month}", t.Format("01"),
		"{day}", t.Format("02"),
		"{hour}", t.Format("15"),
		"{minute}", t.Format("04"),
		"{second}", t.Format("05"),
		"{base}", base,
		"{ext}", ext,
	)
	return r.Replace(pattern)
}

func processFile(src string, rule *config.Rule, conflictStrategy string, info os.FileInfo) (string, error) {
	if rule.Action == "delete" {
		if err := os.Remove(src); err != nil {
			return "", fmt.Errorf("delete file: %w", err)
		}
		return "/dev/null", nil
	}

	destDir := parseTemplates(rule.Destination, info)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}

	targetName := info.Name()
	if rule.Rename != "" {
		targetName = parseTemplates(rule.Rename, info)
	}

	dest, err := resolveDest(destDir, targetName, conflictStrategy)
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil
	}

	switch rule.Action {
	case "copy":
		if err := copyFile(src, dest); err != nil {
			return "", fmt.Errorf("copy file: %w", err)
		}
	default:
		if err := os.Rename(src, dest); err != nil {
			return "", fmt.Errorf("move file: %w", err)
		}
	}

	return dest, nil
}

func MoveFile(src, destDir, conflictStrategy string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	rule := &config.Rule{Action: "move", Destination: destDir}
	return processFile(src, rule, conflictStrategy, info)
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

		info, err := entry.Info()
		if err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: fmt.Sprintf("cannot stat: %v", err),
				Action:     rule.Action,
			})
			continue
		}

		if f, err := os.Open(src); err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: fmt.Sprintf("cannot open: %v", err),
				Action:     rule.Action,
			})
			continue
		} else {
			_ = f.Close()
		}

		if dryRun {
			dest := ""
			if rule.Action != "delete" {
				destDir := parseTemplates(rule.Destination, info)
				targetName := info.Name()
				if rule.Rename != "" {
					targetName = parseTemplates(rule.Rename, info)
				}
				dest = filepath.Join(destDir, targetName)
			}
			results = append(results, Result{
				Source:      src,
				Destination: dest,
				Action:      rule.Action,
			})
			continue
		}

		dest, err := processFile(src, rule, cfg.Settings.ConflictStrategy, info)
		if err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: err.Error(),
				Action:     rule.Action,
			})
			continue
		}
		if dest == "" {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: "file already exists",
				Action:     rule.Action,
			})
			continue
		}
		results = append(results, Result{Source: src, Destination: dest, Action: rule.Action})
	}
	return results, nil
}
