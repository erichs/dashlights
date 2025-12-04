package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPrivilegedPathSignal_Name(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	if signal.Name() != "Privileged Path" {
		t.Errorf("Expected 'Privileged Path', got '%s'", signal.Name())
	}
}

func TestPrivilegedPathSignal_Emoji(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	if signal.Emoji() != "💣" {
		t.Errorf("Expected '💣', got '%s'", signal.Emoji())
	}
}

func TestPrivilegedPathSignal_Diagnostic_NoFindings(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	expected := "Potentially dangerous entries detected in PATH"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Diagnostic_MultipleFindings(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	signal.findings = []string{"Issue A", "Issue B"}

	expected := "Multiple PATH issues detected: Issue A; Issue B"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Remediation(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	expected := "Remove '.' and world-writable or user bin directories from PATH, or move user bin directories after system paths like /usr/bin"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestPrivilegedPathSignal_Check_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	// On Windows, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on Windows")
	}
}

func TestPrivilegedPathSignal_Check_EmptyPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set empty PATH
	os.Setenv("PATH", "")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when PATH is empty")
	}
}

func TestPrivilegedPathSignal_Check_DotInPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set PATH with '.'
	os.Setenv("PATH", "/usr/bin:.:/usr/local/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when '.' is in PATH")
	}

	expected := "Current directory '.' found in PATH"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Check_EmptyEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set PATH with empty entry (::)
	os.Setenv("PATH", "/usr/bin::/usr/local/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when PATH has empty entry")
	}

	expected := "Empty PATH entry (::) found (implies current directory)"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Check_SafePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set safe PATH without '.'
	os.Setenv("PATH", "/usr/bin:/usr/local/bin:/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when PATH is safe")
	}
}

func TestPrivilegedPathSignal_Check_WorldWritablePathEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o777); err != nil {
		t.Skipf("unable to set world-writable permissions on temp dir: %v", err)
	}

	os.Setenv("PATH", dir)

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when PATH has world-writable directory")
	}

	diag := signal.Diagnostic()
	if !strings.Contains(diag, "World-writable PATH entry") || !strings.Contains(diag, dir) {
		t.Errorf("Expected diagnostic to mention world-writable PATH entry and directory, got %q", diag)
	}
}

func TestPrivilegedPathSignal_Check_UserBinBeforeSystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH and HOME
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	homeDir := t.TempDir()
	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	userBin := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatalf("Failed to create user bin directory: %v", err)
	}

	os.Setenv("PATH", userBin+string(os.PathListSeparator)+"/usr/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when user bin directory appears before system directories")
	}

	expected := "User PATH directory $HOME/bin appears before system directories"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_isSystemPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/bin", true},
		{"/sbin", true},
		{"/usr/bin", true},
		{"/usr/sbin", true},
		{"/usr/local/bin", true},
		{"/usr/local/sbin", true},
		{"/home/user/bin", false},
		{"/opt/bin", false},
		{".", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isSystemPath(tt.path)
		if result != tt.expected {
			t.Errorf("isSystemPath(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}
