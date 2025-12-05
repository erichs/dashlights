package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

// RootOwnedHomeSignal checks for root-owned files in home directory
type RootOwnedHomeSignal struct {
	foundFiles []string
}

// NewRootOwnedHomeSignal creates a RootOwnedHomeSignal.
func NewRootOwnedHomeSignal() *RootOwnedHomeSignal {
	return &RootOwnedHomeSignal{}
}

// Name returns the human-readable name of the signal.
func (s *RootOwnedHomeSignal) Name() string {
	return "Root Squatter"
}

// Emoji returns the emoji associated with the signal.
func (s *RootOwnedHomeSignal) Emoji() string {
	return "🍄"
}

// Diagnostic returns a description of the detected root-owned files.
func (s *RootOwnedHomeSignal) Diagnostic() string {
	if len(s.foundFiles) == 0 {
		return "Root-owned files found in home directory"
	}
	return "Root-owned: " + s.foundFiles[0]
}

// Remediation returns guidance on how to fix root-owned files in the home directory.
func (s *RootOwnedHomeSignal) Remediation() string {
	return "Fix ownership with: sudo chown -R $USER:$USER <file>"
}

// Check inspects common home directory paths for files owned by root.
func (s *RootOwnedHomeSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check common config files
	filesToCheck := []string{
		".bashrc",
		".zshrc",
		".profile",
		".bash_profile",
		".config",
		".ssh",
	}

	s.foundFiles = []string{}

	for _, file := range filesToCheck {
		fullPath := filepath.Join(homeDir, file)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // File doesn't exist
		}

		// Get file ownership
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}

		// Check if owned by root (UID 0)
		if stat.Uid == 0 {
			s.foundFiles = append(s.foundFiles, file)
		}
	}

	return len(s.foundFiles) > 0
}
