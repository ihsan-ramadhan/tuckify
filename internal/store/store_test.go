package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func setupTempStore(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	origProfile := os.Getenv("USERPROFILE")
	// point storePath() to temp dir by overriding HOME and USERPROFILE (for Windows)
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	// pre-create config dir
	_ = os.MkdirAll(filepath.Join(tmp, ".config", "tuckify"), 0o755)
	return func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.Setenv("USERPROFILE", origProfile)
	}
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
	if len(ss) != 1 || !reflect.DeepEqual(ss[0], s) {
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

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"downloads", true},
		{"my-schedule", true},
		{"my_schedule", true},
		{"Backup2026", true},
		{"a", true},
		{"0", true},
		{"a-b_c.D", false},
		{"hello world", false},
		{"x; rm -rf ~", false},
		{"$(whoami)", false},
		{"../etc", false},
		{"", false},
		{"with space", false},
		{"tab\there", false},
		{"newline\nhere", false},
	}

	for _, tt := range tests {
		err := ValidateName(tt.name)
		if tt.valid && err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", tt.name, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", tt.name)
		}
	}
}

func TestUpsertRejectsInvalidName(t *testing.T) {
	defer setupTempStore(t)()

	err := Upsert(Schedule{Name: "x; rm -rf ~", Folder: "/tmp", Cron: "0 9 * * *"})
	if err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}

	ss, _ := Load()
	if len(ss) != 0 {
		t.Errorf("expected 0 schedules, got %d", len(ss))
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
	if found == nil || !reflect.DeepEqual(*found, s) {
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

