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

