package signals

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestRootOwnedHomeSignal_Name(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	if signal.Name() != "Root Squatter" {
		t.Errorf("Expected 'Root Squatter', got '%s'", signal.Name())
	}
}

func TestRootOwnedHomeSignal_Emoji(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	if signal.Emoji() != "🍄" {
		t.Errorf("Expected '🍄', got '%s'", signal.Emoji())
	}
}

func TestRootOwnedHomeSignal_Diagnostic_NoFiles(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	signal.foundFiles = []string{}
	expected := "Root-owned files found in home directory"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestRootOwnedHomeSignal_Diagnostic_WithFiles(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	signal.foundFiles = []string{".bashrc", ".zshrc"}
	expected := "Root-owned: .bashrc"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestRootOwnedHomeSignal_Remediation(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	expected := "Fix ownership with: sudo chown -R $USER:$USER <file>"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestRootOwnedHomeSignal_Check_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	signal := NewRootOwnedHomeSignal()
	ctx := context.Background()

	// On Windows, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on Windows")
	}
}

func TestRootOwnedHomeSignal_Check_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	signal := NewRootOwnedHomeSignal()
	ctx := context.Background()

	// Just verify it doesn't panic
	// May return true or false depending on actual file ownership
	signal.Check(ctx)
}

func TestRootOwnedHomeSignal_Check_NoHomeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Set HOME to empty
	os.Setenv("HOME", "")

	signal := NewRootOwnedHomeSignal()
	ctx := context.Background()

	// Should return false when HOME is not set
	if signal.Check(ctx) {
		t.Error("Expected false when HOME is not set")
	}
}

func TestRootOwnedHomeSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_ROOT_OWNED_HOME", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_ROOT_OWNED_HOME")

	signal := NewRootOwnedHomeSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
