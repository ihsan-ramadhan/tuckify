package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.toml")
	content := `
[settings]
conflict_strategy = "skip"

[[rule]]
name        = "PDF"
extensions  = [".pdf"]
destination = "~/Documents/PDF"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Settings.ConflictStrategy != "skip" {
		t.Errorf("conflict_strategy = %q, want skip", cfg.Settings.ConflictStrategy)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(cfg.Rules))
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "Documents/PDF")
	if cfg.Rules[0].Destination != want {
		t.Errorf("destination = %q, want %q", cfg.Rules[0].Destination, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("missing file should yield no rules, got %d", len(cfg.Rules))
	}
	if cfg.Settings.ConflictStrategy != "rename" {
		t.Errorf("default strategy = %q, want rename", cfg.Settings.ConflictStrategy)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.toml")
	if err := os.WriteFile(path, []byte("this is = = not valid"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("invalid TOML should return an error")
	}
}

func TestExpandHomeNoPrefix(t *testing.T) {
	if got := ExpandHome("/abs/path"); got != "/abs/path" {
		t.Errorf("got %q, want unchanged", got)
	}

	home, _ := os.UserHomeDir()
	if got := ExpandHome("~"); got != home {
		t.Errorf("got %q, want %q", got, home)
	}
}

func TestLoadDirectoryError(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil {
		t.Error("expected error when loading directory as config, got nil")
	}
}

func TestLoadActions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.toml")
	content := `
[[rule]]
name        = "Move PDF"
extensions  = [".pdf"]
destination = "~/Documents/PDF"

[[rule]]
name        = "Copy Doc"
extensions  = [".docx"]
destination = "~/Documents/Office"
action      = "copy"

[[rule]]
name        = "Delete Temp"
extensions  = [".tmp"]
action      = "delete"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(cfg.Rules))
	}
	if cfg.Rules[0].Action != "move" {
		t.Errorf("rule 0 action = %q, want move", cfg.Rules[0].Action)
	}
	if cfg.Rules[1].Action != "copy" {
		t.Errorf("rule 1 action = %q, want copy", cfg.Rules[1].Action)
	}
	if cfg.Rules[2].Action != "delete" {
		t.Errorf("rule 2 action = %q, want delete", cfg.Rules[2].Action)
	}
}

func TestLoadInvalidAction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.toml")
	content := `
[[rule]]
name   = "Bad Action"
action = "burn"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid action, got nil")
	}
}

func TestLoadMissingDestination(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.toml")
	content := `
[[rule]]
name   = "No Dest Move"
action = "move"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for missing destination in move action, got nil")
	}
}

func TestParseSizeString(t *testing.T) {
	cases := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"500", 500, false},
		{"500B", 500, false},
		{"10KB", 10240, false},
		{"1.5MB", 1572864, false},
		{"2gb", 2147483648, false},
		{"invalid", 0, true},
		{"10XX", 0, true},
	}

	for _, c := range cases {
		got, err := parseSizeString(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseSizeString(%q) expected error, got nil", c.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseSizeString(%q) unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Errorf("parseSizeString(%q) = %d, want %d", c.input, got, c.want)
			}
		}
	}
}

func TestParseAgeString(t *testing.T) {
	cases := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"24h", 24 * time.Hour, false},
		{"2d", 48 * time.Hour, false},
		{"1w", 7 * 24 * time.Hour, false},
		{"1m", 30 * 24 * time.Hour, false},
		{"1y", 365 * 24 * time.Hour, false},
		{"invalid", 0, true},
	}

	for _, c := range cases {
		got, err := parseAgeString(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseAgeString(%q) expected error, got nil", c.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseAgeString(%q) unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Errorf("parseAgeString(%q) = %v, want %v", c.input, got, c.want)
			}
		}
	}
}

func TestLoadSizeAgeValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rules.toml")
	content := `
[[rule]]
name        = "Filter size and age"
extensions  = [".log"]
destination = "/logs"
min_size    = "10KB"
max_size    = "1MB"
min_age     = "7d"
max_age     = "30d"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(cfg.Rules))
	}
	r := cfg.Rules[0]
	if r.MinSizeBytes() == nil || *r.MinSizeBytes() != 10240 {
		t.Errorf("min_size = %v, want 10240", r.MinSizeBytes())
	}
	if r.MaxSizeBytes() == nil || *r.MaxSizeBytes() != 1024*1024 {
		t.Errorf("max_size = %v, want 1048576", r.MaxSizeBytes())
	}
	if r.MinAgeDuration() == nil || *r.MinAgeDuration() != 7*24*time.Hour {
		t.Errorf("min_age = %v, want 168h", r.MinAgeDuration())
	}
	if r.MaxAgeDuration() == nil || *r.MaxAgeDuration() != 30*24*time.Hour {
		t.Errorf("max_age = %v, want 720h", r.MaxAgeDuration())
	}
}


