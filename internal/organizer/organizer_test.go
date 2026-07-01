package organizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
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
	if MatchRule("noext", rules) != nil {
		t.Error("expected no match for file without extension")
	}
}

func TestMatchRuleFilenamePattern(t *testing.T) {
	rules := []config.Rule{
		{FilenamePatterns: []string{"*Modul*"}, Destination: "/modul"},
		{FilenamePatterns: []string{"Invoice_*"}, Destination: "/invoices"},
	}

	cases := []struct {
		file    string
		dest    string
		wantNil bool
	}{
		{"Modul1_Proyek.pdf", "/modul", false},
		{"Data_Modul_2.docx", "/modul", false},
		{"Modul", "/modul", false},        // no extension
		{"Invoice_2024.pdf", "/invoices", false},
		{"report.pdf", "", true},
		{"MODUL_test.txt", "/modul", false}, // case-insensitive
	}

	for _, c := range cases {
		r := MatchRule(c.file, rules)
		if c.wantNil && r != nil {
			t.Errorf("%s: expected no match, got %s", c.file, r.Destination)
		}
		if !c.wantNil && (r == nil || r.Destination != c.dest) {
			t.Errorf("%s: expected dest %s, got %v", c.file, c.dest, r)
		}
	}
}

func TestConflictRename(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	_ = os.MkdirAll(dest, 0o755)

	_ = os.WriteFile(filepath.Join(dest, "a.pdf"), []byte("orig"), 0o644)

	src := filepath.Join(dir, "a.pdf")
	_ = os.WriteFile(src, []byte("new"), 0o644)

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
	_ = os.WriteFile(src, []byte("data"), 0o644)

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

func TestConflictSkip(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	_ = os.MkdirAll(dest, 0o755)

	_ = os.WriteFile(filepath.Join(dest, "a.pdf"), []byte("orig"), 0o644)

	src := filepath.Join(dir, "a.pdf")
	_ = os.WriteFile(src, []byte("new"), 0o644)

	cfg := makeConfig("skip", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Error("expected result to be skipped")
	}
}

func TestConflictOverwrite(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")
	_ = os.MkdirAll(dest, 0o755)

	_ = os.WriteFile(filepath.Join(dest, "a.pdf"), []byte("orig"), 0o644)

	src := filepath.Join(dir, "a.pdf")
	_ = os.WriteFile(src, []byte("new"), 0o644)

	cfg := makeConfig("overwrite", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Skipped {
		t.Error("expected result not to be skipped")
	}
	if filepath.Base(results[0].Destination) != "a.pdf" {
		t.Errorf("expected destination to be a.pdf, got %s", filepath.Base(results[0].Destination))
	}
	data, _ := os.ReadFile(results[0].Destination)
	if string(data) != "new" {
		t.Errorf("expected overwrite content to be 'new', got %q", string(data))
	}
}

