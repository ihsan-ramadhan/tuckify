package service

import (
	"runtime"
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

func TestCrontabContentUpdate(t *testing.T) {
	initial := "* * * * * old-job\n"
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	updated := updateCrontabContent(initial, binary, folder, cronExpr, cfgPath)
	expected := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify run /data --config /etc/tuckify.toml # tuckify-managed\n"
	if updated != expected {
		t.Errorf("expected %q, got %q", expected, updated)
	}
	
	updatedNoCfg := updateCrontabContent(initial, binary, folder, cronExpr, "")
	expectedNoCfg := "* * * * * old-job\n0 9 * * * /usr/bin/tuckify run /data # tuckify-managed\n"
	if updatedNoCfg != expectedNoCfg {
		t.Errorf("expected %q, got %q", expectedNoCfg, updatedNoCfg)
	}

	removed, ok := removeCrontabContent(updated)
	if !ok {
		t.Error("expected remove to return ok=true")
	}
	if removed != initial {
		t.Errorf("expected %q, got %q", initial, removed)
	}
}

func TestSystemdContent(t *testing.T) {
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildSystemdContent(binary, folder, cronExpr, cfgPath)
	expected := `[Unit]
Description=tuckify file organizer
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule /data --cron "0 9 * * *" --config /etc/tuckify.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}

	contentNoCfg := buildSystemdContent(binary, folder, cronExpr, "")
	expectedNoCfg := `[Unit]
Description=tuckify file organizer
After=default.target

[Service]
ExecStart=/usr/bin/tuckify schedule /data --cron "0 9 * * *"
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
	binary := "/usr/bin/tuckify"
	folder := "/data"
	cronExpr := "0 9 * * *"
	cfgPath := "/etc/tuckify.toml"

	content := buildLaunchdContent(binary, folder, cronExpr, cfgPath)
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

	contentNoCfg := buildLaunchdContent(binary, folder, cronExpr, "")
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



