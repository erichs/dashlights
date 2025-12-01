package signals

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestContextCancellation_ZombieProcesses verifies that zombie_processes respects context cancellation
func TestContextCancellation_ZombieProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	// Create a mock filesystem with many processes
	mockFS := &mockProcFS{
		files: make(map[string][]byte),
		dirs:  make(map[string][]os.DirEntry),
	}

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Create 1000 mock processes (simulating a busy system)
	entries := make([]os.DirEntry, 1000)
	for i := 0; i < 1000; i++ {
		pid := fmt.Sprintf("%d", i+1)
		entries[i] = &mockDirEntry{name: pid, isDir: true}
		mockFS.files[filepath.Join("/proc", pid, "stat")] = []byte(fmt.Sprintf("%d (process) Z 0 1 1", i+1))
	}
	mockFS.dirs["/proc"] = entries

	signal := NewZombieProcessesSignal()

	// Create a context with very short timeout (1ms)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	// The check should return quickly due to context cancellation
	start := time.Now()
	result := signal.checkWithFS(ctx, mockFS)
	elapsed := time.Since(start)

	// Should return false when context is cancelled
	if result {
		t.Error("Expected false when context is cancelled")
	}

	// Should complete quickly (within 10ms) even with 1000 processes
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
	}
}

// TestContextCancellation_PyCachePollution verifies that pycache_pollution respects context cancellation
func TestContextCancellation_PyCachePollution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	// Create a temporary directory with deep nesting
	tmpDir, err := os.MkdirTemp("", "pycache_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create .git directory to make it look like a git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git: %v", err)
	}

	// Create deep directory structure (100 levels)
	currentPath := "."
	for i := 0; i < 100; i++ {
		currentPath = filepath.Join(currentPath, "level")
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	signal := NewPyCachePollutionSignal()

	// Create a context with very short timeout (1ms)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	// The check should return quickly due to context cancellation
	start := time.Now()
	result := signal.Check(ctx)
	elapsed := time.Since(start)

	// Should return false when context is cancelled
	if result {
		t.Error("Expected false when context is cancelled")
	}

	// Should complete quickly (within 10ms) even with deep nesting
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
	}
}

// TestContextCancellation_MissingInitPy verifies that missing_init_py respects context cancellation
func TestContextCancellation_MissingInitPy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	// Create a temporary directory with deep nesting
	tmpDir, err := os.MkdirTemp("", "initpy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create deep directory structure (100 levels)
	currentPath := "."
	for i := 0; i < 100; i++ {
		currentPath = filepath.Join(currentPath, "pkg")
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		// Add a .py file to make it look like a package
		pyFile := filepath.Join(currentPath, "module.py")
		if err := os.WriteFile(pyFile, []byte("# Python module"), 0644); err != nil {
			t.Fatalf("Failed to create .py file: %v", err)
		}
	}

	signal := NewMissingInitPySignal()

	// Create a context with very short timeout (1ms)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	// The check should return quickly due to context cancellation
	start := time.Now()
	result := signal.Check(ctx)
	elapsed := time.Since(start)

	// Should return false when context is cancelled
	if result {
		t.Error("Expected false when context is cancelled")
	}

	// Should complete quickly (within 10ms) even with deep nesting
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
	}
}

// TestContextCancellation_SnapshotDependency verifies that snapshot_dependency respects context cancellation
func TestContextCancellation_SnapshotDependency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	// Create a temporary git repository with many tags
	tmpDir, err := os.MkdirTemp("", "snapshot_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create .git/refs/tags directory
	tagsDir := filepath.Join(".git", "refs", "tags")
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		t.Fatalf("Failed to create tags directory: %v", err)
	}

	// Create .git/HEAD
	headContent := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headContent, 0644); err != nil {
		t.Fatalf("Failed to create HEAD: %v", err)
	}

	// Create refs/heads/main
	if err := os.MkdirAll(".git/refs/heads", 0755); err != nil {
		t.Fatalf("Failed to create refs/heads: %v", err)
	}
	mainSHA := "abc123def456\n"
	if err := os.WriteFile(".git/refs/heads/main", []byte(mainSHA), 0644); err != nil {
		t.Fatalf("Failed to create main ref: %v", err)
	}

	// Create 1000 tag files
	for i := 0; i < 1000; i++ {
		tagFile := filepath.Join(tagsDir, fmt.Sprintf("v1.0.%d", i))
		tagSHA := fmt.Sprintf("different_sha_%d\n", i)
		if err := os.WriteFile(tagFile, []byte(tagSHA), 0644); err != nil {
			t.Fatalf("Failed to create tag file: %v", err)
		}
	}

	signal := NewSnapshotDependencySignal()

	// Create a context with very short timeout (1ms)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	// The check should return quickly due to context cancellation
	start := time.Now()
	result := signal.Check(ctx)
	elapsed := time.Since(start)

	// Should return false when context is cancelled
	if result {
		t.Error("Expected false when context is cancelled")
	}

	// Should complete quickly (within 10ms) even with 1000 tags
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
	}
}

// TestContextCancellation_FileScanning verifies that file scanning signals respect context cancellation
func TestContextCancellation_FileScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	tests := []struct {
		name   string
		signal Signal
		setup  func(t *testing.T, dir string)
	}{
		{
			name:   "root_kube_context",
			signal: NewRootKubeContextSignal(),
			setup: func(t *testing.T, dir string) {
				kubeDir := filepath.Join(dir, ".kube")
				if err := os.MkdirAll(kubeDir, 0755); err != nil {
					t.Fatalf("Failed to create .kube: %v", err)
				}
				// Create a large kubeconfig with 1000 lines
				var content string
				for i := 0; i < 1000; i++ {
					content += "# Comment line " + string(rune(i)) + "\n"
				}
				content += "current-context: test\n"
				if err := os.WriteFile(filepath.Join(kubeDir, "config"), []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create kubeconfig: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "context_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Setup test environment
			tt.setup(t, tmpDir)

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			// Create a context with very short timeout (1ms)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			// Wait for context to expire
			time.Sleep(2 * time.Millisecond)

			// The check should return quickly due to context cancellation
			start := time.Now()
			result := tt.signal.Check(ctx)
			elapsed := time.Since(start)

			// Should return false when context is cancelled
			if result {
				t.Error("Expected false when context is cancelled")
			}

			// Should complete quickly (within 10ms)
			if elapsed > 10*time.Millisecond {
				t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
			}
		})
	}
}
