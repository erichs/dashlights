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

func NewRootOwnedHomeSignal() *RootOwnedHomeSignal {
	return &RootOwnedHomeSignal{}
}

func (s *RootOwnedHomeSignal) Name() string {
	return "Root Squatter"
}

func (s *RootOwnedHomeSignal) Emoji() string {
	return "🍄"
}

func (s *RootOwnedHomeSignal) Diagnostic() string {
	if len(s.foundFiles) == 0 {
		return "Root-owned files found in home directory"
	}
	return "Root-owned: " + s.foundFiles[0]
}

func (s *RootOwnedHomeSignal) Remediation() string {
	return "Fix ownership with: sudo chown -R $USER:$USER <file>"
}

func (s *RootOwnedHomeSignal) Check(ctx context.Context) bool {
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

