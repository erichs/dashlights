package signals

import (
	"context"
	"os"
	"runtime"
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

func TestPrivilegedPathSignal_Diagnostic(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	signal.diagnostic = "Test diagnostic"
	if signal.Diagnostic() != "Test diagnostic" {
		t.Errorf("Expected 'Test diagnostic', got '%s'", signal.Diagnostic())
	}
}

func TestPrivilegedPathSignal_Remediation(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	expected := "Remove '.' from PATH or move it to the end after system directories"
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

	if signal.diagnostic != "Current directory '.' found in PATH" {
		t.Errorf("Expected diagnostic about '.', got '%s'", signal.diagnostic)
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

	expected := "Empty path entry (::) found in PATH (implies current directory)"
	if signal.diagnostic != expected {
		t.Errorf("Expected diagnostic '%s', got '%s'", expected, signal.diagnostic)
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
