package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Entry struct {
	Src    string `json:"src"`
	Dest   string `json:"dest"`
	Action string `json:"action"` // "move", "copy", "delete"
}

// historyPath is a var so tests can override it.
var historyPath = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tuckify", "last_run.json")
}

func Save(entries []Entry) error {
	p := historyPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func Load() ([]Entry, error) {
	data, err := os.ReadFile(historyPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	return entries, json.Unmarshal(data, &entries)
}

// Undo reverses all entries from last run. Returns count of reverted files.
func Undo() (int, error) {
	entries, err := Load()
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, nil
	}

	count := 0
	for _, e := range entries {
		switch e.Action {
		case "copy":
			// ponytail: skip undo for copy — source still exists, dest removal is lossy
			continue
		case "delete":
			// can't undo delete — file is gone
			fmt.Printf("skipped: cannot undo delete of %q\n", e.Src)
			continue
		default: // "move" or ""
			if err := os.MkdirAll(filepath.Dir(e.Src), 0o755); err != nil {
				fmt.Printf("skipped %q: %v\n", e.Dest, err)
				continue
			}
			if err := os.Rename(e.Dest, e.Src); err != nil {
				fmt.Printf("skipped %q: %v\n", e.Dest, err)
				continue
			}
			count++
		}
	}

	// clear history after undo
	_ = os.Remove(historyPath())
	return count, nil
}
