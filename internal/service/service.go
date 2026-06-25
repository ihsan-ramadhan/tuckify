package service

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Service interface {
	Install(folder, cronExpr, configPath string) error
	Uninstall() error
	Exists() (bool, error)
	CheckStatus() (string, error)
}

type placeholderService struct {
	os string
}

func (p *placeholderService) Install(folder string, cronExpr string, configPath string) error {
	return fmt.Errorf("not implemented for %s", p.os)
}

func (p *placeholderService) Uninstall() error {
	return fmt.Errorf("not implemented for %s", p.os)
}

func (p *placeholderService) Exists() (bool, error) {
	return false, nil
}

func (p *placeholderService) CheckStatus() (string, error) {
	return "", nil
}

func NewService() (Service, error) {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err == nil {
			return NewSystemdService(), nil
		}
		return NewCrontabService(), nil
	case "darwin", "windows":
		return &placeholderService{os: runtime.GOOS}, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
