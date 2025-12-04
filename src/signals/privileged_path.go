package signals

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PrivilegedPathSignal checks for dangerous PATH entries including '.',
// world-writable directories, and user-specific bin directories that appear
// before system paths.
type PrivilegedPathSignal struct {
	findings []string
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
	if len(s.findings) == 0 {
		return "Potentially dangerous entries detected in PATH"
	}

	if len(s.findings) == 1 {
		return s.findings[0]
	}

	return "Multiple PATH issues detected: " + strings.Join(s.findings, "; ")
}

func (s *PrivilegedPathSignal) Remediation() string {
	return "Remove '.' and world-writable or user bin directories from PATH, or move user bin directories after system paths like /usr/bin"
}

func (s *PrivilegedPathSignal) Check(ctx context.Context) bool {
	if runtime.GOOS == "windows" {
		return false
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return false
	}

	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	// Reset findings for each check invocation
	s.findings = nil

	// Find the earliest system directory in PATH to detect user bin directories
	// that appear before it.
	earliestSystemIdx := -1
	for i, p := range paths {
		if isSystemPath(p) {
			earliestSystemIdx = i
			break
		}
	}

	userBinDirs := buildUserBinDirMap()

	for i, p := range paths {
		// Check for empty string between separators (::) which implies current directory
		if p == "" {
			msg := "Empty PATH entry (::) found (implies current directory)"
			if earliestSystemIdx != -1 && i < earliestSystemIdx {
				msg = "Empty PATH entry (::) before system directories (implies current directory)"
			}
			s.findings = append(s.findings, msg)
			// No further checks for this entry
			continue
		}

		// Check for explicit '.' (current directory)
		if p == "." {
			msg := "Current directory '.' found in PATH"
			if earliestSystemIdx != -1 && i < earliestSystemIdx {
				msg = "Current directory '.' in PATH before system directories"
			}
			s.findings = append(s.findings, msg)
			// Skip further checks for this entry
			continue
		}

		// World-writable PATH entries: any directory with the others-write bit set
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			perm := fi.Mode().Perm()
			if perm&0o002 != 0 {
				s.findings = append(s.findings,
					fmt.Sprintf("World-writable PATH entry: %s (mode %04o)", p, perm))
			}
		}

		// User-writable PATH entries that precede system directories: common user
		// bin directories (e.g., $HOME/bin, $GOPATH/bin, ~/.cargo/bin) appearing
		// before /usr/bin, /sbin, or /bin.
		if earliestSystemIdx != -1 && i < earliestSystemIdx {
			if label, ok := userBinDirs[p]; ok {
				s.findings = append(s.findings,
					fmt.Sprintf("User PATH directory %s appears before system directories", label))
			}
		}
	}

	return len(s.findings) > 0
}

// buildUserBinDirMap returns a map of absolute user-specific bin directories
// to human-readable labels (e.g., "/home/user/bin" -> "$HOME/bin").
func buildUserBinDirMap() map[string]string {
	result := make(map[string]string)

	home := os.Getenv("HOME")
	if home != "" {
		result[filepath.Join(home, "bin")] = "$HOME/bin"
		result[filepath.Join(home, ".local", "bin")] = "$HOME/.local/bin"
		result[filepath.Join(home, ".cargo", "bin")] = "~/.cargo/bin"
	}

	// GOPATH may contain multiple entries separated by the OS path list separator.
	gopathEnv := os.Getenv("GOPATH")
	var gopaths []string
	if gopathEnv != "" {
		gopaths = strings.Split(gopathEnv, string(os.PathListSeparator))
	} else if home != "" {
		// Default GOPATH when not set explicitly
		gopaths = []string{filepath.Join(home, "go")}
	}

	for _, gp := range gopaths {
		if gp == "" {
			continue
		}
		result[filepath.Join(gp, "bin")] = "$GOPATH/bin"
	}

	cargoHomeEnv := os.Getenv("CARGO_HOME")
	if cargoHomeEnv != "" {
		result[filepath.Join(cargoHomeEnv, "bin")] = "$CARGO_HOME/bin"
	}

	return result
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
