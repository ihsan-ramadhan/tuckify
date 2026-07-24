package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Service interface {
	Install(name string, folders []string, cronExpr, configPath string) error
	Uninstall(name string) error // "" = remove all
	Exists(name string) (bool, error)
	CheckStatus() (string, error)
	Logs(name string, follow bool, lines int) error
}

// cliBinaryName returns the tuckify CLI binary name for the current platform.
func cliBinaryName() string {
	name := "tuckify"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// resolveBinaryPath resolves the path to the tuckify CLI binary.
// When running as the GUI binary (tuckify-gui), it searches multiple locations
// to find the CLI binary (tuckify), so that systemd/launchd/wintask service
// files use the CLI binary instead of the GUI binary.
//
// Search order:
//  1. Same directory as the current binary (production install)
//  2. Parent directories (dev: build/bin/ → project root)
//  3. PATH
//  4. Fallback to current binary (GUI can also handle CLI in dual mode)
func resolveBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return cliBinaryName()
	}

	// Already a CLI binary, not GUI — use as-is
	if !strings.Contains(strings.ToLower(filepath.Base(exe)), "-gui") {
		return exe
	}

	dir := filepath.Dir(exe)
	cliName := cliBinaryName()

	// findFile checks if path exists and is a regular file (not a directory).
	findFile := func(base string) (string, bool) {
		p := filepath.Join(base, cliName)
		fi, err := os.Stat(p)
		if err == nil && !fi.IsDir() {
			return p, true
		}
		return "", false
	}

	// 1. Same directory (e.g., ~/.local/bin/)
	if p, ok := findFile(dir); ok {
		return p
	}

	// 2. Walk up parent directories to find tuckify
	// Handles dev layout: build/bin/tuckify-gui → find ./tuckify
	for i := 0; i < 3; i++ {
		dir = filepath.Dir(dir)
		if p, ok := findFile(dir); ok {
			return p
		}
	}

	// 3. Search PATH
	if path, err := exec.LookPath(cliName); err == nil {
		return path
	}

	// 4. Fallback: return GUI binary (it can handle CLI commands in dual mode)
	return exe
}

func NewService() (Service, error) {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err == nil {
			return NewSystemdService(), nil
		}
		return NewCrontabService(), nil
	case "darwin":
		return NewLaunchdService(), nil
	case "windows":
		return NewWintaskService(), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
