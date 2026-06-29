package service

import (
	"runtime"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	srv, err := NewService()
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil Service")
	}

	switch runtime.GOOS {
	case "linux":
		_, isSystemd := srv.(*SystemdService)
		_, isCrontab := srv.(*CrontabService)
		if !isSystemd && !isCrontab {
			t.Errorf("expected SystemdService or CrontabService on Linux, got %T", srv)
		}
	case "darwin", "windows":
		_, isPlaceholder := srv.(*placeholderService)
		if !isPlaceholder {
			t.Errorf("expected placeholderService on %s, got %T", runtime.GOOS, srv)
		}
	}
}

func TestSystemdContent(t *testing.T) {
	name := "downloads"
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildSystemdContent(name, binary, folder, cronExpr, cfgPath)
	expected := `[Unit]
Description=tuckify file organizer (downloads)
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule downloads /data --cron "0 9 * * *" --config /etc/tuckify.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}

	contentNoCfg := buildSystemdContent(name, binary, folder, cronExpr, "")
	expectedNoCfg := `[Unit]
Description=tuckify file organizer (downloads)
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule downloads /data --cron "0 9 * * *"
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
	if contentNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, contentNoCfg)
	}
}

func TestLaunchdContent(t *testing.T) {
	name := "downloads"
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildLaunchdContent(name, binary, folder, cronExpr, cfgPath)
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ihsan.tuckify</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/tuckify</string>
        <string>schedule</string>
        <string>downloads</string>
        <string>/data</string>
        <string>--cron</string>
        <string>0 9 * * *</string>
        <string>--config</string>
        <string>/etc/tuckify.toml</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}

	contentNoCfg := buildLaunchdContent(name, binary, folder, cronExpr, "")
	expectedNoCfg := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ihsan.tuckify</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/tuckify</string>
        <string>schedule</string>
        <string>downloads</string>
        <string>/data</string>
        <string>--cron</string>
        <string>0 9 * * *</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`
	if contentNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, contentNoCfg)
	}
}

func TestCrontabUpsert(t *testing.T) {
	initial := "* * * * * old-job\n"
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	name := "downloads"

	result := upsertCrontabContent(initial, name, binary, folder, cronExpr, "/etc/tuckify.toml")
	expected := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify run /data --config /etc/tuckify.toml # tuckify-managed:downloads\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	resultNoCfg := upsertCrontabContent(initial, name, binary, folder, cronExpr, "")
	expectedNoCfg := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify run /data # tuckify-managed:downloads\n"
	if resultNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, resultNoCfg)
	}

	// upsert same name replaces existing line
	updated := upsertCrontabContent(result, name, binary, folder, "0 18 * * *", "")
	if strings.Count(updated, "tuckify-managed:downloads") != 1 {
		t.Errorf("upsert should replace existing entry, got: %q", updated)
	}
}

func TestCrontabRemove(t *testing.T) {
	curr := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify run /data # tuckify-managed:downloads\n0 18 * * * /usr/bin/tuckify run /docs # tuckify-managed:docs\n"

	// remove specific name
	result, found := removeCrontabContent(curr, "downloads")
	if !found {
		t.Error("expected found=true")
	}
	if strings.Contains(result, "tuckify-managed:downloads") {
		t.Error("expected downloads entry removed")
	}
	if !strings.Contains(result, "tuckify-managed:docs") {
		t.Error("expected docs entry preserved")
	}

	// remove all
	resultAll, foundAll := removeCrontabContent(curr, "")
	if !foundAll {
		t.Error("expected foundAll=true")
	}
	if strings.Contains(resultAll, "tuckify-managed") {
		t.Error("expected all tuckify entries removed")
	}

	// remove non-existent
	_, notFound := removeCrontabContent(curr, "nonexistent")
	if notFound {
		t.Error("expected found=false for nonexistent name")
	}
}

func TestWintaskCmd(t *testing.T) {
	name := "downloads"
	binary := `C:\tuckify.exe`
	folder := `C:\data`
	cronExpr := "0 9 * * *"
	cfgPath := `C:\config.toml`

	cmd := buildWintaskCmd(name, binary, folder, cronExpr, cfgPath)
	expected := `"` + binary + `" schedule "` + name + `" "` + folder + `" --cron "` + cronExpr + `"` + ` --config "` + cfgPath + `"`
	if cmd != expected {
		t.Errorf("expected %q, got %q", expected, cmd)
	}

	cmdNoCfg := buildWintaskCmd(name, binary, folder, cronExpr, "")
	expectedNoCfg := `"` + binary + `" schedule "` + name + `" "` + folder + `" --cron "` + cronExpr + `"`
	if cmdNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, cmdNoCfg)
	}
}
