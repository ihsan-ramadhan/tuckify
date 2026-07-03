package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadAll(t *testing.T) {
	dir := t.TempDir()
	orig := historyDir
	historyDir = func() string { return dir }
	defer func() { historyDir = orig }()

	// Save first run
	entries1 := []Entry{
		{Src: "/a/file1.pdf", Dest: "/b/file1.pdf", Action: "move"},
	}
	if err := Save([]string{"/a"}, entries1); err != nil {
		t.Fatal(err)
	}

	// Save second run
	time.Sleep(10 * time.Millisecond) // ensure different timestamp
	entries2 := []Entry{
		{Src: "/c/file2.pdf", Dest: "/d/file2.pdf", Action: "move"},
	}
	if err := Save([]string{"/c"}, entries2); err != nil {
		t.Fatal(err)
	}

	runs, err := LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// Should be sorted oldest to newest
	if runs[0].ID != 1 || runs[1].ID != 2 {
		t.Errorf("unexpected IDs: %d, %d", runs[0].ID, runs[1].ID)
	}
	if len(runs[0].Folders) != 1 || runs[0].Folders[0] != "/a" {
		t.Errorf("unexpected folders in run 1: %v", runs[0].Folders)
	}
	if len(runs[1].Entries) != 1 || runs[1].Entries[0].Src != "/c/file2.pdf" {
		t.Errorf("unexpected entries in run 2: %v", runs[1].Entries)
	}
}

func TestLoadAllEmpty(t *testing.T) {
	dir := t.TempDir()
	historyDir = func() string { return filepath.Join(dir, "nonexistent") }
	runs, err := LoadAll()
	if err != nil || runs != nil {
		t.Fatalf("expected nil, nil; got %v, %v", runs, err)
	}
}

func TestUndoLatest(t *testing.T) {
	dir := t.TempDir()
	orig := historyDir
	historyDir = func() string { return dir }
	defer func() { historyDir = orig }()

	// Create a "moved" file at dest
	destDir := filepath.Join(dir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(destDir, "file.txt")
	if err := os.WriteFile(dest, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "src", "file.txt")
	entries := []Entry{{Src: src, Dest: dest, Action: "move"}}
	if err := Save([]string{dir}, entries); err != nil {
		t.Fatal(err)
	}

	// Undo latest (ID 0)
	n, err := Undo(0)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 reverted, got %d", n)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("src not restored: %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("dest should be gone after undo")
	}

	// History file for run 1 should be deleted
	if _, err := os.Stat(filepath.Join(dir, "run_1.json")); !os.IsNotExist(err) {
		t.Fatal("history file should be deleted after undo")
	}
}

func TestUndoByID(t *testing.T) {
	dir := t.TempDir()
	orig := historyDir
	historyDir = func() string { return dir }
	defer func() { historyDir = orig }()

	// Save 2 runs
	if err := Save([]string{"/a"}, []Entry{{Src: "/a/f1", Dest: "/b/f1", Action: "move"}}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := Save([]string{"/c"}, []Entry{{Src: "/c/f2", Dest: "/d/f2", Action: "move"}}); err != nil {
		t.Fatal(err)
	}

	runs, _ := LoadAll()
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// Undo run 1 (oldest) — will fail because files don't exist, but should not error out
	n, err := Undo(1)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Logf("expected 0 reverted (files don't exist), got %d", n)
	}

	// run_1.json should be deleted
	if _, err := os.Stat(filepath.Join(dir, "run_1.json")); !os.IsNotExist(err) {
		t.Fatal("run_1.json should be deleted")
	}

	// run_2.json should still exist
	if _, err := os.Stat(filepath.Join(dir, "run_2.json")); err != nil {
		t.Fatal("run_2.json should still exist")
	}
}

func TestLimitEnforcement(t *testing.T) {
	dir := t.TempDir()
	orig := historyDir
	historyDir = func() string { return dir }
	defer func() { historyDir = orig }()

	// Save 12 runs
	for i := 1; i <= 12; i++ {
		entries := []Entry{{Src: "/a/f", Dest: "/b/f", Action: "move"}}
		if err := Save([]string{"/a"}, entries); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	runs, err := LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	// Should keep only last 10
	if len(runs) != 10 {
		t.Fatalf("expected 10 runs after limit enforcement, got %d", len(runs))
	}

	// Should have IDs 3..12 (oldest 1 and 2 were deleted)
	if runs[0].ID != 3 || runs[len(runs)-1].ID != 12 {
		t.Errorf("unexpected ID range: %d..%d, want 3..12", runs[0].ID, runs[len(runs)-1].ID)
	}
}
