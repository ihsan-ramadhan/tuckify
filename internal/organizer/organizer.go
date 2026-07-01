package organizer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
)

type Result struct {
	Source      string
	Destination string
	Skipped     bool
	SkipReason  string
	Action      string
}

func MatchRule(filename string, info os.FileInfo, rules []config.Rule) *config.Rule {
	ext := strings.ToLower(filepath.Ext(filename))
	lower := strings.ToLower(filename)
	for i := range rules {
		if (rules[i].MinSizeBytes() != nil || rules[i].MaxSizeBytes() != nil ||
			rules[i].MinAgeDuration() != nil || rules[i].MaxAgeDuration() != nil) && info == nil {
			continue
		}

		if info != nil {
			if minSz := rules[i].MinSizeBytes(); minSz != nil && info.Size() < *minSz {
				continue
			}
			if maxSz := rules[i].MaxSizeBytes(); maxSz != nil && info.Size() > *maxSz {
				continue
			}
			if minAge := rules[i].MinAgeDuration(); minAge != nil && time.Since(info.ModTime()) < *minAge {
				continue
			}
			if maxAge := rules[i].MaxAgeDuration(); maxAge != nil && time.Since(info.ModTime()) > *maxAge {
				continue
			}
		}

		hasNameFilter := len(rules[i].Extensions) > 0 || len(rules[i].FilenamePatterns) > 0
		nameMatched := false
		if hasNameFilter {
			for _, e := range rules[i].Extensions {
				if ext != "" && strings.ToLower(e) == ext {
					nameMatched = true
					break
				}
			}
			if !nameMatched {
				for _, pattern := range rules[i].FilenamePatterns {
					if m, _ := filepath.Match(strings.ToLower(pattern), lower); m {
						nameMatched = true
						break
					}
				}
			}
			if !nameMatched {
				continue
			}
		}

		return &rules[i]
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

var templateRegex = regexp.MustCompile(`\{([a-zA-Z]+)(?::([a-zA-Z]+))?\}`)

func slugify(s string) string {
	s = strings.ToLower(s)
	var builder strings.Builder
	lastWasDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastWasDash = false
		} else if r == ' ' || r == '-' || r == '_' || r == '.' {
			if !lastWasDash && builder.Len() > 0 {
				builder.WriteRune('-')
				lastWasDash = true
			}
		}
	}
	return strings.TrimSuffix(builder.String(), "-")
}

func parseTemplates(pattern string, info os.FileInfo) string {
	if pattern == "" {
		return ""
	}
	ext := filepath.Ext(info.Name())
	base := strings.TrimSuffix(info.Name(), ext)
	t := info.ModTime()

	return templateRegex.ReplaceAllStringFunc(pattern, func(m string) string {
		match := templateRegex.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		key := match[1]
		var modifier string
		if len(match) > 2 {
			modifier = match[2]
		}

		var val string
		switch key {
		case "year":
			val = t.Format("2006")
		case "month":
			val = t.Format("01")
		case "day":
			val = t.Format("02")
		case "hour":
			val = t.Format("15")
		case "minute":
			val = t.Format("04")
		case "second":
			val = t.Format("05")
		case "base":
			val = base
		case "ext":
			val = ext
		default:
			return m
		}

		switch modifier {
		case "lower":
			val = strings.ToLower(val)
		case "upper":
			val = strings.ToUpper(val)
		case "slug":
			val = slugify(val)
		}

		return val
	})
}

func calculateHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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

	dest := filepath.Join(destDir, targetName)
	if _, err := os.Stat(dest); err == nil {
		if conflictStrategy == "delete_duplicate" {
			srcHash, err1 := calculateHash(src)
			destHash, err2 := calculateHash(dest)
			if err1 == nil && err2 == nil && srcHash == destHash {
				if rule.Action == "copy" {
					return "", nil
				}
				if err := os.Remove(src); err != nil {
					return "", fmt.Errorf("delete duplicate source: %w", err)
				}
				return dest, nil
			}
			conflictStrategy = "rename"
		}
	}

	resolvedDest, err := resolveDest(destDir, targetName, conflictStrategy)
	if err != nil {
		return "", err
	}
	if resolvedDest == "" {
		return "", nil
	}

	switch rule.Action {
	case "copy":
		if err := copyFile(src, resolvedDest); err != nil {
			return "", fmt.Errorf("copy file: %w", err)
		}
	default:
		if err := os.Rename(src, resolvedDest); err != nil {
			return "", fmt.Errorf("move file: %w", err)
		}
	}

	return resolvedDest, nil
}

func MoveFile(src, destDir, conflictStrategy string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", err
	}
	rule := &config.Rule{Action: "move", Destination: destDir}
	return processFile(src, rule, conflictStrategy, info)
}

func listFiles(folder string, recursive bool) ([]string, error) {
	var files []string
	if !recursive {
		entries, err := os.ReadDir(folder)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(folder, entry.Name()))
			}
		}
		return files, nil
	}

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func deleteEmptyDirs(root string) error {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(dirs); i++ {
		for j := i + 1; j < len(dirs); j++ {
			if len(dirs[i]) < len(dirs[j]) {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			}
		}
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
	return nil
}

func Organize(folder string, cfg *config.Config, dryRun bool, recursive bool) ([]Result, error) {
	files, err := listFiles(folder, recursive)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	var results []Result
	for _, src := range files {
		name := filepath.Base(src)
		info, err := os.Stat(src)
		if err != nil {
			results = append(results, Result{
				Source:     src,
				Skipped:    true,
				SkipReason: fmt.Sprintf("cannot stat: %v", err),
			})
			continue
		}

		rule := MatchRule(name, info, cfg.Rules)
		if rule == nil {
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

	if recursive && !dryRun {
		_ = deleteEmptyDirs(folder)
	}

	return results, nil
}
