package signals

import (
	"context"
	"runtime"
	"testing"
)

func TestDiskSpaceSignal_Name(t *testing.T) {
	signal := NewDiskSpaceSignal()
	if signal.Name() != "Full Tank" {
		t.Errorf("Expected 'Full Tank', got '%s'", signal.Name())
	}
}

func TestDiskSpaceSignal_Emoji(t *testing.T) {
	signal := NewDiskSpaceSignal()
	if signal.Emoji() != "💾" {
		t.Errorf("Expected '💾', got '%s'", signal.Emoji())
	}
}

func TestDiskSpaceSignal_Diagnostic(t *testing.T) {
	signal := NewDiskSpaceSignal()
	signal.path = "/tmp"
	signal.percentUsed = 95
	expected := "/tmp is 95% full"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestDiskSpaceSignal_Remediation(t *testing.T) {
	signal := NewDiskSpaceSignal()
	expected := "Free up disk space to prevent logging and audit trail failures"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestDiskSpaceSignal_Check_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	signal := NewDiskSpaceSignal()
	ctx := context.Background()

	// On Windows, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on Windows")
	}
}

func TestDiskSpaceSignal_Check_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	signal := NewDiskSpaceSignal()
	ctx := context.Background()

	// Just verify it doesn't panic
	// May return true or false depending on actual disk usage
	signal.Check(ctx)
}

func TestDiskSpaceSignal_checkPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	signal := NewDiskSpaceSignal()

	// Check root path - should not panic
	signal.checkPath("/")

	// Check invalid path - should return false
	if signal.checkPath("/nonexistent/path/that/does/not/exist") {
		t.Error("Expected false for nonexistent path")
	}
}
