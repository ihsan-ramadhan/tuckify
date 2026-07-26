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

func cliBinaryName() string {
	name := "tuckify"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func resolveBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return cliBinaryName()
	}

	if !strings.Contains(strings.ToLower(filepath.Base(exe)), "-gui") {
		return exe
	}

	dir := filepath.Dir(exe)
	cliName := cliBinaryName()

	findFile := func(base string) (string, bool) {
		p := filepath.Join(base, cliName)
		fi, err := os.Stat(p)
		if err == nil && !fi.IsDir() {
			return p, true
		}
		if runtime.GOOS == "windows" {
			pNoExt := filepath.Join(base, "tuckify")
			fi, err := os.Stat(pNoExt)
			if err == nil && !fi.IsDir() {
				return pNoExt, true
			}
		}
		return "", false
	}

	if p, ok := findFile(dir); ok {
		return p
	}

	for i := 0; i < 3; i++ {
		dir = filepath.Dir(dir)
		if p, ok := findFile(dir); ok {
			return p
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if p, ok := findFile(filepath.Join(wd, "build", "bin")); ok {
			return p
		}
		if p, ok := findFile(wd); ok {
			return p
		}
	}

	if path, err := exec.LookPath(cliName); err == nil {
		return path
	}

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
