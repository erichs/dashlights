package signals

import (
	"context"
	"os"
	"runtime"
	"strings"
)

// PrivilegedPathSignal checks for '.' in PATH
type PrivilegedPathSignal struct {
	diagnostic string
}

func NewPrivilegedPathSignal() *PrivilegedPathSignal {
	return &PrivilegedPathSignal{}
}

func (s *PrivilegedPathSignal) Name() string {
	return "Privileged Path"
}

func (s *PrivilegedPathSignal) Emoji() string {
	return "💣"
}

func (s *PrivilegedPathSignal) Diagnostic() string {
	return s.diagnostic
}

func (s *PrivilegedPathSignal) Remediation() string {
	return "Remove '.' from PATH or move it to the end after system directories"
}

func (s *PrivilegedPathSignal) Check(ctx context.Context) bool {
	if runtime.GOOS == "windows" {
		return false
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return false
	}

	pathSep := ":"
	paths := strings.Split(pathEnv, pathSep)

	for i, p := range paths {
		// Check for explicit '.' or empty string (which means current directory)
		if p == "." {
			s.diagnostic = "Current directory '.' found in PATH"
			return true
		}

		// Check for empty string between colons (::)
		if p == "" {
			s.diagnostic = "Empty path entry (::) found in PATH (implies current directory)"
			return true
		}

		// Extra dangerous: '.' before system directories
		if p == "." && i < len(paths)-1 {
			// Check if any subsequent path is a system directory
			for j := i + 1; j < len(paths); j++ {
				if isSystemPath(paths[j]) {
					s.diagnostic = "Current directory '.' in PATH before system directories"
					return true
				}
			}
		}
	}

	return false
}

func isSystemPath(path string) bool {
	systemPaths := []string{
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}

	for _, sp := range systemPaths {
		if path == sp {
			return true
		}
	}
	return false
}
