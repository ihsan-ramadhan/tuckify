package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	orig := historyPath
	historyPath = func() string { return filepath.Join(dir, "last_run.json") }
	defer func() { historyPath = orig }()

	entries := []Entry{
		{Src: "/a/file.pdf", Dest: "/b/file.pdf", Action: "move"},
	}
	if err := Save(entries); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Src != "/a/file.pdf" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	historyPath = func() string { return filepath.Join(dir, "nonexistent.json") }
	got, err := Load()
	if err != nil || got != nil {
		t.Fatalf("expected nil, nil; got %v, %v", got, err)
	}
}

func TestUndoMove(t *testing.T) {
	dir := t.TempDir()
	historyPath = func() string { return filepath.Join(dir, "last_run.json") }

	// create a "moved" file at dest
	destDir := filepath.Join(dir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(destDir, "file.txt")
	if err := os.WriteFile(dest, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "file.txt")
	if err := Save([]Entry{{Src: src, Dest: dest, Action: "move"}}); err != nil {
		t.Fatal(err)
	}

	n, err := Undo()
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
	// history file cleared
	if _, err := os.Stat(historyPath()); !os.IsNotExist(err) {
		t.Fatal("history should be cleared after undo")
	}
}
