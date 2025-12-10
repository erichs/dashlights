package signals

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestDockerSocketSignal_NoDockerSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Ensure DOCKER_HOST is not set
	originalHost := os.Getenv("DOCKER_HOST")
	os.Unsetenv("DOCKER_HOST")
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	// This test assumes /var/run/docker.sock doesn't exist or has proper permissions
	// We can't easily test this without Docker installed
	result := signal.Check(ctx)

	// Just verify it doesn't crash
	_ = result
}

func TestDockerSocketSignal_OrphanedDockerHost_UnixPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Set DOCKER_HOST to a non-existent socket with unix:// prefix
	originalHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "unix:///tmp/nonexistent_docker_socket_12345.sock")
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when DOCKER_HOST points to non-existent socket")
	}

	if signal.issue != "orphaned" {
		t.Errorf("Expected issue='orphaned', got '%s'", signal.issue)
	}

	if signal.orphanedPath != "/tmp/nonexistent_docker_socket_12345.sock" {
		t.Errorf("Expected orphanedPath='/tmp/nonexistent_docker_socket_12345.sock', got '%s'", signal.orphanedPath)
	}

	diagnostic := signal.Diagnostic()
	if diagnostic == "" {
		t.Error("Expected non-empty diagnostic message")
	}

	remediation := signal.Remediation()
	if remediation == "" {
		t.Error("Expected non-empty remediation message")
	}
}

func TestDockerSocketSignal_OrphanedDockerHost_DirectPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Set DOCKER_HOST to a non-existent socket without unix:// prefix
	originalHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "/tmp/nonexistent_docker_socket_67890.sock")
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when DOCKER_HOST points to non-existent socket (direct path)")
	}

	if signal.issue != "orphaned" {
		t.Errorf("Expected issue='orphaned', got '%s'", signal.issue)
	}
}

func TestDockerSocketSignal_ValidDockerHost(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temporary socket file
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/docker.sock"

	// Create the socket file (just a regular file for testing)
	err := os.WriteFile(socketPath, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create test socket: %v", err)
	}

	// Set DOCKER_HOST to the existing socket
	originalHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "unix://"+socketPath)
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	// Should not trigger orphaned check since socket exists
	// May still trigger if /var/run/docker.sock has bad permissions
	_ = result
}

func TestDockerSocketSignal_TCPDockerHost(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Set DOCKER_HOST to TCP (not a socket path)
	originalHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "tcp://localhost:2375")
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		} else {
			os.Unsetenv("DOCKER_HOST")
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	// Should not trigger orphaned check for TCP connections
	// May still check default socket permissions
	_ = result
}

func TestDockerSocketSignal_MacOSDockerDesktop(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS system")
	}

	// Ensure DOCKER_HOST is not set
	originalHost := os.Getenv("DOCKER_HOST")
	os.Unsetenv("DOCKER_HOST")
	defer func() {
		if originalHost != "" {
			os.Setenv("DOCKER_HOST", originalHost)
		}
	}()

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	// On macOS with Docker Desktop, the socket should be properly configured
	// This test verifies the check doesn't crash and handles symlinks
	result := signal.Check(ctx)

	// If Docker Desktop is installed and running, should not detect issues
	// (unless there's an actual problem)
	_ = result

	// Verify the signal has proper metadata
	if signal.Name() == "" {
		t.Error("Expected non-empty name")
	}
	if signal.Emoji() == "" {
		t.Error("Expected non-empty emoji")
	}
}

func TestDockerSocketSignal_MacOSSymlinkToUserSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a mock Docker Desktop setup
	tmpDir := t.TempDir()

	// Create user's .docker/run directory
	userDockerDir := tmpDir + "/user/.docker/run"
	err := os.MkdirAll(userDockerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user docker dir: %v", err)
	}

	// Create the actual socket file (mock)
	userSocket := userDockerDir + "/docker.sock"
	err = os.WriteFile(userSocket, []byte(""), 0755)
	if err != nil {
		t.Fatalf("Failed to create user socket: %v", err)
	}

	// Create /var/run directory
	varRunDir := tmpDir + "/var/run"
	err = os.MkdirAll(varRunDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create var/run dir: %v", err)
	}

	// Create symlink from /var/run/docker.sock to user socket
	symlinkPath := varRunDir + "/docker.sock"
	err = os.Symlink(userSocket, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Test the checkDarwinSocket function directly
	signal := NewDockerSocketSignal()
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}

	// This should not trigger on macOS because the user socket is safe
	result := signal.checkDarwinSocket(info, symlinkPath)
	if result {
		t.Errorf("Expected false for Docker Desktop symlink with safe permissions, got true. Issue: %s", signal.issue)
	}
}

func TestDockerSocketSignal_MacOSWorldWritableUserSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a mock Docker Desktop setup with world-writable socket
	tmpDir := t.TempDir()

	userDockerDir := tmpDir + "/user/.docker/run"
	err := os.MkdirAll(userDockerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user docker dir: %v", err)
	}

	// Create world-writable socket (dangerous)
	userSocket := userDockerDir + "/docker.sock"
	err = os.WriteFile(userSocket, []byte(""), 0666)
	if err != nil {
		t.Fatalf("Failed to create user socket: %v", err)
	}
	// Explicitly set permissions to ensure world-writable
	err = os.Chmod(userSocket, 0666)
	if err != nil {
		t.Fatalf("Failed to chmod socket: %v", err)
	}

	varRunDir := tmpDir + "/var/run"
	err = os.MkdirAll(varRunDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create var/run dir: %v", err)
	}

	symlinkPath := varRunDir + "/docker.sock"
	err = os.Symlink(userSocket, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	signal := NewDockerSocketSignal()
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}

	// Should trigger because socket is world-writable
	result := signal.checkDarwinSocket(info, symlinkPath)
	if !result {
		t.Error("Expected true for world-writable socket")
	}
	if signal.issue != "permissions" {
		t.Errorf("Expected issue='permissions', got '%s'", signal.issue)
	}
}

func TestDockerSocketSignal_MacOSOrphanedSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a symlink to a non-existent Docker Desktop socket
	tmpDir := t.TempDir()

	varRunDir := tmpDir + "/var/run"
	err := os.MkdirAll(varRunDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create var/run dir: %v", err)
	}

	// Create symlink to non-existent user socket
	symlinkPath := varRunDir + "/docker.sock"
	nonExistentTarget := tmpDir + "/user/.docker/run/docker.sock"
	err = os.Symlink(nonExistentTarget, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	signal := NewDockerSocketSignal()
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}

	// Should NOT trigger - orphaned Docker Desktop symlink just means
	// Docker isn't running, not a misconfiguration. Commands fail fast.
	result := signal.checkDarwinSocket(info, symlinkPath)
	if result {
		t.Error("Expected false for orphaned Docker Desktop symlink (not a real issue)")
	}
}

func TestDockerSocketSignal_LinuxWorldReadable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a mock socket with world-readable permissions
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/docker.sock"

	// Create world-readable socket (0644)
	err := os.WriteFile(socketPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create socket: %v", err)
	}

	signal := NewDockerSocketSignal()
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("Failed to stat socket: %v", err)
	}

	// Should trigger on Linux for world-readable
	result := signal.checkLinuxSocket(info)
	if !result {
		t.Error("Expected true for world-readable socket on Linux")
	}
	if signal.issue != "permissions" {
		t.Errorf("Expected issue='permissions', got '%s'", signal.issue)
	}
}

func TestDockerSocketSignal_LinuxWorldWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a mock socket with world-writable permissions
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/docker.sock"

	// Create world-writable socket (0666)
	err := os.WriteFile(socketPath, []byte(""), 0666)
	if err != nil {
		t.Fatalf("Failed to create socket: %v", err)
	}

	signal := NewDockerSocketSignal()
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("Failed to stat socket: %v", err)
	}

	// Should trigger on Linux for world-writable
	result := signal.checkLinuxSocket(info)
	if !result {
		t.Error("Expected true for world-writable socket on Linux")
	}
	if signal.issue != "permissions" {
		t.Errorf("Expected issue='permissions', got '%s'", signal.issue)
	}
}

func TestDockerSocketSignal_LinuxSecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a mock socket with secure permissions
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/docker.sock"

	// Create socket with secure permissions (0660)
	err := os.WriteFile(socketPath, []byte(""), 0660)
	if err != nil {
		t.Fatalf("Failed to create socket: %v", err)
	}

	signal := NewDockerSocketSignal()
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("Failed to stat socket: %v", err)
	}

	// Should NOT trigger for secure permissions
	result := signal.checkLinuxSocket(info)
	if result {
		t.Errorf("Expected false for secure socket permissions, got true. Issue: %s", signal.issue)
	}
}

func TestDockerSocketSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_DOCKER_SOCKET", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_DOCKER_SOCKET")

	signal := NewDockerSocketSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
