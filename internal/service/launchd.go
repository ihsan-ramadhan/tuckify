package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	launchctlCmd        = "launchctl"
	launchdPlistRelPath = "Library/LaunchAgents/com.ihsan.tuckify.plist"
)

type LaunchdService struct{}

func NewLaunchdService() *LaunchdService {
	return &LaunchdService{}
}

func (l *LaunchdService) Install(folder, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	plistPath := filepath.Join(home, launchdPlistRelPath)
	plistDir := filepath.Dir(plistPath)
	if err := os.MkdirAll(plistDir, 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents directory: %w", err)
	}

	content := buildLaunchdContent(binaryPath, folder, cronExpr, configPath)

	if err := os.WriteFile(plistPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write plist file: %w", err)
	}

	lctl, err := exec.LookPath(launchctlCmd)
	if err != nil {
		return fmt.Errorf("find launchctl: %w", err)
	}

	cmdLoad := exec.Command(lctl, "load", plistPath)
	if err := cmdLoad.Run(); err != nil {
		return fmt.Errorf("launchctl load plist: %w", err)
	}

	return nil
}

func (l *LaunchdService) Uninstall() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	plistPath := filepath.Join(home, launchdPlistRelPath)

	lctl, err := exec.LookPath(launchctlCmd)
	if err != nil {
		return fmt.Errorf("find launchctl: %w", err)
	}

	if _, err := os.Stat(plistPath); err == nil {
		cmdUnload := exec.Command(lctl, "unload", plistPath)
		_ = cmdUnload.Run()
	}

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist file: %w", err)
	}

	return nil
}

func (l *LaunchdService) Exists() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("get home directory: %w", err)
	}
	plistPath := filepath.Join(home, launchdPlistRelPath)
	_, err = os.Stat(plistPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (l *LaunchdService) CheckStatus() (string, error) {
	return "To check status, run: launchctl list | grep tuckify", nil
}

func buildLaunchdContent(binaryPath, folder, cronExpr, configPath string) string {
	argsXml := fmt.Sprintf(`        <string>%s</string>
        <string>schedule</string>
        <string>%s</string>
        <string>--cron</string>
        <string>%s</string>
`, binaryPath, folder, cronExpr)

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
    <string>com.ihsan.tuckify</string>
    <key>ProgramArguments</key>
    <array>
%s    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`, argsXml)
}
