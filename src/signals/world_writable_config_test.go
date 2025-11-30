package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWorldWritableConfigSignal_Name(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	if signal.Name() != "World Writable Config" {
		t.Errorf("Expected 'World Writable Config', got '%s'", signal.Name())
	}
}

func TestWorldWritableConfigSignal_Emoji(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	if signal.Emoji() != "🖊️" {
		t.Errorf("Expected '🖊️', got '%s'", signal.Emoji())
	}
}

func TestWorldWritableConfigSignal_Diagnostic_NoFiles(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	signal.foundFiles = []string{}
	expected := "World-writable config files detected"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestWorldWritableConfigSignal_Diagnostic_WithFiles(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	signal.foundFiles = []string{".bashrc", ".zshrc"}
	expected := "World-writable: .bashrc"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestWorldWritableConfigSignal_Remediation(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	expected := "Fix permissions with: chmod 644 <file>"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestWorldWritableConfigSignal_Check_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	signal := NewWorldWritableConfigSignal()
	ctx := context.Background()

	// On Windows, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on Windows")
	}
}

func TestWorldWritableConfigSignal_Check_NoConfigFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory without config files
	tmpDir := t.TempDir()

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewWorldWritableConfigSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when no config files exist")
	}
}

func TestWorldWritableConfigSignal_Check_SecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory with secure config file
	tmpDir := t.TempDir()
	bashrc := filepath.Join(tmpDir, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewWorldWritableConfigSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when config file has secure permissions (0644)")
	}
}

func TestWorldWritableConfigSignal_Check_WorldWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory with world-writable config file
	tmpDir := t.TempDir()
	bashrc := filepath.Join(tmpDir, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Explicitly set world-writable permissions
	if err := os.Chmod(bashrc, 0666); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewWorldWritableConfigSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when config file is world-writable (0666)")
	}

	if len(signal.foundFiles) == 0 {
		t.Error("Expected foundFiles to contain .bashrc")
	}
}
