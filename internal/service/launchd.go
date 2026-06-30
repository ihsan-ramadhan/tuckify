package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const launchctlCmd = "launchctl"

type LaunchdService struct{}

func NewLaunchdService() *LaunchdService {
	return &LaunchdService{}
}

func launchdAgentsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents")
}

func launchdPlistPath(name string) string {
	label := "com.ihsan.tuckify"
	if name != "" {
		label += "." + name
	}
	return filepath.Join(launchdAgentsDir(), label+".plist")
}

func launchdLabel(name string) string {
	if name == "" {
		return "com.ihsan.tuckify"
	}
	return "com.ihsan.tuckify." + name
}

func (l *LaunchdService) Install(name, folder, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	plistPath := launchdPlistPath(name)
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents directory: %w", err)
	}

	content := buildLaunchdContent(name, binaryPath, folder, cronExpr, configPath)
	if err := os.WriteFile(plistPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write plist file: %w", err)
	}

	lctl, err := exec.LookPath(launchctlCmd)
	if err != nil {
		return fmt.Errorf("find launchctl: %w", err)
	}

	if err := exec.Command(lctl, "load", plistPath).Run(); err != nil {
		return fmt.Errorf("launchctl load plist: %w", err)
	}

	return nil
}

func (l *LaunchdService) Uninstall(name string) error {
	lctl, err := exec.LookPath(launchctlCmd)
	if err != nil {
		return fmt.Errorf("find launchctl: %w", err)
	}

	if name != "" {
		return l.removeOne(lctl, name)
	}

	dir := launchdAgentsDir()
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read LaunchAgents dir: %w", err)
	}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "com.ihsan.tuckify") && strings.HasSuffix(n, ".plist") {
			label := strings.TrimSuffix(n, ".plist")
			scheduleName := strings.TrimPrefix(label, "com.ihsan.tuckify.")
			if scheduleName == "com.ihsan.tuckify" {
				scheduleName = ""
			}
			_ = l.removeOne(lctl, scheduleName)
		}
	}
	return nil
}

func (l *LaunchdService) removeOne(lctl, name string) error {
	plistPath := launchdPlistPath(name)
	if _, err := os.Stat(plistPath); err == nil {
		_ = exec.Command(lctl, "unload", plistPath).Run()
	}
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist file: %w", err)
	}
	return nil
}

func (l *LaunchdService) Exists(name string) (bool, error) {
	_, err := os.Stat(launchdPlistPath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LaunchdService) CheckStatus() (string, error) {
	return "Check status: launchctl list | grep tuckify", nil
}

func (l *LaunchdService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("logs not available for launchd — check Console.app or ~/Library/Logs")
}

func buildLaunchdContent(name, binaryPath, folder, cronExpr, configPath string) string {
	argsXml := fmt.Sprintf(`        <string>%s</string>
        <string>schedule</string>
        <string>%s</string>
        <string>%s</string>
        <string>--cron</string>
        <string>%s</string>
`, binaryPath, name, folder, cronExpr)

	if configPath != "" {
		argsXml += fmt.Sprintf(`        <string>--config</string>
        <string>%s</string>
`, configPath)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
%s    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`, launchdLabel(name), argsXml)
}
