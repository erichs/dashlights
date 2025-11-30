package signals

import (
	"context"
	"os"
	"runtime"
	"strings"
)

// DockerSocketSignal checks Docker socket permissions and orphaned DOCKER_HOST
type DockerSocketSignal struct {
	perms        os.FileMode
	issue        string // "permissions" or "orphaned"
	orphanedPath string
	socketPath   string // actual socket path checked
}

func NewDockerSocketSignal() *DockerSocketSignal {
	return &DockerSocketSignal{}
}

func (s *DockerSocketSignal) Name() string {
	return "Exposed Socket"
}

func (s *DockerSocketSignal) Emoji() string {
	return "🐳"
}

func (s *DockerSocketSignal) Diagnostic() string {
	if s.issue == "orphaned" {
		return "DOCKER_HOST points to non-existent socket: " + s.orphanedPath
	}
	return "Docker socket has overly permissive permissions"
}

func (s *DockerSocketSignal) Remediation() string {
	if s.issue == "orphaned" {
		return "Unset DOCKER_HOST or fix the socket path (Docker commands will hang)"
	}
	if runtime.GOOS == "darwin" {
		return "Docker Desktop on macOS: socket permissions are managed automatically, no action needed"
	}
	return "Restrict Docker socket access to docker group only"
}

func (s *DockerSocketSignal) Check(ctx context.Context) bool {
	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}

	// First check if DOCKER_HOST is set to a socket path
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost != "" {
		// Check if it's a Unix socket (starts with unix://)
		socketPath := ""
		if strings.HasPrefix(dockerHost, "unix://") {
			socketPath = strings.TrimPrefix(dockerHost, "unix://")
		} else if strings.HasPrefix(dockerHost, "/") {
			// Direct path without unix:// prefix
			socketPath = dockerHost
		}

		// If it's a socket path, check if it exists
		if socketPath != "" {
			if _, err := os.Stat(socketPath); os.IsNotExist(err) {
				s.issue = "orphaned"
				s.orphanedPath = socketPath
				return true
			}
		}
	}

	// Check default Docker socket permissions
	socketPath := "/var/run/docker.sock"
	info, err := os.Lstat(socketPath) // Use Lstat to check symlink itself
	if err != nil {
		return false // Docker socket doesn't exist
	}

	s.socketPath = socketPath

	// On macOS with Docker Desktop, the socket is typically a symlink
	// to the user's ~/.docker/run/docker.sock
	if runtime.GOOS == "darwin" {
		return s.checkDarwinSocket(info, socketPath)
	}

	// Linux: check socket permissions directly
	return s.checkLinuxSocket(info)
}

// checkDarwinSocket handles Docker Desktop on macOS
func (s *DockerSocketSignal) checkDarwinSocket(info os.FileInfo, socketPath string) bool {
	// On macOS, Docker Desktop creates a symlink from /var/run/docker.sock
	// to ~/.docker/run/docker.sock (owned by the user)
	if info.Mode()&os.ModeSymlink != 0 {
		// Follow the symlink to check the actual socket
		targetPath, err := os.Readlink(socketPath)
		if err != nil {
			// Can't read symlink, skip check
			return false
		}

		// Check if it points to a user's Docker Desktop socket
		// Typical pattern: /Users/<user>/.docker/run/docker.sock
		if strings.Contains(targetPath, "/.docker/run/docker.sock") {
			// This is Docker Desktop - check the actual socket permissions
			targetInfo, err := os.Stat(targetPath)
			if err != nil {
				// Target doesn't exist - this is an orphaned symlink
				s.issue = "orphaned"
				s.orphanedPath = targetPath
				return true
			}

			// Check the actual socket permissions
			// Docker Desktop sets this to user-owned with 0755 (srwxr-xr-x)
			// This is safe because it's in the user's home directory
			perms := targetInfo.Mode().Perm()
			s.perms = perms

			// On macOS Docker Desktop, world-readable is acceptable
			// because the socket is in the user's home directory
			// Only flag if world-writable
			if perms&0002 != 0 {
				s.issue = "permissions"
				return true
			}

			return false
		}
	}

	// Not a Docker Desktop symlink, check as regular socket
	return s.checkLinuxSocket(info)
}

// checkLinuxSocket checks socket permissions on Linux
func (s *DockerSocketSignal) checkLinuxSocket(info os.FileInfo) bool {
	perms := info.Mode().Perm()
	s.perms = perms

	// Check if world-writable (others have write permission)
	if perms&0002 != 0 {
		s.issue = "permissions"
		return true
	}

	// Also flag if world-readable (less critical but still a concern)
	if perms&0004 != 0 {
		s.issue = "permissions"
		return true
	}

	return false
}
