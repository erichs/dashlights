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
	info, err := os.Stat(socketPath)
	if err != nil {
		return false // Docker socket doesn't exist
	}

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
