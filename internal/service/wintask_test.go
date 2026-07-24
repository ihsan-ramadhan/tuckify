//go:build windows

package service

import (
	"testing"
)

func TestWintaskCmd(t *testing.T) {
	name := "downloads"
	binary := `C:\tuckify.exe`
	folders := []string{`C:\data`}
	cronExpr := "0 9 * * *"
	cfgPath := `C:\config.toml`

	cmd := buildWintaskCmd(name, binary, folders, cronExpr, cfgPath)
	expected := `"` + binary + `" schedule "` + name + `" "` + folders[0] + `" --cron "` + cronExpr + `" --run --force` + ` --config "` + cfgPath + `"`
	if cmd != expected {
		t.Errorf("expected %q, got %q", expected, cmd)
	}

	cmdNoCfg := buildWintaskCmd(name, binary, folders, cronExpr, "")
	expectedNoCfg := `"` + binary + `" schedule "` + name + `" "` + folders[0] + `" --cron "` + cronExpr + `" --run --force`
	if cmdNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, cmdNoCfg)
	}
}
