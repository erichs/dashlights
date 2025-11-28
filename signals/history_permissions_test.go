package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryPermissionsSignal_NoHistoryFiles(t *testing.T) {
	// Create a temporary directory with no history files
	tmpDir := t.TempDir()
	
	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no history files exist")
	}
}

func TestHistoryPermissionsSignal_BashHistory600(t *testing.T) {
	// Create a temporary directory with secure bash_history
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, ".bash_history")
	
	// Create file with secure permissions (600)
	err := os.WriteFile(histFile, []byte("echo hello\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when bash_history has 600 permissions")
	}
}

func TestHistoryPermissionsSignal_ZshHistory600(t *testing.T) {
	// Create a temporary directory with secure zsh_history
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, ".zsh_history")
	
	// Create file with secure permissions (600)
	err := os.WriteFile(histFile, []byte("echo hello\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when zsh_history has 600 permissions")
	}
}

func TestHistoryPermissionsSignal_BashHistory644(t *testing.T) {
	// Create a temporary directory with world-readable bash_history
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, ".bash_history")
	
	// Create file with insecure permissions (644 - world readable)
	err := os.WriteFile(histFile, []byte("echo hello\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when bash_history has 644 permissions")
	}
}

func TestHistoryPermissionsSignal_ZshHistory644(t *testing.T) {
	// Create a temporary directory with world-readable zsh_history
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, ".zsh_history")
	
	// Create file with insecure permissions (644 - world readable)
	err := os.WriteFile(histFile, []byte("echo hello\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when zsh_history has 644 permissions")
	}
}

func TestHistoryPermissionsSignal_GroupReadable(t *testing.T) {
	// Create a temporary directory with group-readable history
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, ".bash_history")
	
	// Create file with group-readable permissions (640)
	err := os.WriteFile(histFile, []byte("echo hello\n"), 0640)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when bash_history has 640 permissions (group readable)")
	}
}

func TestHistoryPermissionsSignal_MultipleFiles(t *testing.T) {
	// Create a temporary directory with multiple history files
	tmpDir := t.TempDir()
	
	// Create bash_history with secure permissions
	bashHist := filepath.Join(tmpDir, ".bash_history")
	err := os.WriteFile(bashHist, []byte("echo hello\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create bash_history: %v", err)
	}

	// Create zsh_history with insecure permissions
	zshHist := filepath.Join(tmpDir, ".zsh_history")
	err = os.WriteFile(zshHist, []byte("echo hello\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create zsh_history: %v", err)
	}

	// Temporarily change HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewHistoryPermissionsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when any history file has insecure permissions")
	}
}

