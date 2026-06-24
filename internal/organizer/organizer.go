package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihsan-ramadhan/tuck/internal/config"
)

type Result struct {
	Source      string
	Destination string
	Skipped     bool
	SkipReason  string
}

func MoveFile(src, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}
	dest := filepath.Join(destDir, filepath.Base(src))
	if err := os.Rename(src, dest); err != nil {
		return "", fmt.Errorf("move file: %w", err)
	}
	return dest, nil
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
