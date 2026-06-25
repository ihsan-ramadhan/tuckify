package service

import (
	"fmt"
	"os"
	"os/exec"
)

const schtasksCmd = "schtasks"

type WintaskService struct{}

func NewWintaskService() *WintaskService {
	return &WintaskService{}
}

func (w *WintaskService) Install(folder, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	execCmd := buildWintaskCmd(binaryPath, folder, cronExpr, configPath)

	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return fmt.Errorf("find schtasks: %w", err)
	}

	cmd := exec.Command(winSch, "/create", "/tn", "tuckify", "/tr", execCmd, "/sc", "onlogon", "/f")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("create scheduled task: %w", err)
	}

	return nil
}

func (w *WintaskService) Uninstall() error {
	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return fmt.Errorf("find schtasks: %w", err)
	}

	cmd := exec.Command(winSch, "/delete", "/tn", "tuckify", "/f")
	_ = cmd.Run()
	return nil
}

func (w *WintaskService) Exists() (bool, error) {
	winSch, err := exec.LookPath(schtasksCmd)
	if err != nil {
		return false, fmt.Errorf("find schtasks: %w", err)
	}

	cmd := exec.Command(winSch, "/query", "/tn", "tuckify")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (w *WintaskService) CheckStatus() (string, error) {
	return `To check status, run in cmd: schtasks /query /tn "tuckify"`, nil
}

func buildWintaskCmd(binaryPath, folder, cronExpr, configPath string) string {
	execCmd := fmt.Sprintf(`"%s" schedule "%s" --cron "%s"`, binaryPath, folder, cronExpr)
	if configPath != "" {
		execCmd += fmt.Sprintf(` --config "%s"`, configPath)
	}
	return execCmd
}
