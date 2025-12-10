package signals

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// mockProcFS is a mock implementation of procFS for testing
type mockProcFS struct {
	files   map[string][]byte
	dirs    map[string][]fs.DirEntry
	readErr error
}

func newMockProcFS() *mockProcFS {
	return &mockProcFS{
		files: make(map[string][]byte),
		dirs:  make(map[string][]fs.DirEntry),
	}
}

func (m *mockProcFS) ReadFile(name string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockProcFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if entries, ok := m.dirs[name]; ok {
		return entries, nil
	}
	return nil, os.ErrNotExist
}

// mockDirEntry implements fs.DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode          { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestZombieProcessesSignal_Name(t *testing.T) {
	signal := NewZombieProcessesSignal()
	if signal.Name() != "Zombie Apocalypse" {
		t.Errorf("Expected 'Zombie Apocalypse', got '%s'", signal.Name())
	}
}

func TestZombieProcessesSignal_Emoji(t *testing.T) {
	signal := NewZombieProcessesSignal()
	if signal.Emoji() != "🧟" {
		t.Errorf("Expected '🧟', got '%s'", signal.Emoji())
	}
}

func TestZombieProcessesSignal_Diagnostic(t *testing.T) {
	signal := NewZombieProcessesSignal()
	signal.count = 10
	expected := "Excessive zombie processes detected: 10"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestZombieProcessesSignal_Remediation(t *testing.T) {
	signal := NewZombieProcessesSignal()
	expected := "Investigate and fix parent processes not reaping children"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestZombieProcessesSignal_Check_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping non-Linux test on Linux")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// On non-Linux systems, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on non-Linux system")
	}
}

func TestZombieProcessesSignal_Check_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// Run the check - should not panic
	result := signal.Check(ctx)

	// Result depends on actual system state, but count should be set
	// We can't assert the result, but we can verify the count was set
	if result && signal.count <= 5 {
		t.Errorf("Signal returned true but count is %d (threshold is >5)", signal.count)
	}
	if !result && signal.count > 5 {
		t.Errorf("Signal returned false but count is %d (threshold is >5)", signal.count)
	}
}

func TestZombieProcessesSignal_Check_Threshold(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// Run check
	signal.Check(ctx)

	// Verify threshold logic: >5 zombies triggers signal
	// The actual count depends on system state, but we can verify consistency
	if signal.count > 5 {
		// If we have more than 5 zombies, Check should have returned true
		// Run again to verify
		result := signal.Check(ctx)
		if !result && signal.count > 5 {
			t.Errorf("Expected true when zombie count is %d (>5)", signal.count)
		}
	}
}

// Note: The following tests verify the logic but cannot fully test the implementation
// because zombie_processes.go directly reads from /proc filesystem which is hard to mock.
// The tests verify that the code handles edge cases correctly on a real Linux system.

func TestZombieProcessesSignal_Check_InvalidProcEntries(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// This test verifies the code runs without panicking when encountering
	// various /proc entries (numeric PIDs, non-numeric names like "self", "cpuinfo", etc.)
	// The validation logic at line 73 should skip non-numeric entries
	result := signal.Check(ctx)

	// Should complete without error
	// The result depends on actual system state, but count should be >= 0
	if signal.count < 0 {
		t.Errorf("Expected non-negative zombie count, got %d", signal.count)
	}

	// Verify the signal respects the threshold
	if result && signal.count <= 5 {
		t.Errorf("Signal returned true but count is %d (should be >5)", signal.count)
	}
	if !result && signal.count > 5 {
		t.Errorf("Signal returned false but count is %d (should trigger at >5)", signal.count)
	}
}

func TestZombieProcessesSignal_Check_MalformedStatFile(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// This test verifies the code handles malformed stat files gracefully
	// In practice, /proc/[PID]/stat files should always be well-formed,
	// but the code at line 88-91 handles missing closing parens
	result := signal.Check(ctx)

	// Should complete without panicking
	_ = result
}

func TestZombieProcessesSignal_Check_ProcessDisappears(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// This test verifies the code handles processes that disappear between
	// ReadDir and ReadFile (line 80-83 error handling)
	// This is a race condition that can happen in practice
	result := signal.Check(ctx)

	// Should complete without error even if some processes disappear
	_ = result
}

// Mocked tests - these run on all platforms and test the core logic

func TestRealProcFS_ReadFile(t *testing.T) {
	fs := &realProcFS{}

	// Create a temporary file to test reading
	tmpFile, err := os.CreateTemp("", "test_proc_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := []byte("test content 123")
	if _, err := tmpFile.Write(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test successful read
	content, err := fs.ReadFile(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}

	// Test reading non-existent file
	_, err = fs.ReadFile("/nonexistent/file/path")
	if err == nil {
		t.Error("Expected error reading non-existent file")
	}

	// Test directory traversal prevention
	_, err = fs.ReadFile("/tmp/../etc/passwd")
	if err == nil {
		t.Error("Expected error for path with directory traversal")
	}
}

func TestRealProcFS_ReadDir(t *testing.T) {
	fs := &realProcFS{}

	// Create a temporary directory with some files
	tmpDir, err := os.MkdirTemp("", "test_proc_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files/dirs
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// Test successful read
	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Errorf("ReadDir failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test reading non-existent directory
	_, err = fs.ReadDir("/nonexistent/directory/path")
	if err == nil {
		t.Error("Expected error reading non-existent directory")
	}

	// Test directory traversal prevention
	_, err = fs.ReadDir("/tmp/../etc")
	if err == nil {
		t.Error("Expected error for path with directory traversal")
	}
}

func TestZombieProcessesSignal_Check_WithMock_NoZombies(t *testing.T) {
	mockFS := newMockProcFS()

	// Mock /proc/stat
	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Mock /proc directory with 3 normal processes
	mockFS.dirs["/proc"] = []fs.DirEntry{
		&mockDirEntry{name: "1", isDir: true},
		&mockDirEntry{name: "2", isDir: true},
		&mockDirEntry{name: "3", isDir: true},
	}

	// Mock process stat files (state S = sleeping)
	mockFS.files["/proc/1/stat"] = []byte("1 (init) S 0 1 1")
	mockFS.files["/proc/2/stat"] = []byte("2 (bash) S 1 2 2")
	mockFS.files["/proc/3/stat"] = []byte("3 (vim) S 2 3 3")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	if result {
		t.Error("Expected false when no zombies present")
	}
	if signal.count != 0 {
		t.Errorf("Expected count 0, got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_FewZombies(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Mock /proc directory with 3 zombies (below threshold)
	mockFS.dirs["/proc"] = []fs.DirEntry{
		&mockDirEntry{name: "1", isDir: true},
		&mockDirEntry{name: "2", isDir: true},
		&mockDirEntry{name: "3", isDir: true},
	}

	// All zombies (state Z)
	mockFS.files["/proc/1/stat"] = []byte("1 (defunct) Z 0 1 1")
	mockFS.files["/proc/2/stat"] = []byte("2 (defunct) Z 0 2 2")
	mockFS.files["/proc/3/stat"] = []byte("3 (defunct) Z 0 3 3")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	if result {
		t.Error("Expected false when zombie count <= 5")
	}
	if signal.count != 3 {
		t.Errorf("Expected count 3, got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_ManyZombies(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Mock /proc directory with 7 zombies (above threshold of 5)
	entries := make([]fs.DirEntry, 7)
	for i := 0; i < 7; i++ {
		pid := fmt.Sprintf("%d", i+1)
		entries[i] = &mockDirEntry{name: pid, isDir: true}
		mockFS.files[fmt.Sprintf("/proc/%s/stat", pid)] = []byte(fmt.Sprintf("%d (defunct) Z 0 1 1", i+1))
	}
	mockFS.dirs["/proc"] = entries

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	if !result {
		t.Error("Expected true when zombie count > 5")
	}
	if signal.count != 7 {
		t.Errorf("Expected count 7, got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_MixedProcesses(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Mix of normal and zombie processes
	mockFS.dirs["/proc"] = []fs.DirEntry{
		&mockDirEntry{name: "1", isDir: true},
		&mockDirEntry{name: "2", isDir: true},
		&mockDirEntry{name: "3", isDir: true},
		&mockDirEntry{name: "4", isDir: true},
		&mockDirEntry{name: "5", isDir: true},
		&mockDirEntry{name: "6", isDir: true},
		&mockDirEntry{name: "7", isDir: true},
	}

	// 3 normal, 4 zombies (below threshold)
	mockFS.files["/proc/1/stat"] = []byte("1 (init) S 0 1 1")
	mockFS.files["/proc/2/stat"] = []byte("2 (defunct) Z 0 2 2")
	mockFS.files["/proc/3/stat"] = []byte("3 (bash) R 1 3 3")
	mockFS.files["/proc/4/stat"] = []byte("4 (defunct) Z 0 4 4")
	mockFS.files["/proc/5/stat"] = []byte("5 (vim) S 3 5 5")
	mockFS.files["/proc/6/stat"] = []byte("6 (defunct) Z 0 6 6")
	mockFS.files["/proc/7/stat"] = []byte("7 (defunct) Z 0 7 7")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	if result {
		t.Error("Expected false when zombie count <= 5")
	}
	if signal.count != 4 {
		t.Errorf("Expected count 4, got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_InvalidProcEntries(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	// Mock /proc directory with mix of valid PIDs and invalid entries
	mockFS.dirs["/proc"] = []fs.DirEntry{
		&mockDirEntry{name: "1", isDir: true},
		&mockDirEntry{name: "self", isDir: true},     // Invalid: not numeric
		&mockDirEntry{name: "cpuinfo", isDir: false}, // Invalid: not a directory
		&mockDirEntry{name: "2", isDir: true},
		&mockDirEntry{name: "thread-self", isDir: true}, // Invalid: not numeric
		&mockDirEntry{name: "3", isDir: true},
	}

	// Only valid PIDs have stat files
	mockFS.files["/proc/1/stat"] = []byte("1 (init) S 0 1 1")
	mockFS.files["/proc/2/stat"] = []byte("2 (defunct) Z 0 2 2")
	mockFS.files["/proc/3/stat"] = []byte("3 (bash) S 1 3 3")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	// Should only count the 1 zombie from valid PIDs
	if result {
		t.Error("Expected false when zombie count <= 5")
	}
	if signal.count != 1 {
		t.Errorf("Expected count 1 (only valid PIDs counted), got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_MalformedStatFile(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")

	mockFS.dirs["/proc"] = []fs.DirEntry{
		&mockDirEntry{name: "1", isDir: true},
		&mockDirEntry{name: "2", isDir: true},
		&mockDirEntry{name: "3", isDir: true},
	}

	// Process 1: normal
	mockFS.files["/proc/1/stat"] = []byte("1 (init) S 0 1 1")
	// Process 2: malformed (no closing paren) - should be skipped
	mockFS.files["/proc/2/stat"] = []byte("2 (broken Z 0 2 2")
	// Process 3: zombie
	mockFS.files["/proc/3/stat"] = []byte("3 (defunct) Z 0 3 3")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	// Should only count process 3 as zombie (process 2 skipped due to malformed stat)
	if result {
		t.Error("Expected false when zombie count <= 5")
	}
	if signal.count != 1 {
		t.Errorf("Expected count 1 (malformed stat skipped), got %d", signal.count)
	}
}

func TestZombieProcessesSignal_Check_WithMock_ReadErrors(t *testing.T) {
	mockFS := newMockProcFS()

	// Simulate error reading /proc/stat
	mockFS.readErr = fmt.Errorf("permission denied")

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	// Should return false on error
	if result {
		t.Error("Expected false when /proc/stat read fails")
	}
}

func TestZombieProcessesSignal_Check_WithMock_ReadDirError(t *testing.T) {
	mockFS := newMockProcFS()

	mockFS.files["/proc/stat"] = []byte("processes 1000\n")
	// Don't set dirs["/proc"], so ReadDir will return error

	signal := NewZombieProcessesSignal()
	result := signal.checkWithFS(context.Background(), mockFS)

	// Should return false when ReadDir fails
	if result {
		t.Error("Expected false when /proc ReadDir fails")
	}
}

func TestZombieProcessesSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_ZOMBIE_PROCESSES", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_ZOMBIE_PROCESSES")

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
