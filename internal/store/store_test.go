package store

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTempStore(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	orig := os.Getenv("HOME")
	// point storePath() to temp dir by overriding HOME
	t.Setenv("HOME", tmp)
	// pre-create config dir
	_ = os.MkdirAll(filepath.Join(tmp, ".config", "tuckify"), 0o755)
	return func() { _ = os.Setenv("HOME", orig) }
}

func TestLoadEmpty(t *testing.T) {
	defer setupTempStore(t)()
	ss, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 0 {
		t.Errorf("expected empty, got %v", ss)
	}
}

func TestUpsertAndLoad(t *testing.T) {
	defer setupTempStore(t)()

	s := Schedule{Name: "downloads", Folder: "/data", Cron: "0 9 * * *"}
	if err := Upsert(s); err != nil {
		t.Fatal(err)
	}

	ss, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 1 || ss[0] != s {
		t.Errorf("expected [%v], got %v", s, ss)
	}
}

func TestUpsertUpdatesExisting(t *testing.T) {
	defer setupTempStore(t)()

	_ = Upsert(Schedule{Name: "dl", Folder: "/old", Cron: "0 9 * * *"})
	_ = Upsert(Schedule{Name: "dl", Folder: "/new", Cron: "0 18 * * *"})

	ss, _ := Load()
	if len(ss) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(ss))
	}
	if ss[0].Folder != "/new" {
		t.Errorf("expected /new, got %s", ss[0].Folder)
	}
}

func TestDelete(t *testing.T) {
	defer setupTempStore(t)()

	_ = Upsert(Schedule{Name: "a", Folder: "/a", Cron: "0 1 * * *"})
	_ = Upsert(Schedule{Name: "b", Folder: "/b", Cron: "0 2 * * *"})

	found, err := Delete("a")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("expected found=true")
	}

	ss, _ := Load()
	if len(ss) != 1 || ss[0].Name != "b" {
		t.Errorf("expected only 'b' remaining, got %v", ss)
	}

	found2, _ := Delete("nonexistent")
	if found2 {
		t.Error("expected found=false for nonexistent")
	}
}

func TestFind(t *testing.T) {
	defer setupTempStore(t)()

	s := Schedule{Name: "downloads", Folder: "/data", Cron: "0 9 * * *"}
	_ = Upsert(s)

	found, err := Find("downloads")
	if err != nil {
		t.Fatal(err)
	}
	if found == nil || *found != s {
		t.Errorf("expected to find %v, got %v", s, found)
	}

	notFound, err := Find("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if notFound != nil {
		t.Errorf("expected nil for nonexistent, got %v", notFound)
	}
}

