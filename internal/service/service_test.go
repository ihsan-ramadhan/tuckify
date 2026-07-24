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
	case "darwin":
		_, ok := srv.(*LaunchdService)
		if !ok {
			t.Errorf("expected LaunchdService on darwin, got %T", srv)
		}
	case "windows":
		_, ok := srv.(*WintaskService)
		if !ok {
			t.Errorf("expected WintaskService on windows, got %T", srv)
		}
	}
}

func TestSystemdContent(t *testing.T) {
	name := "downloads"
	binary := "/usr/bin/tuckify"
	folders := []string{"/data"}
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildSystemdContent(name, binary, folders, cronExpr, cfgPath)
	expected := `[Unit]
Description=tuckify file organizer (downloads)
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule downloads "/data" --cron "0 9 * * *" --run --force --config "/etc/tuckify.toml"
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}

	contentNoCfg := buildSystemdContent(name, binary, folders, cronExpr, "")
	expectedNoCfg := `[Unit]
Description=tuckify file organizer (downloads)
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule downloads "/data" --cron "0 9 * * *" --run --force
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
	folders := []string{"/data"}
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildLaunchdContent(name, binary, folders, cronExpr, cfgPath)
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ihsan.tuckify.downloads</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/tuckify</string>
        <string>schedule</string>
        <string>downloads</string>
        <string>/data</string>
        <string>--cron</string>
        <string>0 9 * * *</string>
        <string>--run</string>
        <string>--force</string>
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

	contentNoCfg := buildLaunchdContent(name, binary, folders, cronExpr, "")
	expectedNoCfg := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ihsan.tuckify.downloads</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/tuckify</string>
        <string>schedule</string>
        <string>downloads</string>
        <string>/data</string>
        <string>--cron</string>
        <string>0 9 * * *</string>
        <string>--run</string>
        <string>--force</string>
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
	folders := []string{"/data"}
	cronExpr := "0 9 * * *"
	name := "downloads"

	result := upsertCrontabContent(initial, name, binary, folders, cronExpr, "/etc/tuckify.toml")
	expected := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify schedule downloads \"/data\" --cron \"0 9 * * *\" --run --force --config \"/etc/tuckify.toml\" # tuckify-managed:downloads\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	resultNoCfg := upsertCrontabContent(initial, name, binary, folders, cronExpr, "")
	expectedNoCfg := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify schedule downloads \"/data\" --cron \"0 9 * * *\" --run --force # tuckify-managed:downloads\n"
	if resultNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, resultNoCfg)
	}

	// upsert same name replaces existing line
	updated := upsertCrontabContent(result, name, binary, folders, "0 18 * * *", "")
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
