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

func matchMetadata(rule *config.Rule, info os.FileInfo) bool {
	if rule.MinSizeBytes() == nil && rule.MaxSizeBytes() == nil &&
		rule.MinAgeDuration() == nil && rule.MaxAgeDuration() == nil {
		return true
	}
	if info == nil {
		return false
	}
	if minSz := rule.MinSizeBytes(); minSz != nil && info.Size() < *minSz {
		return false
	}
	if maxSz := rule.MaxSizeBytes(); maxSz != nil && info.Size() > *maxSz {
		return false
	}
	if minAge := rule.MinAgeDuration(); minAge != nil && time.Since(info.ModTime()) < *minAge {
		return false
	}
	if maxAge := rule.MaxAgeDuration(); maxAge != nil && time.Since(info.ModTime()) > *maxAge {
		return false
	}
	return true
}

func matchName(rule *config.Rule, filename string) bool {
	if len(rule.Extensions) == 0 && len(rule.FilenamePatterns) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	lower := strings.ToLower(filename)
	for _, e := range rule.Extensions {
		if ext != "" && strings.ToLower(e) == ext {
			return true
		}
	}
	for _, pattern := range rule.FilenamePatterns {
		if m, _ := filepath.Match(strings.ToLower(pattern), lower); m {
			return true
		}
	}
	return false
}

func MatchRule(filename string, info os.FileInfo, rules []config.Rule) *config.Rule {
	for i := range rules {
		if matchMetadata(&rules[i], info) && matchName(&rules[i], filename) {
			return &rules[i]
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

func isDuplicate(src, dest string) bool {
	srcHash, err1 := calculateHash(src)
	destHash, err2 := calculateHash(dest)
	return err1 == nil && err2 == nil && srcHash == destHash
}

func handleDuplicate(src, dest, action string) (bool, string, error) {
	if !isDuplicate(src, dest) {
		return false, "", nil
	}
	if action == "copy" {
		return true, "", nil
	}
	if err := os.Remove(src); err != nil {
		return true, "", fmt.Errorf("delete duplicate source: %w", err)
	}
	return true, dest, nil
}

func executeFileAction(src, dest, action string) error {
	if action == "copy" {
		if err := copyFile(src, dest); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}
		return nil
	}
	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("move file: %w", err)
	}
	return nil
}

func buildAndResolveDest(src string, rule *config.Rule, conflictStrategy string, info os.FileInfo) (string, bool, error) {
	destDir := parseTemplates(rule.Destination, info)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", false, fmt.Errorf("create destination: %w", err)
	}

	targetName := info.Name()
	if rule.Rename != "" {
		targetName = parseTemplates(rule.Rename, info)
	}

	dest := filepath.Join(destDir, targetName)
	if _, err := os.Stat(dest); err == nil {
		if conflictStrategy == "delete_duplicate" {
			handled, retDest, err := handleDuplicate(src, dest, rule.Action)
			if err != nil {
				return "", false, err
			}
			if handled {
				return retDest, true, nil
			}
			conflictStrategy = "rename"
		}
	}

	resolvedDest, err := resolveDest(destDir, targetName, conflictStrategy)
	return resolvedDest, false, err
}

func processFile(src string, rule *config.Rule, conflictStrategy string, info os.FileInfo) (string, error) {
	if rule.Action == "delete" {
		if err := os.Remove(src); err != nil {
			return "", fmt.Errorf("delete file: %w", err)
		}
		return "/dev/null", nil
	}

	resolvedDest, handled, err := buildAndResolveDest(src, rule, conflictStrategy, info)
	if err != nil {
		return "", err
	}
	if resolvedDest == "" || handled {
		return resolvedDest, nil
	}

	if err := executeFileAction(src, resolvedDest, rule.Action); err != nil {
		return "", err
	}

	return resolvedDest, nil
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

func sortDirs(dirs []string) {
	for i := 0; i < len(dirs); i++ {
		for j := i + 1; j < len(dirs); j++ {
			if len(dirs[i]) < len(dirs[j]) {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			}
		}
	}
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

	sortDirs(dirs)

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

func organizeFile(src string, cfg *config.Config, dryRun bool) (Result, bool) {
	name := filepath.Base(src)
	info, err := os.Stat(src)
	if err != nil {
		return Result{
			Source:     src,
			Skipped:    true,
			SkipReason: fmt.Sprintf("cannot stat: %v", err),
		}, true
	}

	rule := MatchRule(name, info, cfg.Rules)
	if rule == nil {
		return Result{}, false
	}

	if f, err := os.Open(src); err != nil {
		return Result{
			Source:     src,
			Skipped:    true,
			SkipReason: fmt.Sprintf("cannot open: %v", err),
			Action:     rule.Action,
		}, true
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
		return Result{
			Source:      src,
			Destination: dest,
			Action:      rule.Action,
		}, true
	}

	dest, err := processFile(src, rule, cfg.Settings.ConflictStrategy, info)
	if err != nil {
		return Result{
			Source:     src,
			Skipped:    true,
			SkipReason: err.Error(),
			Action:     rule.Action,
		}, true
	}
	if dest == "" {
		return Result{
			Source:     src,
			Skipped:    true,
			SkipReason: "file already exists",
			Action:     rule.Action,
		}, true
	}
	return Result{Source: src, Destination: dest, Action: rule.Action}, true
}

func Organize(folder string, cfg *config.Config, dryRun bool, recursive bool) ([]Result, error) {
	files, err := listFiles(folder, recursive)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	var results []Result
	for _, src := range files {
		if res, processed := organizeFile(src, cfg, dryRun); processed {
			results = append(results, res)
		}
	}

	if recursive && !dryRun {
		_ = deleteEmptyDirs(folder)
	}

	return results, nil
}
