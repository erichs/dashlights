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

func NewWorldWritableConfigSignal() *WorldWritableConfigSignal {
	return &WorldWritableConfigSignal{}
}

func (s *WorldWritableConfigSignal) Name() string {
	return "World Writable Config"
}

func (s *WorldWritableConfigSignal) Emoji() string {
	return "🖊️"
}

func (s *WorldWritableConfigSignal) Diagnostic() string {
	if len(s.foundFiles) == 0 {
		return "World-writable config files detected"
	}
	return "World-writable: " + s.foundFiles[0]
}

func (s *WorldWritableConfigSignal) Remediation() string {
	return "Fix permissions with: chmod 644 <file>"
}

func (s *WorldWritableConfigSignal) Check(ctx context.Context) bool {
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

