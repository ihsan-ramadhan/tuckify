package organizer

import (
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
