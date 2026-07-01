package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	systemctlCmd      = "systemctl"
	systemdUserFlag   = "--user"
	systemdPrefix     = "tuckify-"
	systemdSuffix     = ".service"
	systemdDaemonLoad = "daemon-reload"
)

type SystemdService struct{}

func NewSystemdService() *SystemdService {
	return &SystemdService{}
}

func systemdServicePath(name string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", systemdPrefix+name+systemdSuffix)
}

func systemdServiceDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user")
}

func (s *SystemdService) Install(name, folder, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	servicePath := systemdServicePath(name)
	if err := os.MkdirAll(filepath.Dir(servicePath), 0o755); err != nil {
		return fmt.Errorf("create systemd directory: %w", err)
	}

	content := buildSystemdContent(name, binaryPath, folder, cronExpr, configPath)
	if err := os.WriteFile(servicePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write systemd service file: %w", err)
	}

	sysctl, err := exec.LookPath(systemctlCmd)
	if err != nil {
		return fmt.Errorf("find systemctl: %w", err)
	}

	if err := exec.Command(sysctl, systemdUserFlag, systemdDaemonLoad).Run(); err != nil {
		return fmt.Errorf("systemd daemon-reload: %w", err)
	}

	unitName := systemdPrefix + name
	if err := exec.Command(sysctl, systemdUserFlag, "enable", "--now", unitName).Run(); err != nil {
		return fmt.Errorf("enable and start %s service: %w", unitName, err)
	}

	return nil
}

func (s *SystemdService) Uninstall(name string) error {
	sysctl, err := exec.LookPath(systemctlCmd)
	if err != nil {
		return fmt.Errorf("find systemctl: %w", err)
	}

	if name != "" {
		return s.removeOne(sysctl, name)
	}

	dir := systemdServiceDir()
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read systemd dir: %w", err)
	}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, systemdPrefix) && strings.HasSuffix(n, systemdSuffix) {
			unitName := strings.TrimSuffix(n, systemdSuffix)
			_ = s.removeOne(sysctl, strings.TrimPrefix(unitName, systemdPrefix))
		}
	}

	_ = exec.Command(sysctl, systemdUserFlag, systemdDaemonLoad).Run()
	return nil
}

func (s *SystemdService) removeOne(sysctl, name string) error {
	servicePath := systemdServicePath(name)
	unitName := systemdPrefix + name

	if _, err := os.Stat(servicePath); err == nil {
		_ = exec.Command(sysctl, systemdUserFlag, "disable", "--now", unitName).Run()
	}

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove service file: %w", err)
	}

	_ = exec.Command(sysctl, systemdUserFlag, systemdDaemonLoad).Run()
	return nil
}

func (s *SystemdService) Exists(name string) (bool, error) {
	_, err := os.Stat(systemdServicePath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *SystemdService) CheckStatus() (string, error) {
	return "Check status: systemctl --user status tuckify-<name>", nil
}

func (s *SystemdService) Logs(name string, follow bool, lines int) error {
	args := []string{"--user", "-u", systemdPrefix + name, "-n", fmt.Sprintf("%d", lines), "-o", "short-monotonic"}
	if follow {
		args = append(args, "-f")
	}

	jctl, err := exec.LookPath("journalctl")
	if err != nil {
		return fmt.Errorf("journalctl not found: %w", err)
	}

	fmt.Printf("\033[1;34m==> logs: %s%s\033[0m\n", systemdPrefix, name)
	cmd := exec.Command(jctl, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildSystemdContent(name, binaryPath, folder, cronExpr, configPath string) string {
	execStart := fmt.Sprintf("%s schedule %s %s --cron %q", binaryPath, name, folder, cronExpr)
	if configPath != "" {
		execStart += fmt.Sprintf(" --config %s", configPath)
	}

	return fmt.Sprintf(`[Unit]
Description=tuckify file organizer (%s)
After=default.target

[Service]
ExecStart=%s
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, name, execStart)
}
