package homedirutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeHomePathFrom_ValidPath(t *testing.T) {
	// Use a valid absolute path
	homeDir := "/home/testuser"
	result, err := SafeHomePathFrom(homeDir, ".kube", "config")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Join(homeDir, ".kube", "config")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSafeHomePathFrom_DirectoryTraversal(t *testing.T) {
	// Path with directory traversal should be rejected
	maliciousPath := "/tmp/../etc"
	_, err := SafeHomePathFrom(maliciousPath, ".kube", "config")

	if err == nil {
		t.Error("Expected error for directory traversal attempt")
	}
	if !errors.Is(err, ErrTraversalAttempt) {
		t.Errorf("Expected ErrTraversalAttempt, got %v", err)
	}
}

func TestSafeHomePathFrom_RelativePath(t *testing.T) {
	// Relative path should be rejected
	relativePath := "relative/path"
	_, err := SafeHomePathFrom(relativePath, ".kube", "config")

	if err == nil {
		t.Error("Expected error for relative path")
	}
	if !errors.Is(err, ErrRelativePath) {
		t.Errorf("Expected ErrRelativePath, got %v", err)
	}
}

func TestSafeHomePathFrom_EmptyComponents(t *testing.T) {
	// Should work with no path components
	homeDir := "/home/testuser"
	result, err := SafeHomePathFrom(homeDir)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != homeDir {
		t.Errorf("Expected %q, got %q", homeDir, result)
	}
}

func TestSafeHomePathFrom_MultipleComponents(t *testing.T) {
	// Should work with multiple path components
	homeDir := "/home/testuser"
	result, err := SafeHomePathFrom(homeDir, ".aws", "cli", "alias")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Join(homeDir, ".aws", "cli", "alias")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSafeHomePath_Integration(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a valid HOME
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	result, err := SafeHomePath(".kube", "config")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Join(tmpDir, ".kube", "config")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestSafeHomePath_DirectoryTraversalPrevention(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Try to set a malicious HOME path with directory traversal
	maliciousPath := "/tmp/../etc"
	os.Setenv("HOME", maliciousPath)

	_, err := SafeHomePath(".kube", "config")
	if err == nil {
		t.Error("Expected error when HOME contains '..' (directory traversal attempt)")
	}
	if !errors.Is(err, ErrTraversalAttempt) {
		t.Errorf("Expected ErrTraversalAttempt, got %v", err)
	}
}

func TestSafeHomePath_RelativePathPrevention(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Try to set a relative HOME path
	relativePath := "relative/path"
	os.Setenv("HOME", relativePath)

	_, err := SafeHomePath(".kube", "config")
	if err == nil {
		t.Error("Expected error when HOME is relative")
	}
	if !errors.Is(err, ErrRelativePath) {
		t.Errorf("Expected ErrRelativePath, got %v", err)
	}
}
