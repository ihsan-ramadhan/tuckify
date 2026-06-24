package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ihsan-ramadhan/tuck/internal/config"
)

func makeConfig(strategy string, rules ...config.Rule) *config.Config {
	return &config.Config{
		Settings: config.Settings{ConflictStrategy: strategy},
		Rules:    rules,
	}
}

func TestMatchRule(t *testing.T) {
	rules := []config.Rule{
		{Extensions: []string{".pdf"}, Destination: "/docs"},
		{Extensions: []string{".jpg", ".PNG"}, Destination: "/pics"},
	}
	if r := MatchRule("file.PDF", rules); r == nil || r.Destination != "/docs" {
		t.Error("expected PDF rule match")
	}
	if r := MatchRule("file.png", rules); r == nil || r.Destination != "/pics" {
		t.Error("expected PNG rule match")
	}
	if r := MatchRule("noext", rules); r != nil {
		t.Error("expected no match for file without extension")
	}
}

func TestConflictRename(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	os.MkdirAll(dest, 0o755)

	os.WriteFile(filepath.Join(dest, "a.pdf"), []byte("orig"), 0o644)

	src := filepath.Join(dir, "a.pdf")
	os.WriteFile(src, []byte("new"), 0o644)

	cfg := makeConfig("rename", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if filepath.Base(results[0].Destination) != "a_1.pdf" {
		t.Errorf("expected a_1.pdf, got %s", filepath.Base(results[0].Destination))
	}
}

func TestDryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.pdf")
	os.WriteFile(src, []byte("data"), 0o644)

	cfg := makeConfig("rename", config.Rule{Extensions: []string{".pdf"}, Destination: "/docs"})
	results, err := Organize(dir, cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		t.Error("dry-run must not move the file")
	}
}
