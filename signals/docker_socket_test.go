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

