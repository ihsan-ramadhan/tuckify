package service

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Service interface {
	Install(name, folder, cronExpr, configPath string) error
	Uninstall(name string) error // "" = remove all
	Exists(name string) (bool, error)
	CheckStatus() (string, error)
	Logs(name string, follow bool, lines int) error
}

type placeholderService struct {
	os string
}

func (p *placeholderService) Install(name, folder, cronExpr, configPath string) error {
	return fmt.Errorf("not implemented for %s", p.os)
}

func (p *placeholderService) Uninstall(name string) error {
	return fmt.Errorf("not implemented for %s", p.os)
}

func (p *placeholderService) Exists(name string) (bool, error) {
	return false, nil
}

func (p *placeholderService) CheckStatus() (string, error) {
	return "", nil
}

func (p *placeholderService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("logs not implemented for %s", p.os)
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
