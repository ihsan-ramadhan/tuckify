//go:build windows
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	schtasksCmd   = "schtasks"
	wintaskPrefix = "tuckify-"
	regRunKey     = `Software\Microsoft\Windows\CurrentVersion\Run`
	cmdExe        = "cmd.exe"
)

func safeSystemCmd(binary string) string {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	switch binary {
	case "cmd", cmdExe:
		return filepath.Join(sysRoot, "System32", cmdExe)
	case "powershell", "powershell.exe":
		return filepath.Join(sysRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	case "reg", "reg.exe":
		return filepath.Join(sysRoot, "System32", "reg.exe")
	case "schtasks", "schtasks.exe":
		return filepath.Join(sysRoot, "System32", "schtasks.exe")
	}
	return binary
}

type WintaskService struct{}

func NewWintaskService() *WintaskService {
	return &WintaskService{}
}

func (w *WintaskService) Install(name string, folders []string, cronExpr, configPath string) error {
	_ = w.Uninstall(name)

	binaryPath := resolveBinaryPath()
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found at %s: %w", binaryPath, err)
	}

	tuckifyCmd := buildWintaskCmd(name, binaryPath, folders, cronExpr, configPath)
	taskName := wintaskPrefix + name

	batPath, _, err := writeRestartBat(name, tuckifyCmd)
	if err != nil {
		return fmt.Errorf("write restart wrapper: %w", err)
	}

	if err := cmd("reg", "add", `HKCU\`+regRunKey, "/v", taskName, "/t", "REG_SZ", "/d", fmt.Sprintf(`"%s"`, batPath), "/f").Run(); err != nil {
		return fmt.Errorf("add to startup registry: %w", err)
	}

	args := []string{"schedule", name}
	for _, f := range folders {
		args = append(args, f)
	}
	args = append(args, "--cron", cronExpr, "--run", "--force")
	if configPath != "" {
		args = append(args, "--config", configPath)
	}

	c := exec.Command(binaryPath, args...)
	c.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("start background service process: %w", err)
	}

	return nil
}

func (w *WintaskService) uninstallSingle(name string) error {
	taskName := wintaskPrefix + name

	psCmd := exec.Command(safeSystemCmd("powershell"), "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf("Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -like '*"+wintaskPrefix+"%s.bat*' -or $_.CommandLine -like '*schedule*%s*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }", name, name))
	psCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	_ = psCmd.Run()

	_ = cmd("reg", "delete", `HKCU\`+regRunKey, "/v", taskName, "/f").Run()

	_ = cmd(safeSystemCmd(schtasksCmd), "/delete", "/tn", taskName, "/f").Run()

	appDataDir, err := os.UserConfigDir()
	if err == nil {
		batPath := filepath.Join(appDataDir, "tuckify", fmt.Sprintf(wintaskPrefix+"%s.bat", name))
		_ = os.Remove(batPath)
	}
	return nil
}

func (w *WintaskService) uninstallAll() error {
	psCmd := exec.Command(safeSystemCmd("powershell"), "-NoProfile", "-NonInteractive", "-Command",
		"Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -like '*"+wintaskPrefix+"*.bat*' -or $_.CommandLine -like '*schedule*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }")
	psCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	_ = psCmd.Run()

	appDataDir, err := os.UserConfigDir()
	if err == nil {
		entries, _ := os.ReadDir(filepath.Join(appDataDir, "tuckify"))
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), wintaskPrefix) && strings.HasSuffix(e.Name(), ".bat") {
				schedName := strings.TrimSuffix(strings.TrimPrefix(e.Name(), wintaskPrefix), ".bat")
				_ = cmd("reg", "delete", `HKCU\`+regRunKey, "/v", wintaskPrefix+schedName, "/f").Run()
				_ = os.Remove(filepath.Join(appDataDir, "tuckify", e.Name()))
			}
		}
	}
	return nil
}

func (w *WintaskService) Uninstall(name string) error {
	if name != "" {
		return w.uninstallSingle(name)
	}
	return w.uninstallAll()
}

func (w *WintaskService) Exists(name string) (bool, error) {
	taskName := wintaskPrefix + name
	if err := cmd("reg", "query", `HKCU\`+regRunKey, "/v", taskName).Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (w *WintaskService) CheckStatus() (string, error) {
	out, err := cmd("reg", "query", `HKCU\`+regRunKey).Output()
	if err != nil {
		return "", fmt.Errorf("query startup registry: %w", err)
	}
	return fmt.Sprintf("Startup entries:\n%s", string(out)), nil
}

func (w *WintaskService) Logs(name string, follow bool, lines int) error {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("get user config dir: %w", err)
	}
	logPath := filepath.Join(appDataDir, "tuckify", fmt.Sprintf("tuckify-%s.log", name))
	data, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("read log file: %w", err)
	}

	allLines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(allLines) > lines {
		allLines = allLines[len(allLines)-lines:]
	}
	fmt.Println(strings.Join(allLines, "\n"))
	return nil
}

func buildWintaskCmd(name, binaryPath string, folders []string, cronExpr, configPath string) string {
	parts := []string{
		fmt.Sprintf(`"%s"`, binaryPath),
		"schedule", fmt.Sprintf(`"%s"`, name),
	}
	for _, f := range folders {
		parts = append(parts, fmt.Sprintf(`"%s"`, f))
	}
	parts = append(parts, "--cron", fmt.Sprintf(`"%s"`, cronExpr), "--run", "--force")
	if configPath != "" {
		parts = append(parts, "--config", fmt.Sprintf(`"%s"`, configPath))
	}
	return strings.Join(parts, " ")
}

// cmd wraps exec.Command with HideWindow to prevent console window flashes.
func cmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(safeSystemCmd(name), args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return c
}

func writeRestartBat(name, tuckifyCmd string) (string, string, error) {
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("get user config dir: %w", err)
	}
	batDir := filepath.Join(appDataDir, "tuckify")
	if err := os.MkdirAll(batDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create bat dir: %w", err)
	}
	batPath := filepath.Join(batDir, fmt.Sprintf(wintaskPrefix+"%s.bat", name))
	logPath := filepath.Join(batDir, fmt.Sprintf(wintaskPrefix+"%s.log", name))

	content := fmt.Sprintf("@echo off\r\n:loop\r\n%s >> \"%s\" 2>&1\r\nif %%ERRORLEVEL%% NEQ 0 (\r\n    timeout /t 5 /nobreak >nul\r\n    goto loop\r\n)", tuckifyCmd, logPath)
	if err := os.WriteFile(batPath, []byte(content), 0o644); err != nil {
		return "", "", fmt.Errorf("write bat file: %w", err)
	}
	return batPath, logPath, nil
}
