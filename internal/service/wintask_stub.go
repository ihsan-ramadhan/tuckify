//go:build !windows

package service

import (
	"fmt"
)

type WintaskService struct{}

func NewWintaskService() *WintaskService {
	return &WintaskService{}
}

func (w *WintaskService) Install(name string, folders []string, cronExpr, configPath string) error {
	return fmt.Errorf("wintask service only supported on Windows")
}

func (w *WintaskService) Uninstall(name string) error {
	return fmt.Errorf("wintask service only supported on Windows")
}

func (w *WintaskService) Exists(name string) (bool, error) {
	return false, nil
}

func (w *WintaskService) CheckStatus() (string, error) {
	return "", fmt.Errorf("wintask service only supported on Windows")
}

func (w *WintaskService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("wintask service only supported on Windows")
}
