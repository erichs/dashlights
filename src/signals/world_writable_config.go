package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
)

// WorldWritableConfigSignal checks for world-writable shell RC files
type WorldWritableConfigSignal struct {
	foundFiles []string
}

// NewWorldWritableConfigSignal creates a WorldWritableConfigSignal.
func NewWorldWritableConfigSignal() *WorldWritableConfigSignal {
	return &WorldWritableConfigSignal{}
}

// Name returns the human-readable name of the signal.
func (s *WorldWritableConfigSignal) Name() string {
	return "World Writable Config"
}

// Emoji returns the emoji associated with the signal.
func (s *WorldWritableConfigSignal) Emoji() string {
	return "🖊️"
}

// Diagnostic returns a description of the detected world-writable configuration files.
func (s *WorldWritableConfigSignal) Diagnostic() string {
	if len(s.foundFiles) == 0 {
		return "World-writable config files detected"
	}
	return "World-writable: " + s.foundFiles[0]
}

// Remediation returns guidance on how to correct world-writable configuration files.
func (s *WorldWritableConfigSignal) Remediation() string {
	return "Fix permissions with: chmod 644 <file>"
}

// Check inspects common shell RC files for world-writable permissions.
func (s *WorldWritableConfigSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check common shell RC files
	filesToCheck := []string{
		".bashrc",
		".zshrc",
		".profile",
		".bash_profile",
		".zprofile",
	}

	s.foundFiles = []string{}

	for _, file := range filesToCheck {
		fullPath := filepath.Join(homeDir, file)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // File doesn't exist
		}

		perms := info.Mode().Perm()

		// Check if world-writable (others have write permission)
		if perms&0002 != 0 {
			s.foundFiles = append(s.foundFiles, file)
		}
	}

	return len(s.foundFiles) > 0
}
