package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Settings Settings `toml:"settings"`
	Rules    []Rule   `toml:"rule"`
}

type Settings struct {
	ConflictStrategy string `toml:"conflict_strategy"`
}

type Rule struct {
	Name             string   `toml:"name"`
	Extensions       []string `toml:"extensions"`
	FilenamePatterns []string `toml:"filename_patterns"`
	Destination      string   `toml:"destination"`
	Action           string   `toml:"action"`
	Rename           string   `toml:"rename"`
	MinSize          string   `toml:"min_size"`
	MaxSize          string   `toml:"max_size"`
	MinAge           string   `toml:"min_age"`
	MaxAge           string   `toml:"max_age"`

	minSizeBytes   *int64
	maxSizeBytes   *int64
	minAgeDuration *time.Duration
	maxAgeDuration *time.Duration
}

func (r *Rule) MinSizeBytes() *int64 {
	return r.minSizeBytes
}

func (r *Rule) MaxSizeBytes() *int64 {
	return r.maxSizeBytes
}

func (r *Rule) MinAgeDuration() *time.Duration {
	return r.minAgeDuration
}

func (r *Rule) MaxAgeDuration() *time.Duration {
	return r.maxAgeDuration
}

func parseSizeString(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, nil
	}

	var numStr string
	var unit string
	for i, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			numStr += string(r)
		} else {
			unit = s[i:]
			break
		}
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size number %q: %w", numStr, err)
	}

	var multiplier int64
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB", "KIB", "K":
		multiplier = 1024
	case "MB", "MIB", "M":
		multiplier = 1024 * 1024
	case "GB", "GIB", "G":
		multiplier = 1024 * 1024 * 1024
	case "TB", "TIB", "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit %q", unit)
	}

	return int64(val * float64(multiplier)), nil
}

func parseAgeString(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, nil
	}

	var numStr string
	var unit string
	for i, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			numStr += string(r)
		} else {
			unit = s[i:]
			break
		}
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid age duration number %q: %w", numStr, err)
	}

	var multiplier time.Duration
	switch unit {
	case "h", "hour", "hours":
		multiplier = time.Hour
	case "d", "day", "days":
		multiplier = 24 * time.Hour
	case "w", "week", "weeks":
		multiplier = 7 * 24 * time.Hour
	case "m", "month", "months":
		multiplier = 30 * 24 * time.Hour
	case "y", "year", "years":
		multiplier = 365 * 24 * time.Hour
	default:
		dur, err := time.ParseDuration(s)
		if err == nil {
			return dur, nil
		}
		return 0, fmt.Errorf("unknown age unit %q", unit)
	}

	return time.Duration(val * float64(multiplier)), nil
}

func DefaultConfigPath() string {
	return ExpandHome("~/.tuckify/rules.toml")
}

func parseRuleSizes(r *Rule) error {
	if r.MinSize != "" {
		sz, err := parseSizeString(r.MinSize)
		if err != nil {
			return fmt.Errorf("invalid min_size: %w", err)
		}
		r.minSizeBytes = &sz
	}
	if r.MaxSize != "" {
		sz, err := parseSizeString(r.MaxSize)
		if err != nil {
			return fmt.Errorf("invalid max_size: %w", err)
		}
		r.maxSizeBytes = &sz
	}
	return nil
}

func parseRuleAges(r *Rule) error {
	if r.MinAge != "" {
		age, err := parseAgeString(r.MinAge)
		if err != nil {
			return fmt.Errorf("invalid min_age: %w", err)
		}
		r.minAgeDuration = &age
	}
	if r.MaxAge != "" {
		age, err := parseAgeString(r.MaxAge)
		if err != nil {
			return fmt.Errorf("invalid max_age: %w", err)
		}
		r.maxAgeDuration = &age
	}
	return nil
}

func validateAndParseRule(r *Rule) error {
	if r.Action == "" {
		r.Action = "move"
	}
	act := r.Action
	if act != "move" && act != "copy" && act != "delete" {
		return fmt.Errorf("invalid action %q", act)
	}
	if (act == "move" || act == "copy") && r.Destination == "" {
		return fmt.Errorf("destination is required for action %q", act)
	}
	r.Destination = ExpandHome(r.Destination)

	if err := parseRuleSizes(r); err != nil {
		return err
	}
	return parseRuleAges(r)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Config{Settings: Settings{ConflictStrategy: "rename"}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Settings.ConflictStrategy == "" {
		cfg.Settings.ConflictStrategy = "rename"
	}
	for i := range cfg.Rules {
		if err := validateAndParseRule(&cfg.Rules[i]); err != nil {
			return nil, fmt.Errorf("rule %q: %w", cfg.Rules[i].Name, err)
		}
	}
	return &cfg, nil
}

func ExpandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}
