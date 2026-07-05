//go:build desktop

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ihsan-ramadhan/tuckify/internal/config"
	"github.com/ihsan-ramadhan/tuckify/internal/history"
	"github.com/ihsan-ramadhan/tuckify/internal/organizer"
	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context // NOSONAR: wails lifecycle requires ctx field
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectDirectory(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

type scheduleView struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Service   bool      `json:"service"`
	Cron      string    `json:"cron"`
	Folders   []string  `json:"folders"`
	Config    string    `json:"config"`
	LastRun   time.Time `json:"last_run"`
	LastFiles int       `json:"last_files"`
}

func (a *App) GetRulesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tuckify", "rules.toml")
}

func (a *App) GetRules() (string, error) {
	p := a.GetRulesPath()
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(b), err
}

func (a *App) SaveRules(content string) error {
	p := a.GetRulesPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0644)
}

func (a *App) ValidateRules(content string) (string, error) {
	// write to secure temp file to validate
	tmpFile, err := os.CreateTemp("", "tuckify-validate-*.toml")
	if err != nil {
		return "", err
	}
	tmp := tmpFile.Name()
	defer func() { _ = os.Remove(tmp) }()

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	_ = tmpFile.Close()

	_, err = config.Load(tmp)
	if err != nil {
		return err.Error(), nil
	}
	return "", nil
}

func runMatchesFolders(r history.Run, folders []string) bool {
	for _, sf := range folders {
		for _, rf := range r.Folders {
			if sf == rf {
				return true
			}
		}
	}
	return false
}

func countMovedEntries(entries []history.Entry) int {
	count := 0
	for _, e := range entries {
		if e.Action == "move" || e.Action == "" {
			count++
		}
	}
	return count
}

func findLastRunInfo(runs []history.Run, folders []string) (time.Time, int) {
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		if runMatchesFolders(r, folders) {
			return r.Timestamp, countMovedEntries(r.Entries)
		}
	}
	return time.Time{}, 0
}

func (a *App) GetSchedules() ([]scheduleView, error) {
	schedules, err := store.Load()
	if err != nil {
		return nil, err
	}
	srv, err := service.NewService()
	if err != nil {
		return nil, err
	}
	runs, _ := history.LoadAll()

	views := make([]scheduleView, 0, len(schedules))
	for _, s := range schedules {
		online, _ := srv.Exists(s.Name)
		status := "inactive"
		if online {
			status = "active"
		}
		folders := s.GetFolders()
		lastRun, lastFiles := findLastRunInfo(runs, folders)

		views = append(views, scheduleView{
			Name:      s.Name,
			Status:    status,
			Service:   online,
			Cron:      s.Cron,
			Folders:   folders,
			Config:    s.Config,
			LastRun:   lastRun,
			LastFiles: lastFiles,
		})
	}
	return views, nil
}

func (a *App) SaveSchedule(name string, folders []string, cronExpr string, configPath string) error {
	folderStr := ""
	if len(folders) > 0 {
		folderStr = folders[0]
		for i := 1; i < len(folders); i++ {
			folderStr += "," + folders[i]
		}
	}
	return store.Upsert(store.Schedule{
		Name:   name,
		Folder: folderStr,
		Cron:   cronExpr,
		Config: configPath,
	})
}

func (a *App) StartSchedule(name string) error {
	schedules, err := store.Load()
	if err != nil {
		return err
	}
	var target *store.Schedule
	for i := range schedules {
		if schedules[i].Name == name {
			target = &schedules[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("schedule %q not found", name)
	}

	srv, err := service.NewService()
	if err != nil {
		return err
	}
	return srv.Install(target.Name, target.GetFolders(), target.Cron, target.Config)
}

func (a *App) StopSchedule(name string) error {
	srv, err := service.NewService()
	if err != nil {
		return err
	}
	return srv.Uninstall(name)
}

func (a *App) DeleteSchedule(name string) error {
	srv, err := service.NewService()
	if err == nil {
		_ = srv.Uninstall(name)
	}
	_, err = store.Delete(name)
	return err
}

type runResult struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Skipped     bool   `json:"skipped"`
	SkipReason  string `json:"skip_reason"`
	Action      string `json:"action"`
}

func (a *App) RunOrganize(folders []string, dryRun bool) ([]runResult, error) {
	p := a.GetRulesPath()
	cfg, err := config.Load(p)
	if err != nil {
		return nil, err
	}

	var allResults []runResult
	var histEntries []history.Entry
	for _, folder := range folders {
		res, err := organizer.Organize(folder, cfg, dryRun, false)
		if err != nil {
			return nil, err
		}
		for _, r := range res {
			allResults = append(allResults, runResult{
				Source:      r.Source,
				Destination: r.Destination,
				Skipped:     r.Skipped,
				SkipReason:  r.SkipReason,
				Action:      r.Action,
			})
			if !dryRun && !r.Skipped && (r.Action == "" || r.Action == "move") {
				histEntries = append(histEntries, history.Entry{
					Src:    r.Source,
					Dest:   r.Destination,
					Action: "move",
				})
			}
		}
	}

	if !dryRun && len(histEntries) > 0 {
		_ = history.Save(folders, histEntries)
	}

	return allResults, nil
}

func (a *App) GetHistory() ([]history.Run, error) {
	return history.LoadAll()
}

func (a *App) UndoRun(id int) (int, error) {
	return history.Undo(id)
}
