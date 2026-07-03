package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const crontabCmd = "crontab"

type CrontabService struct{}

func NewCrontabService() *CrontabService {
	return &CrontabService{}
}

func managedTag(name string) string {
	if name == "" {
		return "# tuckify-managed"
	}
	return "# tuckify-managed:" + name
}

func buildCrontabLine(name, binaryPath string, folders []string, cronExpr, configPath string) string {
	escapedFolders := make([]string, len(folders))
	for i, f := range folders {
		escapedFolders[i] = fmt.Sprintf("%q", f)
	}
	cronCmd := fmt.Sprintf("%s schedule %s %s --cron %q --run", binaryPath, name, strings.Join(escapedFolders, " "), cronExpr)
	if configPath != "" {
		cronCmd += fmt.Sprintf(" --config %q", configPath)
	}
	cronCmd += " " + managedTag(name)
	return fmt.Sprintf("%s %s", cronExpr, cronCmd)
}

func upsertCrontabContent(curr, name, binaryPath string, folders []string, cronExpr, configPath string) string {
	tag := managedTag(name)
	var kept []string
	for _, line := range strings.Split(curr, "\n") {
		if strings.TrimSpace(line) == "" || strings.Contains(line, tag) {
			continue
		}
		kept = append(kept, line)
	}
	kept = append(kept, buildCrontabLine(name, binaryPath, folders, cronExpr, configPath))
	return strings.Join(kept, "\n") + "\n"
}

func removeCrontabContent(curr, name string) (string, bool) {
	var kept []string
	found := false
	for _, line := range strings.Split(curr, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if name == "" {
			if strings.Contains(trimmed, "# tuckify-managed") {
				found = true
				continue
			}
		} else {
			if strings.Contains(trimmed, managedTag(name)) {
				found = true
				continue
			}
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return "", found
	}
	return strings.Join(kept, "\n") + "\n", found
}

func (c *CrontabService) Install(name string, folders []string, cronExpr, configPath string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	curr, err := c.readCrontab()
	if err != nil {
		return fmt.Errorf("read crontab: %w", err)
	}

	content := upsertCrontabContent(curr, name, binaryPath, folders, cronExpr, configPath)
	if err := c.writeCrontab(content); err != nil {
		return fmt.Errorf("write crontab: %w", err)
	}
	return nil
}

func (c *CrontabService) Uninstall(name string) error {
	curr, err := c.readCrontab()
	if err != nil {
		return fmt.Errorf("read crontab: %w", err)
	}

	content, found := removeCrontabContent(curr, name)
	if !found {
		return nil
	}

	if content == "" {
		cronPath, err := exec.LookPath(crontabCmd)
		if err != nil {
			return fmt.Errorf("find crontab: %w", err)
		}
		_ = exec.Command(cronPath, "-r").Run()
		return nil
	}

	if err := c.writeCrontab(content); err != nil {
		return fmt.Errorf("write crontab: %w", err)
	}
	return nil
}

func (c *CrontabService) Exists(name string) (bool, error) {
	curr, err := c.readCrontab()
	if err != nil {
		return false, err
	}
	return strings.Contains(curr, managedTag(name)), nil
}

func (c *CrontabService) CheckStatus() (string, error) {
	return "Check crontab: crontab -l", nil
}

func (c *CrontabService) Logs(name string, follow bool, lines int) error {
	return fmt.Errorf("logs not available for crontab-based schedules — check syslog or /var/log/syslog")
}

func (c *CrontabService) readCrontab() (string, error) {
	cronPath, err := exec.LookPath(crontabCmd)
	if err != nil {
		return "", fmt.Errorf("find crontab: %w", err)
	}
	cmd := exec.Command(cronPath, "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return out.String(), nil
}

func (c *CrontabService) writeCrontab(content string) error {
	cronPath, err := exec.LookPath(crontabCmd)
	if err != nil {
		return fmt.Errorf("find crontab: %w", err)
	}
	cmd := exec.Command(cronPath, "-")
	cmd.Stdin = strings.NewReader(content)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec crontab -: %v (stderr: %q)", err, errOut.String())
	}
	return nil
}
