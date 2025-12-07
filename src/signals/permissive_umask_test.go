package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
)

func TestPermissiveUmaskSignal_Name(t *testing.T) {
	signal := NewPermissiveUmaskSignal()
	if signal.Name() != "Loose Cannon" {
		t.Errorf("Expected 'Loose Cannon', got '%s'", signal.Name())
	}
}

func TestPermissiveUmaskSignal_Emoji(t *testing.T) {
	signal := NewPermissiveUmaskSignal()
	if signal.Emoji() != "😷" {
		t.Errorf("Expected '😷', got '%s'", signal.Emoji())
	}
}

func TestPermissiveUmaskSignal_Diagnostic(t *testing.T) {
	signal := NewPermissiveUmaskSignal()
	signal.currentUmask = 0002
	expected := "Permissive umask detected: 0002"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestPermissiveUmaskSignal_Remediation(t *testing.T) {
	signal := NewPermissiveUmaskSignal()
	expected := "Set umask to 0022 or 0027 for better security"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestPermissiveUmaskSignal_Check(t *testing.T) {
	t.Helper()
	// Save and restore umask
	oldUmask := syscall.Umask(0)
	syscall.Umask(oldUmask)
	defer syscall.Umask(oldUmask)

	signal := NewPermissiveUmaskSignal()
	ctx := context.Background()

	// Just verify it doesn't panic
	signal.Check(ctx)
}

func TestPermissiveUmaskSignal_formatUmask(t *testing.T) {
	tests := []struct {
		umask    int
		expected string
	}{
		{0000, "0000"},
		{0002, "0002"},
		{0022, "0022"},
		{0027, "0027"},
		{0077, "0077"},
		{0777, "0777"},
	}

	for _, tt := range tests {
		result := formatUmask(tt.umask)
		if result != tt.expected {
			t.Errorf("formatUmask(0%o) = %s, expected %s", tt.umask, result, tt.expected)
		}
	}
}

func TestPermissiveUmaskSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_PERMISSIVE_UMASK", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_PERMISSIVE_UMASK")

	signal := NewPermissiveUmaskSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
