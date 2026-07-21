package service

import (
	"fmt"
	"os"
	"os/exec"
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

	execCmd := buildWintaskCmd(name, binaryPath, folders, cronExpr, configPath)
	taskName := wintaskPrefix + name

	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return fmt.Errorf("find schtasks: %w", err)
	}

	cmd := exec.Command(winSch, "/create", "/tn", taskName, "/tr", execCmd, "/sc", "onlogon", "/rl", "highest", "/f")
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
