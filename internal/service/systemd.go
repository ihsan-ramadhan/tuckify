package service

import (
	"fmt"
	"os"
	"path/filepath"
)

type SystemdService struct{}

func NewSystemdService() *SystemdService {
	return &SystemdService{}
}

func (s *SystemdService) Install(folder string, cronExpr string, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	servicePath := filepath.Join(home, ".config/systemd/user/tuckify.service")
	serviceDir := filepath.Dir(servicePath)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return fmt.Errorf("create systemd directory: %w", err)
	}

	execStart := fmt.Sprintf("%s schedule %s --cron %q", binaryPath, folder, cronExpr)
	if configPath != "" {
		execStart += fmt.Sprintf(" --config %s", configPath)
	}

	content := fmt.Sprintf(`[Unit]
Description=tuckify file organizer
After=default.target

[Service]
ExecStart=%s
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, execStart)

	if err := os.WriteFile(servicePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write systemd service file: %w", err)
	}

	return nil
}

func (s *SystemdService) Uninstall() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	servicePath := filepath.Join(home, ".config/systemd/user/tuckify.service")
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove service file: %w", err)
	}

	return nil
}

func (s *SystemdService) Exists() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("get home directory: %w", err)
	}
	servicePath := filepath.Join(home, ".config/systemd/user/tuckify.service")
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
	return "", nil
}
