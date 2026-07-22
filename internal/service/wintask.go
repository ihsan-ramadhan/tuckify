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
	regRunKey     = `Software\Microsoft\Windows\CurrentVersion\Run`
)

type WintaskService struct{}

func NewWintaskService() *WintaskService {
	return &WintaskService{}
}

func (w *WintaskService) Install(name string, folders []string, cronExpr, configPath string) error {
	binaryPath := resolveBinaryPath()

	tuckifyCmd := buildWintaskCmd(name, binaryPath, folders, cronExpr, configPath)
	taskName := wintaskPrefix + name

	batPath, err := writeRestartBat(name, tuckifyCmd)
	if err != nil {
		return fmt.Errorf("write restart wrapper: %w", err)
	}

	if err := exec.Command("reg", "add", `HKCU\`+regRunKey, "/v", taskName, "/t", "REG_SZ", "/d", batPath, "/f").Run(); err != nil {
		return fmt.Errorf("add to startup registry: %w", err)
	}

	return nil
}

func (w *WintaskService) Uninstall(name string) error {
	taskName := "tuckify"
	if name != "" {
		taskName = wintaskPrefix + name
	}

	// Remove from HKCU\...\Run
	_ = exec.Command("reg", "delete", `HKCU\`+regRunKey, "/v", taskName, "/f").Run()

	// Also try to clean up old schtasks tasks (backwards compat)
	if winSch, err := exec.LookPath(schtasksCmd); err == nil {
		_ = exec.Command(winSch, "/delete", "/tn", taskName, "/f").Run()
	}

	// Remove .bat wrapper file
	appDataDir, err := os.UserConfigDir()
	if err == nil {
		batPath := filepath.Join(appDataDir, "tuckify", fmt.Sprintf("tuckify-%s.bat", name))
		_ = os.Remove(batPath)
	}

	return nil
}

func (w *WintaskService) Exists(name string) (bool, error) {
	taskName := wintaskPrefix + name
	if err := exec.Command("reg", "query", `HKCU\`+regRunKey, "/v", taskName).Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (w *WintaskService) CheckStatus() (string, error) {
	out, err := exec.Command("reg", "query", `HKCU\`+regRunKey).Output()
	if err != nil {
		return "", fmt.Errorf("query startup registry: %w", err)
	}
	return fmt.Sprintf("Startup entries:\n%s", string(out)), nil
}

func (w *WintaskService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("logs not available on Windows — check the .bat wrapper in %%APPDATA%%\\tuckify")
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

func resolveBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "tuckify" // fallback to PATH
	}

	if !strings.Contains(strings.ToLower(filepath.Base(exe)), "-gui") {
		return exe
	}

	dir := filepath.Dir(exe)

	base := filepath.Base(exe)
	cliName := strings.ReplaceAll(strings.ReplaceAll(base, "-gui", ""), "_gui", "")
	cliPath := filepath.Join(dir, cliName)
	if _, err := os.Stat(cliPath); err == nil {
		return cliPath
	}

	cliPath = filepath.Join(dir, "tuckify.exe")
	if _, err := os.Stat(cliPath); err == nil {
		return cliPath
	}

	if path, err := exec.LookPath("tuckify"); err == nil {
		return path
	}

	return exe
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
	if err := os.WriteFile(batPath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("write bat file: %w", err)
	}
	return batPath, nil
}
