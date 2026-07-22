package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("schedule name cannot be empty")
	}
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("invalid schedule name %q: must match %s", name, validNameRegex.String())
	}
	return nil
}

type Schedule struct {
	Name      string   `json:"name"`
	Folder    string   `json:"folder,omitempty"`
	Folders   []string `json:"folders,omitempty"`
	Cron      string   `json:"cron"`
	Config    string   `json:"config,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
	Yes       bool     `json:"yes,omitempty"`
}

func (s *Schedule) GetFolders() []string {
	if len(s.Folders) > 0 {
		return s.Folders
	}
	if s.Folder != "" {
		return []string{s.Folder}
	}
	return nil
}

func storePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tuckify", "schedules.json")
}

func Load() ([]Schedule, error) {
	data, err := os.ReadFile(storePath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ss []Schedule
	return ss, json.Unmarshal(data, &ss)
}

func save(ss []Schedule) error {
	p := storePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

func Upsert(s Schedule) error {
	if err := ValidateName(s.Name); err != nil {
		return err
	}
	ss, err := Load()
	if err != nil {
		return err
	}
	for i, e := range ss {
		if e.Name == s.Name {
			ss[i] = s
			return save(ss)
		}
	}
	return save(append(ss, s))
}

func Find(name string) (*Schedule, error) {
	ss, err := Load()
	if err != nil {
		return nil, err
	}
	for i := range ss {
		if ss[i].Name == name {
			return &ss[i], nil
		}
	}
	return nil, nil
}

func Delete(name string) (bool, error) {
	ss, err := Load()
	if err != nil {
		return false, err
	}
	var kept []Schedule
	found := false
	for _, e := range ss {
		if e.Name == name {
			found = true
			continue
		}
		kept = append(kept, e)
	}
	if !found {
		return false, nil
	}
	return true, save(kept)
}
