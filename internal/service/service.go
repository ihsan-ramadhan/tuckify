package service

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Service interface {
	Install(name string, folders []string, cronExpr, configPath string) error
	Uninstall(name string) error // "" = remove all
	Exists(name string) (bool, error)
	CheckStatus() (string, error)
	Logs(name string, follow bool, lines int) error
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
