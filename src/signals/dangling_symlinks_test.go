package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDanglingSymlinksSignal_NoSymlinks(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create some regular files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	signal := NewDanglingSymlinksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no symlinks exist")
	}
}

func TestDanglingSymlinksSignal_ValidSymlinks(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Create valid symlink
	symlinkPath := filepath.Join(tmpDir, "link.txt")
	err := os.Symlink(targetFile, symlinkPath)
	if err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	signal := NewDanglingSymlinksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when all symlinks are valid")
	}
}

func TestDanglingSymlinksSignal_DanglingSymlink(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create target file temporarily
	targetFile := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Create symlink
	symlinkPath := filepath.Join(tmpDir, "link.txt")
	err := os.Symlink(targetFile, symlinkPath)
	if err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Remove target file to make symlink dangling
	os.Remove(targetFile)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	signal := NewDanglingSymlinksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when dangling symlink exists")
	}
}

func TestDanglingSymlinksSignal_MixedSymlinks(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create valid target
	validTarget := filepath.Join(tmpDir, "valid_target.txt")
	os.WriteFile(validTarget, []byte("content"), 0644)

	// Create valid symlink
	validLink := filepath.Join(tmpDir, "valid_link.txt")
	err := os.Symlink(validTarget, validLink)
	if err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Create dangling symlink (target never existed)
	danglingLink := filepath.Join(tmpDir, "dangling_link.txt")
	err = os.Symlink(filepath.Join(tmpDir, "nonexistent.txt"), danglingLink)
	if err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	signal := NewDanglingSymlinksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when at least one dangling symlink exists")
	}
}

func TestDanglingSymlinksSignal_RelativeSymlink(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(targetFile, []byte("content"), 0644)

	// Create relative symlink
	symlinkPath := filepath.Join(tmpDir, "link.txt")
	err := os.Symlink("target.txt", symlinkPath)
	if err != nil {
		t.Skipf("Cannot create symlinks on this system: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	signal := NewDanglingSymlinksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false for valid relative symlink")
	}

	// Now remove target to make it dangling
	os.Remove(targetFile)

	result = signal.Check(ctx)
	if !result {
		t.Error("Expected true for dangling relative symlink")
	}
}
