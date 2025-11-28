package signals

import (
	"context"
	"os"
	"runtime"
)

// DockerSocketSignal checks Docker socket permissions
type DockerSocketSignal struct {
	perms os.FileMode
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
	return "Docker socket has overly permissive permissions"
}

func (s *DockerSocketSignal) Remediation() string {
	return "Restrict Docker socket access to docker group only"
}

func (s *DockerSocketSignal) Check(ctx context.Context) bool {
	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}
	
	socketPath := "/var/run/docker.sock"
	info, err := os.Stat(socketPath)
	if err != nil {
		return false // Docker socket doesn't exist
	}
	
	perms := info.Mode().Perm()
	s.perms = perms
	
	// Check if world-writable (others have write permission)
	if perms&0002 != 0 {
		return true
	}
	
	// Also flag if world-readable (less critical but still a concern)
	if perms&0004 != 0 {
		return true
	}
	
	return false
}

