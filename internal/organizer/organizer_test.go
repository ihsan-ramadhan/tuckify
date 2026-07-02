package organizer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
	if r := MatchRule("file.PDF", nil, rules, "/test"); r == nil || r.Destination != "/docs" {
		t.Error("expected PDF rule match")
	}
	if r := MatchRule("file.png", nil, rules, "/test"); r == nil || r.Destination != "/pics" {
		t.Error("expected PNG rule match")
	}
	if MatchRule("noext", nil, rules, "/test") != nil {
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
		r := MatchRule(c.file, nil, rules, "/test")
		if c.wantNil && r != nil {
			t.Errorf("%s: expected no match, got %s", c.file, r.Destination)
		}
		if !c.wantNil && (r == nil || r.Destination != c.dest) {
			t.Errorf("%s: expected dest %s, got %v", c.file, c.dest, r)
		}
	}
}

func setupConflictTest(t *testing.T) (dir, src, dest string) {
	t.Helper()
	dir = t.TempDir()
	dest = filepath.Join(dir, "dest")
	_ = os.MkdirAll(dest, 0o755)

	_ = os.WriteFile(filepath.Join(dest, "a.pdf"), []byte("orig"), 0o644)

	src = filepath.Join(dir, "a.pdf")
	_ = os.WriteFile(src, []byte("new"), 0o644)
	return dir, src, dest
}

func TestConflictRename(t *testing.T) {
	dir, _, dest := setupConflictTest(t)

	cfg := makeConfig("rename", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false, false)
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
	results, err := Organize(dir, cfg, true, false)
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
	dir, _, dest := setupConflictTest(t)

	cfg := makeConfig("skip", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false, false)
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
	dir, _, dest := setupConflictTest(t)

	cfg := makeConfig("overwrite", config.Rule{Extensions: []string{".pdf"}, Destination: dest})
	results, err := Organize(dir, cfg, false, false)
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

func TestParseTemplates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "My Resume 2024.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		pattern string
		want    string
	}{
		{"{base}_new{ext}", "My Resume 2024_new.txt"},
		{"{base:lower}{ext}", "my resume 2024.txt"},
		{"{base:upper}{ext}", "MY RESUME 2024.txt"},
		{"{base:slug}{ext}", "my-resume-2024.txt"},
		{"{year}/{month}/{day}", info.ModTime().Format("2006/01/02")},
		{"{hour}:{minute}:{second}", info.ModTime().Format("15:04:05")},
		{"static_text", "static_text"},
	}

	for _, c := range cases {
		got := parseTemplates(c.pattern, info)
		if got != c.want {
			t.Errorf("parseTemplates(%q) = %q, want %q", c.pattern, got, c.want)
		}
	}
}

func TestOrganizeCopyAction(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "backup")

	src := filepath.Join(dir, "test.pdf")
	if err := os.WriteFile(src, []byte("pdfcontent"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := makeConfig("rename", config.Rule{
		Name:        "Copy PDF",
		Extensions:  []string{".pdf"},
		Destination: dest,
		Action:      "copy",
	})

	results, err := Organize(dir, cfg, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		t.Error("source file should not be deleted for copy action")
	}

	destFile := filepath.Join(dest, "test.pdf")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("destination file should exist for copy action")
	}

	data, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "pdfcontent" {
		t.Errorf("expected copied content to be 'pdfcontent', got %q", string(data))
	}
}

func TestOrganizeDeleteAction(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.tmp")
	if err := os.WriteFile(src, []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := makeConfig("rename", config.Rule{
		Name:       "Delete Temp",
		Extensions: []string{".tmp"},
		Action:     "delete",
	})

	results, err := Organize(dir, cfg, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source file should be deleted for delete action")
	}
}

func TestOrganizeRenameAndDestinationTemplate(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	year := info.ModTime().Format("2006")
	month := info.ModTime().Format("01")

	destDir := filepath.Join(dir, year, month)
	cfg := makeConfig("rename", config.Rule{
		Name:        "Template Rule",
		Extensions:  []string{".pdf"},
		Destination: filepath.Join(dir, "{year}", "{month}"),
		Rename:      "renamed_{base}{ext}",
		Action:      "move",
	})

	results, err := Organize(dir, cfg, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	expectedDest := filepath.Join(destDir, "renamed_report.pdf")
	if results[0].Destination != expectedDest {
		t.Errorf("expected destination %q, got %q", expectedDest, results[0].Destination)
	}

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Error("expected moved file to exist at templated destination")
	}
}

func runTestOrganize(t *testing.T, dir, tomlContent string) []Result {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rules.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgParsed, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	results, err := Organize(dir, cfgParsed, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	return results
}

func TestOrganizeSizeFilter(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")

	smallFile := filepath.Join(dir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}

	largeFile := filepath.Join(dir, "large.txt")
	if err := os.WriteFile(largeFile, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}

	content := `
[[rule]]
name        = "Large Files Only"
extensions  = [".txt"]
destination = '` + dest + `'
min_size    = "50B"
`
	results := runTestOrganize(t, dir, content)
	if filepath.Base(results[0].Source) != "large.txt" {
		t.Errorf("expected large.txt to be organized, got %s", filepath.Base(results[0].Source))
	}
}

func TestOrganizeAgeFilter(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "dest")

	newFile := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := filepath.Join(dir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, twoHoursAgo, twoHoursAgo); err != nil {
		t.Fatal(err)
	}

	content := `
[[rule]]
name        = "Old Files Only"
extensions  = [".txt"]
destination = '` + dest + `'
min_age     = "1h"
`
	results := runTestOrganize(t, dir, content)
	if filepath.Base(results[0].Source) != "old.txt" {
		t.Errorf("expected old.txt to be organized, got %s", filepath.Base(results[0].Source))
	}
}

func TestOrganizeRecursive(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "source")
	destDir := filepath.Join(dir, "dest")

	subDir := filepath.Join(srcDir, "subdir1", "subdir2")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	srcFile := filepath.Join(subDir, "test.pdf")
	if err := os.WriteFile(srcFile, []byte("pdfcontent"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := makeConfig("rename", config.Rule{
		Name:        "Recursive PDF",
		Extensions:  []string{".pdf"},
		Destination: destDir,
		Action:      "move",
	})

	results, err := Organize(srcDir, cfg, false, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	expectedDest := filepath.Join(destDir, "test.pdf")
	if results[0].Destination != expectedDest {
		t.Errorf("expected destination %q, got %q", expectedDest, results[0].Destination)
	}

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Error("expected moved file to exist at destination")
	}

	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("expected empty subdirectories to be deleted")
	}
}

func TestMatchRuleFilenameRegex(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "matched")
	_ = os.MkdirAll(dest, 0o755)

	for _, f := range []string{"invoice_2024.pdf", "report_final.txt"} {
		_ = os.WriteFile(filepath.Join(dir, f), []byte("data"), 0o644)
	}

	content := `
[[rule]]
name           = "Invoice regex"
filename_regex = ["^invoice_\\d{4}\\.pdf$"]
destination    = '` + dest + `'
action         = "move"
`
	results := runTestOrganize(t, dir, content)
	if filepath.Base(results[0].Source) != "invoice_2024.pdf" {
		t.Errorf("expected invoice_2024.pdf, got %s", filepath.Base(results[0].Source))
	}
}

func TestMatchRuleFilenameRegexNoMatch(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "matched")
	_ = os.MkdirAll(dest, 0o755)

	_ = os.WriteFile(filepath.Join(dir, "report_final.txt"), []byte("data"), 0o644)

	content := `
[[rule]]
name           = "Invoice regex"
filename_regex = ["^invoice_\\d{4}\\.pdf$"]
destination    = '` + dest + `'
action         = "move"
`
	path := filepath.Join(t.TempDir(), "rules.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgParsed, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	results, err := Organize(dir, cfgParsed, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching regex, got %d", len(results))
	}
}

func TestMatchRuleLocations(t *testing.T) {
	rules := []config.Rule{
		{Locations: []string{"/home/user/Downloads"}, Extensions: []string{".pdf"}, Destination: "/docs"},
		{Locations: []string{"/home/user/Desktop"}, Extensions: []string{".jpg"}, Destination: "/pics"},
	}
	// Rule 1 should match for Downloads folder
	if r := MatchRule("file.pdf", nil, rules, "/home/user/Downloads"); r == nil || r.Destination != "/docs" {
		t.Error("expected PDF rule match for Downloads")
	}
	// Rule 2 should match for Desktop folder
	if r := MatchRule("file.jpg", nil, rules, "/home/user/Desktop"); r == nil || r.Destination != "/pics" {
		t.Error("expected JPG rule match for Desktop")
	}
	// Rule 1 should NOT match for Desktop folder
	if r := MatchRule("file.pdf", nil, rules, "/home/user/Desktop"); r != nil {
		t.Error("expected no match for PDF in Desktop folder")
	}
	// Rule 2 should NOT match for Downloads folder
	if r := MatchRule("file.jpg", nil, rules, "/home/user/Downloads"); r != nil {
		t.Error("expected no match for JPG in Downloads folder")
	}
	// Rule without locations should match anywhere
	rulesNoLoc := []config.Rule{
		{Extensions: []string{".txt"}, Destination: "/text"},
	}
	if r := MatchRule("file.txt", nil, rulesNoLoc, "/any/folder"); r == nil || r.Destination != "/text" {
		t.Error("expected txt rule match without locations")
	}
}

func TestDeleteEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := deleteEmptyDirs(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "a")); !os.IsNotExist(err) {
		t.Error("expected empty folder 'a' to be deleted")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected root folder to not be deleted")
	}
}

func setupDuplicateTestDirs(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	destDir := filepath.Join(dir, "dest")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return srcDir, destDir
}

func TestOrganizeDuplicateDetection(t *testing.T) {
	srcDir, destDir := setupDuplicateTestDirs(t)

	content := []byte("identical-content")
	src := filepath.Join(srcDir, "a.pdf")
	dest := filepath.Join(destDir, "a.pdf")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := makeConfig("delete_duplicate", config.Rule{
		Name:        "Move identical",
		Extensions:  []string{".pdf"},
		Destination: destDir,
		Action:      "move",
	})

	results, err := Organize(srcDir, cfg, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("expected duplicate source file to be deleted")
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Error("expected destination file to remain")
	}
}

func TestOrganizeDuplicateDetectionDifferentContent(t *testing.T) {
	srcDir, destDir := setupDuplicateTestDirs(t)

	src := filepath.Join(srcDir, "a.pdf")
	dest := filepath.Join(destDir, "a.pdf")
	if err := os.WriteFile(src, []byte("content-a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("content-b"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := makeConfig("delete_duplicate", config.Rule{
		Name:        "Move different",
		Extensions:  []string{".pdf"},
		Destination: destDir,
		Action:      "move",
	})

	results, err := Organize(srcDir, cfg, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	expectedDest := filepath.Join(destDir, "a_1.pdf")
	if results[0].Destination != expectedDest {
		t.Errorf("expected destination %q, got %q", expectedDest, results[0].Destination)
	}

	if _, err := os.Stat(expectedDest); os.IsNotExist(err) {
		t.Error("expected source file to be renamed and moved")
	}
}


