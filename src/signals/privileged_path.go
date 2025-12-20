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

// NewPrivilegedPathSignal creates a PrivilegedPathSignal.
func NewPrivilegedPathSignal() *PrivilegedPathSignal {
	return &PrivilegedPathSignal{}
}

// Name returns the human-readable name of the signal.
func (s *PrivilegedPathSignal) Name() string {
	return "Privileged Path"
}

// Emoji returns the emoji associated with the signal.
func (s *PrivilegedPathSignal) Emoji() string {
	return "💣"
}

// Diagnostic returns a description of the detected PATH issues.
func (s *PrivilegedPathSignal) Diagnostic() string {
	if len(s.findings) == 0 {
		return "Potentially dangerous entries detected in PATH"
	}

	if len(s.findings) == 1 {
		return s.findings[0]
	}

	return "Multiple PATH issues detected: " + strings.Join(s.findings, "; ")
}

// Remediation returns guidance on how to harden the PATH configuration.
func (s *PrivilegedPathSignal) Remediation() string {
	return "Remove '.' and world-writable or user bin directories from PATH, or move user bin directories after system paths like /usr/bin"
}

// Check analyzes PATH for unsafe entries and ordering.
func (s *PrivilegedPathSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_PRIVILEGED_PATH") != "" {
		return false
	}

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
		select {
		case <-ctx.Done():
			return false
		default:
		}

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

		// Track if this entry was already reported as world-writable to avoid
		// emitting a second, redundant finding for the same path as a user bin.
		worldWritable := false

		// World-writable PATH entries: any directory with the others-write bit set
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			perm := fi.Mode().Perm()
			if perm&0o002 != 0 {
				worldWritable = true
				// If this is also a known user bin directory that appears before
				// system paths, report a single combined finding instead of two
				// separate messages.
				if earliestSystemIdx != -1 && i < earliestSystemIdx {
					if label, ok := userBinDirs[p]; ok {
						s.findings = append(s.findings,
							fmt.Sprintf("World-writable user PATH directory %s appears before system directories: %s (mode %04o)", label, p, perm))
						continue
					}
				}
				// Fallback: generic world-writable PATH entry message.
				s.findings = append(s.findings,
					fmt.Sprintf("World-writable PATH entry: %s (mode %04o)", p, perm))
			}
		}

		// User-writable PATH entries that precede system directories: common user
		// bin directories (e.g., $HOME/bin, $GOPATH/bin, ~/.cargo/bin) appearing
		// before /usr/bin, /sbin, or /bin. If we've already emitted a
		// world-writable finding for this path, skip adding a second message.
		if !worldWritable && earliestSystemIdx != -1 && i < earliestSystemIdx {
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
		result[filepath.Join(home, ".cargo", "bin")] = "$HOME/.cargo/bin"
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
	for _, sp := range trustedSystemPaths {
		if path == sp {
			return true
		}
	}
	return false
}

// trustedSystemPaths is the ordered list of system paths that should appear
// at the beginning of PATH. Order matters: more specific paths first.
var trustedSystemPaths = []string{
	"/usr/local/bin",
	"/usr/local/sbin",
	"/usr/bin",
	"/usr/sbin",
	"/bin",
	"/sbin",
}

// SuggestCorrectedPath generates a corrected PATH string with trusted system
// paths first, followed by user paths. Dangerous entries like '.' and empty
// segments are removed. Duplicates are eliminated (keeping the first occurrence
// in the corrected order).
func SuggestCorrectedPath(currentPath string) string {
	if currentPath == "" {
		return strings.Join(trustedSystemPaths, ":")
	}

	paths := strings.Split(currentPath, string(os.PathListSeparator))

	// Track seen paths to eliminate duplicates
	seen := make(map[string]bool)

	// Collect system paths that exist in current PATH (preserving trustedSystemPaths order)
	var systemPaths []string
	for _, sp := range trustedSystemPaths {
		for _, p := range paths {
			if p == sp && !seen[sp] {
				systemPaths = append(systemPaths, sp)
				seen[sp] = true
				break
			}
		}
	}

	// Collect non-system paths, excluding dangerous entries
	var userPaths []string
	for _, p := range paths {
		// Skip empty entries (:: implies current directory)
		if p == "" {
			continue
		}
		// Skip current directory
		if p == "." {
			continue
		}
		// Skip if already seen (system path or duplicate)
		if seen[p] {
			continue
		}
		// Skip system paths (already handled above)
		if isSystemPath(p) {
			continue
		}
		userPaths = append(userPaths, p)
		seen[p] = true
	}

	// Combine: system paths first, then user paths
	result := append(systemPaths, userPaths...)
	return strings.Join(result, ":")
}

// VerboseRemediation returns a ready-to-use export PATH command that fixes
// the detected PATH ordering issues.
func (s *PrivilegedPathSignal) VerboseRemediation() string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return ""
	}

	correctedPath := SuggestCorrectedPath(pathEnv)
	if correctedPath == "" {
		return ""
	}

	return fmt.Sprintf("Run this command to fix your PATH:\n\n   export PATH=\"%s\"\n\nTo make this permanent, add the export command to your shell configuration file (e.g., ~/.zshrc or ~/.bashrc).", correctedPath)
}
