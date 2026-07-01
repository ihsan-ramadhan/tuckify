package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
}

func DefaultConfigPath() string {
	return ExpandHome("~/.tuckify/rules.toml")
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
		if cfg.Rules[i].Action == "" {
			cfg.Rules[i].Action = "move"
		}
		act := cfg.Rules[i].Action
		if act != "move" && act != "copy" && act != "delete" {
			return nil, fmt.Errorf("invalid action %q for rule %q", act, cfg.Rules[i].Name)
		}
		if (act == "move" || act == "copy") && cfg.Rules[i].Destination == "" {
			return nil, fmt.Errorf("destination is required for action %q in rule %q", act, cfg.Rules[i].Name)
		}
		cfg.Rules[i].Destination = ExpandHome(cfg.Rules[i].Destination)
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
