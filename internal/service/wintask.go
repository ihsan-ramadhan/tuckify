package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	schtasksCmd   = "schtasks"
	wintaskPrefix = "tuckify-"
)

type WintaskService struct{}

func NewWintaskService() *WintaskService {
	return &WintaskService{}
}

func (w *WintaskService) Install(name string, folders []string, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	tuckifyCmd := buildWintaskCmd(name, binaryPath, folders, cronExpr, configPath)
	taskName := wintaskPrefix + name

	batPath, err := writeRestartBat(name, tuckifyCmd)
	if err != nil {
		return fmt.Errorf("write restart wrapper: %w", err)
	}

	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return fmt.Errorf("find schtasks: %w", err)
	}

	cmd := exec.Command(winSch, "/create", "/tn", taskName, "/tr", batPath, "/sc", "onlogon", "/rl", "highest", "/f")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create scheduled task: %w", err)
	}

	return nil
}

func (w *WintaskService) Uninstall(name string) error {
	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return fmt.Errorf("find schtasks: %w", err)
	}

	taskName := "tuckify"
	if name != "" {
		taskName = wintaskPrefix + name
	}

	_ = exec.Command(winSch, "/delete", "/tn", taskName, "/f").Run()

	// remove .bat wrapper file
	appDataDir, err := os.UserConfigDir()
	if err == nil {
		batPath := filepath.Join(appDataDir, "tuckify", fmt.Sprintf("tuckify-%s.bat", name))
		_ = os.Remove(batPath)
	}

	return nil
}

func (w *WintaskService) Exists(name string) (bool, error) {
	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return false, fmt.Errorf("find schtasks: %w", err)
	}

	taskName := wintaskPrefix + name
	if err := exec.Command(winSch, "/query", "/tn", taskName).Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (w *WintaskService) CheckStatus() (string, error) {
	return `To check status, run in cmd: schtasks /query /tn "tuckify-<name>"`, nil
}

func (w *WintaskService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("logs not available for Windows Task Scheduler — check Event Viewer")
}

func buildWintaskCmd(name, binaryPath string, folders []string, cronExpr, configPath string) string {
	parts := []string{
		fmt.Sprintf(`"%s"`, binaryPath),
		"schedule", fmt.Sprintf(`"%s"`, name),
	}
	for _, f := range folders {
		parts = append(parts, fmt.Sprintf(`"%s"`, f))
	}
	parts = append(parts, "--cron", fmt.Sprintf(`"%s"`, cronExpr), "--run")
	if configPath != "" {
		parts = append(parts, "--config", fmt.Sprintf(`"%s"`, configPath))
	}
	return strings.Join(parts, " ")
}

func writeRestartBat(name, tuckifyCmd string) (string, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	batDir := filepath.Join(appDataDir, "tuckify")
	if err := os.MkdirAll(batDir, 0o755); err != nil {
		return "", fmt.Errorf("create bat dir: %w", err)
	}
	batPath := filepath.Join(batDir, fmt.Sprintf("tuckify-%s.bat", name))
	content := fmt.Sprintf("@echo off\r\n:loop\r\n%s\r\nif %%ERRORLEVEL%% NEQ 0 (\r\n    timeout /t 5 /nobreak >nul\r\n    goto loop\r\n)", tuckifyCmd)
	if err := os.WriteFile(batPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write bat file: %w", err)
	}
	return batPath, nil
}
