package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CrontabService struct{}

func NewCrontabService() *CrontabService {
	return &CrontabService{}
}

func (c *CrontabService) Install(folder string, cronExpr string, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	currCrontab, err := c.readCrontab()
	if err != nil {
		return fmt.Errorf("read crontab: %w", err)
	}

	newContent := updateCrontabContent(currCrontab, binaryPath, folder, cronExpr, configPath)

	if err := c.writeCrontab(newContent); err != nil {
		return fmt.Errorf("write crontab: %w", err)
	}

	return nil
}

func (c *CrontabService) Uninstall() error {
	currCrontab, err := c.readCrontab()
	if err != nil {
		return fmt.Errorf("read crontab: %w", err)
	}

	newContent, hasManaged := removeCrontabContent(currCrontab)
	if !hasManaged {
		return nil
	}

	if strings.TrimSpace(newContent) == "" {
		cmd := exec.Command("crontab", "-r")
		_ = cmd.Run()
		return nil
	}

	if err := c.writeCrontab(newContent); err != nil {
		return fmt.Errorf("clear crontab entries: %w", err)
	}

	return nil
}

func (c *CrontabService) Exists() (bool, error) {
	curr, err := c.readCrontab()
	if err != nil {
		return false, err
	}
	return strings.Contains(curr, "# tuckify-managed"), nil
}

func (c *CrontabService) CheckStatus() (string, error) {
	return "To check crontab, run: crontab -l", nil
}

func (c *CrontabService) readCrontab() (string, error) {
	cmd := exec.Command("crontab", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return out.String(), nil
}

func (c *CrontabService) writeCrontab(content string) error {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec crontab -: %v (stderr: %q)", err, errOut.String())
	}
	return nil
}

func updateCrontabContent(currCrontab, binaryPath, folder, cronExpr, configPath string) string {
	lines := strings.Split(currCrontab, "\n")
	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.Contains(trimmed, "# tuckify-managed") {
			continue
		}
		newLines = append(newLines, line)
	}

	cronCmd := fmt.Sprintf("%s run %s", binaryPath, folder)
	if configPath != "" {
		cronCmd += fmt.Sprintf(" --config %s", configPath)
	}
	cronCmd += " # tuckify-managed"

	newLine := fmt.Sprintf("%s %s", cronExpr, cronCmd)
	newLines = append(newLines, newLine)

	return strings.Join(newLines, "\n") + "\n"
}

func removeCrontabContent(currCrontab string) (string, bool) {
	lines := strings.Split(currCrontab, "\n")
	var newLines []string
	hasManaged := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "# tuckify-managed") {
			hasManaged = true
			continue
		}
		if trimmed != "" || len(newLines) > 0 {
			newLines = append(newLines, line)
		}
	}

	newCrontabContent := strings.Join(newLines, "\n")
	if len(newLines) > 0 && !strings.HasSuffix(newCrontabContent, "\n") {
		newCrontabContent += "\n"
	}
	return newCrontabContent, hasManaged
}
