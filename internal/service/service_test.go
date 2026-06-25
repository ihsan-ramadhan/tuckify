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
