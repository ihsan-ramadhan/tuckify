//go:build desktop

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
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

type RuleView struct {
	Extensions       []string `json:"extensions"`
	FilenamePatterns []string `json:"filename_patterns"`
	FilenameRegex    []string `json:"filename_regex"`
	Destination      string   `json:"destination"`
	Action           string   `json:"action"`
	MinSize          string   `json:"min_size"`
	MaxSize          string   `json:"max_size"`
	MinAge           string   `json:"min_age"`
	MaxAge           string   `json:"max_age"`
}

func (a *App) GetVisualRules() ([]RuleView, error) {
	p := a.GetRulesPath()
	cfg, err := config.Load(p)
	if os.IsNotExist(err) {
		return []RuleView{}, nil
	}
	if err != nil {
		return nil, err
	}

	views := make([]RuleView, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		action := r.Action
		if action == "" {
			action = "move"
		}
		views = append(views, RuleView{
			Extensions:       r.Extensions,
			FilenamePatterns: r.FilenamePatterns,
			FilenameRegex:    r.FilenameRegex,
			Destination:      r.Destination,
			Action:           action,
			MinSize:          r.MinSize,
			MaxSize:          r.MaxSize,
			MinAge:           r.MinAge,
			MaxAge:           r.MaxAge,
		})
	}
	return views, nil
}

func encodeRulesToConfig(rules []RuleView) (config.Config, error) {
	cfg := config.Config{Settings: config.Settings{ConflictStrategy: "rename"}}
	for _, r := range rules {
		cfg.Rules = append(cfg.Rules, config.Rule{
			Extensions:       r.Extensions,
			FilenamePatterns: r.FilenamePatterns,
			FilenameRegex:    r.FilenameRegex,
			Destination:      r.Destination,
			Action:           r.Action,
			MinSize:          r.MinSize,
			MaxSize:          r.MaxSize,
			MinAge:           r.MinAge,
			MaxAge:           r.MaxAge,
		})
	}
	return cfg, nil
}

func (a *App) SaveVisualRules(rules []RuleView) error {
	p := a.GetRulesPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}

	existing, err := config.Load(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if existing == nil {
		existing = &config.Config{}
	}

	cfg, err := encodeRulesToConfig(rules)
	if err != nil {
		return err
	}
	if existing.Settings.ConflictStrategy != "" {
		cfg.Settings.ConflictStrategy = existing.Settings.ConflictStrategy
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(p, buf.Bytes(), 0644)
}

func (a *App) ValidateVisualRules(rules []RuleView) (string, error) {
	cfg, err := encodeRulesToConfig(rules)
	if err != nil {
		return err.Error(), nil
	}

	tmpFile, err := os.CreateTemp("", "tuckify-validate-*.toml")
	if err != nil {
		return "", err
	}
	tmp := tmpFile.Name()
	defer func() { _ = os.Remove(tmp) }()

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cfg); err != nil {
		_ = tmpFile.Close()
		return err.Error(), nil
	}
	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	_ = tmpFile.Close()

	if _, err := config.Load(tmp); err != nil {
		return err.Error(), nil
	}
	return "", nil
}

type scheduleView struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	Service   bool       `json:"service"`
	Cron      string     `json:"cron"`
	Folders   []string   `json:"folders"`
	Config    string     `json:"config"`
	Recursive bool       `json:"recursive"`
	Yes       bool       `json:"yes"`
	LastRun   *time.Time `json:"last_run"`
	LastFiles int        `json:"last_files"`
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

		var lastRunPtr *time.Time
		if !lastRun.IsZero() {
			lastRunPtr = &lastRun
		}

		views = append(views, scheduleView{
			Name:      s.Name,
			Status:    status,
			Service:   online,
			Cron:      s.Cron,
			Folders:   folders,
			Config:    s.Config,
			Recursive: s.Recursive,
			Yes:       s.Yes,
			LastRun:   lastRunPtr,
			LastFiles: lastFiles,
		})
	}
	return views, nil
}

func (a *App) SaveSchedule(name string, folders []string, cronExpr string, configPath string, recursive, yes bool) error {
	return store.Upsert(store.Schedule{
		Name:      name,
		Folders:   folders,
		Cron:      cronExpr,
		Config:    configPath,
		Recursive: recursive,
		Yes:       yes,
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

func (a *App) RestartSchedule(name string) error {
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

	// Stop first, then start
	_ = srv.Uninstall(name)
	return srv.Install(target.Name, target.GetFolders(), target.Cron, target.Config)
}

func (a *App) StartupAll() error {
	schedules, err := store.Load()
	if err != nil {
		return err
	}
	if len(schedules) == 0 {
		return nil
	}

	srv, err := service.NewService()
	if err != nil {
		return err
	}

	var lastErr error
	for _, s := range schedules {
		if err := srv.Install(s.Name, s.GetFolders(), s.Cron, s.Config); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (a *App) UnstartupAll() error {
	srv, err := service.NewService()
	if err != nil {
		return err
	}
	return srv.Uninstall("")
}

type runResult struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Skipped     bool   `json:"skipped"`
	SkipReason  string `json:"skip_reason"`
	Action      string `json:"action"`
}

func (a *App) RunOrganize(folders []string, dryRun, recursive bool) ([]runResult, error) {
	p := a.GetRulesPath()
	cfg, err := config.Load(p)
	if err != nil {
		return nil, err
	}

	var allResults []runResult
	var histEntries []history.Entry
	for _, folder := range folders {
		res, err := organizer.Organize(folder, cfg, dryRun, recursive)
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

func (a *App) DeleteHistoryRun(id int) error {
	return history.Delete(id)
}

func (a *App) ClearHistory() error {
	return history.ClearAll()
}

func (a *App) GetLogs(name string, lines int, follow bool) (string, error) {
	srv, err := service.NewService()
	if err != nil {
		return "", err
	}

	if _, errSys := exec.LookPath("systemctl"); errSys == nil {
		jctl, errJ := exec.LookPath("journalctl")
		if errJ != nil {
			return "", fmt.Errorf("journalctl not found: %w", errJ)
		}
		args := []string{"--user", "-u", "tuckify-" + name, "-n", fmt.Sprintf("%d", lines), "--no-pager", "-o", "short-monotonic"}
		if follow {
			args = append(args, "-f")
		}
		ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, jctl, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		_ = cmd.Run()
		return out.String(), nil
	}

	status, err := srv.CheckStatus()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Logs not directly fetchable on this platform.\n%s\n", status), nil
}

func (a *App) GetConflictStrategy() (string, error) {
	p := a.GetRulesPath()
	var cfg config.Config
	_, err := toml.DecodeFile(p, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return "rename", nil
		}
		return "", err
	}
	strategy := cfg.Settings.ConflictStrategy
	if strategy == "" {
		strategy = "rename"
	}
	return strategy, nil
}

func (a *App) SaveConflictStrategy(strategy string) error {
	p := a.GetRulesPath()
	var cfg config.Config
	_, err := toml.DecodeFile(p, &cfg)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	cfg.Settings.ConflictStrategy = strategy

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(p, buf.Bytes(), 0644)
}
