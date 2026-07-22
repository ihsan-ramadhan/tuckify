package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Src    string `json:"src"`
	Dest   string `json:"dest"`
	Action string `json:"action"` // "move", "copy", "delete"
}

type Run struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Folders   []string  `json:"folders"`
	Entries   []Entry   `json:"entries"`
}

// historyDir returns the path to the history directory. It can be overridden in tests.
var historyDir = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tuckify", "history")
}

// Save stores the run history. It limits the total stored runs to 10.
func Save(folders []string, entries []Entry) error {
	dir := historyDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	runs, err := LoadAll()
	if err != nil {
		return err
	}

	nextID := 1
	if len(runs) > 0 {
		// Find highest ID to assign stable sequential ID
		highest := 0
		for _, r := range runs {
			if r.ID > highest {
				highest = r.ID
			}
		}
		nextID = highest + 1
	}

	newRun := Run{
		ID:        nextID,
		Timestamp: time.Now().Truncate(time.Millisecond),
		Folders:   folders,
		Entries:   entries,
	}

	// Write new run
	newFile := filepath.Join(dir, fmt.Sprintf("run_%d.json", nextID))
	data, err := json.MarshalIndent(newRun, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(newFile, data, 0o600); err != nil {
		return err
	}

	// reload and enforce limit of 10 runs
	runs, err = LoadAll()
	if err != nil {
		return nil // don't fail the save if cleanup fails
	}

	if len(runs) > 10 {
		// LoadAll returns runs sorted by timestamp ascending (oldest first).
		// Delete oldest runs until we have 10.
		toDelete := len(runs) - 10
		for i := 0; i < toDelete; i++ {
			p := filepath.Join(dir, fmt.Sprintf("run_%d.json", runs[i].ID))
			_ = os.Remove(p)
		}
	}

	return nil
}

// LoadAll loads all saved runs sorted by timestamp ascending (oldest first).
func LoadAll() ([]Run, error) {
	dir := historyDir()
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var runs []Run
	for _, f := range files {
		if f.IsDir() || !strings.HasPrefix(f.Name(), "run_") || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		p := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var r Run
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		// If ID is not set in filename but we read it, verify ID matches filename to be safe
		idStr := strings.TrimSuffix(strings.TrimPrefix(f.Name(), "run_"), ".json")
		if id, err := strconv.Atoi(idStr); err == nil && r.ID == 0 {
			r.ID = id
		}
		runs = append(runs, r)
	}

	// Sort oldest to newest
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Timestamp.Before(runs[j].Timestamp)
	})

	return runs, nil
}

// findRun locates a run by ID. If id is 0, returns the latest run.
func findRun(runs []Run, id int) *Run {
	if id == 0 && len(runs) > 0 {
		return &runs[len(runs)-1]
	}
	for i := range runs {
		if runs[i].ID == id {
			return &runs[i]
		}
	}
	return nil
}

// revertEntry attempts to reverse a single move entry. Returns true if reverted.
func revertEntry(e Entry) bool {
	if err := os.MkdirAll(filepath.Dir(e.Src), 0o755); err != nil {
		fmt.Printf("skipped %q: %v\n", e.Dest, err)
		return false
	}
	if err := os.Rename(e.Dest, e.Src); err != nil {
		fmt.Printf("skipped %q: %v\n", e.Dest, err)
		return false
	}
	return true
}

func Delete(id int) error {
	p := filepath.Join(historyDir(), fmt.Sprintf("run_%d.json", id))
	return os.Remove(p)
}

func ClearAll() error {
	runs, err := LoadAll()
	if err != nil {
		return err
	}
	for _, r := range runs {
		p := filepath.Join(historyDir(), fmt.Sprintf("run_%d.json", r.ID))
		_ = os.Remove(p)
	}
	return nil
}

// Undo reverses a specific run by ID. If ID is 0, reverses the latest run.
// Returns count of reverted files.
func Undo(id int) (int, error) {
	runs, err := LoadAll()
	if err != nil {
		return 0, err
	}
	if len(runs) == 0 {
		return 0, nil
	}

	target := findRun(runs, id)
	if target == nil {
		return 0, fmt.Errorf("run with ID %d not found", id)
	}

	count := 0
	for _, e := range target.Entries {
		switch e.Action {
		case "copy":
			// ponytail: skip undo for copy — source still exists, dest removal is lossy
			continue
		case "delete":
			// can't undo delete — file is gone
			fmt.Printf("skipped: cannot undo delete of %q\n", e.Src)
			continue
		default: // "move" or ""
			if revertEntry(e) {
				count++
			}
		}
	}

	// Remove the run history file
	p := filepath.Join(historyDir(), fmt.Sprintf("run_%d.json", target.ID))
	_ = os.Remove(p)

	return count, nil
}
