package signals

import (
	"context"
	"os"
	"path/filepath"
)

// DanglingSymlinksSignal detects symlinks in the current directory
// that point to non-existent targets
type DanglingSymlinksSignal struct{}

func NewDanglingSymlinksSignal() Signal {
	return &DanglingSymlinksSignal{}
}

func (s *DanglingSymlinksSignal) Name() string {
	return "Dangling Symlinks"
}

func (s *DanglingSymlinksSignal) Emoji() string {
	return "💔" // Broken heart emoji - represents broken links
}

func (s *DanglingSymlinksSignal) Diagnostic() string {
	return "Current directory contains symlinks pointing to non-existent targets"
}

func (s *DanglingSymlinksSignal) Remediation() string {
	return "Remove or fix broken symlinks in current directory"
}

func (s *DanglingSymlinksSignal) Check(ctx context.Context) bool {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Read directory entries
	entries, err := os.ReadDir(cwd)
	if err != nil {
		return false
	}

	// Check each entry for dangling symlinks
	for _, entry := range entries {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		// Only check symlinks
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}

		// Get full path
		fullPath := filepath.Join(cwd, entry.Name())

		// Try to stat the target (not the symlink itself)
		// If this fails, the symlink is dangling
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return true // Found a dangling symlink
			}
		}
	}

	return false
}
