package signals

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRootLoginSignal_Name(t *testing.T) {
	signal := NewRootLoginSignal()
	if signal.Name() != "Danger Zone" {
		t.Errorf("Expected 'Danger Zone', got '%s'", signal.Name())
	}
}

func TestRootLoginSignal_Emoji(t *testing.T) {
	signal := NewRootLoginSignal()
	if signal.Emoji() != "👑" {
		t.Errorf("Expected '👑', got '%s'", signal.Emoji())
	}
}

func TestRootLoginSignal_Diagnostic(t *testing.T) {
	signal := NewRootLoginSignal()
	expected := "Running as root user (UID 0)"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestRootLoginSignal_Remediation(t *testing.T) {
	signal := NewRootLoginSignal()
	expected := "Use a non-root account and run privileged commands with sudo instead"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestRootLoginSignal_Check_NonRoot(t *testing.T) {
	// Skip this test if running as root
	if os.Geteuid() == 0 {
		t.Skip("Skipping non-root test when running as root")
	}

	signal := NewRootLoginSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when not running as root")
	}
}

func TestRootLoginSignal_Check_Root(t *testing.T) {
	// This test verifies behavior when running as root
	// It will only pass when tests are run with root privileges
	if os.Geteuid() != 0 {
		t.Skip("Skipping root test when not running as root")
	}

	signal := NewRootLoginSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when running as root")
	}
}

func TestRootLoginSignal_Check_ContextCancelled(t *testing.T) {
	signal := NewRootLoginSignal()

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return false when context is cancelled
	if signal.Check(ctx) {
		t.Error("Expected false when context is cancelled")
	}
}

func TestRootLoginSignal_Check_ContextTimeout(t *testing.T) {
	signal := NewRootLoginSignal()

	// Create a context that has already timed out
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout to expire
	time.Sleep(1 * time.Millisecond)

	// Should return false when context has timed out
	if signal.Check(ctx) {
		t.Error("Expected false when context has timed out")
	}
}

func TestRootLoginSignal_Interface(t *testing.T) {
	t.Helper()
	// Verify that RootLoginSignal implements the Signal interface
	var _ Signal = (*RootLoginSignal)(nil)
	var _ Signal = NewRootLoginSignal()
}

func TestRootLoginSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_ROOT_LOGIN", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_ROOT_LOGIN")

	signal := NewRootLoginSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
