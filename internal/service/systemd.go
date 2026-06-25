package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	systemctlCmd       = "systemctl"
	systemdUserFlag    = "--user"
	systemdServicePath = ".config/systemd/user/tuckify.service"
)

type SystemdService struct{}

func NewSystemdService() *SystemdService {
	return &SystemdService{}
}

func (s *SystemdService) Install(folder, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	servicePath := filepath.Join(home, systemdServicePath)
	serviceDir := filepath.Dir(servicePath)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return fmt.Errorf("create systemd directory: %w", err)
	}

	content := buildSystemdContent(binaryPath, folder, cronExpr, configPath)

	if err := os.WriteFile(servicePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write systemd service file: %w", err)
	}

	sysctl, err := exec.LookPath(systemctlCmd)
	if err != nil {
		return fmt.Errorf("find systemctl: %w", err)
	}

	cmdReload := exec.Command(sysctl, systemdUserFlag, "daemon-reload")
	if err := cmdReload.Run(); err != nil {
		return fmt.Errorf("systemd daemon-reload: %w", err)
	}

	cmdEnable := exec.Command(sysctl, systemdUserFlag, "enable", "--now", "tuckify")
	if err := cmdEnable.Run(); err != nil {
		return fmt.Errorf("enable and start tuckify service: %w", err)
	}

	return nil
}

func (s *SystemdService) Uninstall() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	servicePath := filepath.Join(home, systemdServicePath)

	sysctl, err := exec.LookPath(systemctlCmd)
	if err != nil {
		return fmt.Errorf("find systemctl: %w", err)
	}

	if _, err := os.Stat(servicePath); err == nil {
		cmdDisable := exec.Command(sysctl, systemdUserFlag, "disable", "--now", "tuckify")
		_ = cmdDisable.Run()
	}

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove service file: %w", err)
	}

	cmdReload := exec.Command(sysctl, systemdUserFlag, "daemon-reload")
	_ = cmdReload.Run()

	return nil
}

func (s *SystemdService) Exists() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("get home directory: %w", err)
	}
	servicePath := filepath.Join(home, systemdServicePath)
	_, err = os.Stat(servicePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *SystemdService) CheckStatus() (string, error) {
	return "To check status, run: systemctl --user status tuckify", nil
}

func buildSystemdContent(binaryPath, folder, cronExpr, configPath string) string {
	execStart := fmt.Sprintf("%s schedule %s --cron %q", binaryPath, folder, cronExpr)
	if configPath != "" {
		execStart += fmt.Sprintf(" --config %s", configPath)
	}

	return fmt.Sprintf(`[Unit]
Description=tuckify file organizer
After=default.target

[Service]
ExecStart=%s
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, execStart)
}
