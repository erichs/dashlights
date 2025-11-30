package signals

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestRebootPendingSignal_Name(t *testing.T) {
	signal := NewRebootPendingSignal()
	if signal.Name() != "Reboot Pending" {
		t.Errorf("Expected 'Reboot Pending', got '%s'", signal.Name())
	}
}

func TestRebootPendingSignal_Emoji(t *testing.T) {
	signal := NewRebootPendingSignal()
	if signal.Emoji() != "♻️" {
		t.Errorf("Expected '♻️', got '%s'", signal.Emoji())
	}
}

func TestRebootPendingSignal_Diagnostic(t *testing.T) {
	signal := NewRebootPendingSignal()
	expected := "System reboot required to activate security patches"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestRebootPendingSignal_Remediation(t *testing.T) {
	signal := NewRebootPendingSignal()
	expected := "Reboot system to activate installed kernel patches"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestRebootPendingSignal_Check_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping non-Linux test on Linux")
	}

	signal := NewRebootPendingSignal()
	ctx := context.Background()

	// On non-Linux systems, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on non-Linux system")
	}
}

func TestRebootPendingSignal_Check_Linux_NoRebootRequired(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	// Check if the file exists
	_, err := os.Stat("/var/run/reboot-required")
	if err == nil {
		t.Skip("Skipping test because /var/run/reboot-required exists on this system")
	}

	signal := NewRebootPendingSignal()
	ctx := context.Background()

	// Should return false when file doesn't exist
	if signal.Check(ctx) {
		t.Error("Expected false when /var/run/reboot-required doesn't exist")
	}
}

func TestRebootPendingSignal_Check_Linux_RebootRequired(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	// This test can only verify behavior if the file exists
	_, err := os.Stat("/var/run/reboot-required")
	if err != nil {
		t.Skip("Skipping test because /var/run/reboot-required doesn't exist on this system")
	}

	signal := NewRebootPendingSignal()
	ctx := context.Background()

	// Should return true when file exists
	if !signal.Check(ctx) {
		t.Error("Expected true when /var/run/reboot-required exists")
	}
}
