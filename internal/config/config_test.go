package config

import (
	"os"
	"path/filepath"
	"testing"
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
}
