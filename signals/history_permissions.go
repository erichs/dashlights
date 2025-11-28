package signals

import (
	"context"
	"os"
	"path/filepath"
)

// HistoryPermissionsSignal detects world-readable shell history files
// Shell history often contains sensitive commands, credentials, and secrets
type HistoryPermissionsSignal struct{}

func NewHistoryPermissionsSignal() Signal {
	return &HistoryPermissionsSignal{}
}

func (s *HistoryPermissionsSignal) Name() string {
	return "Shell History World-Readable"
}

func (s *HistoryPermissionsSignal) Emoji() string {
	return "📜" // Scroll emoji (for history)
}

func (s *HistoryPermissionsSignal) Diagnostic() string {
	return "Shell history files are world-readable (other users can read your typed secrets)"
}

func (s *HistoryPermissionsSignal) Remediation() string {
	return "Run: chmod 600 ~/.bash_history ~/.zsh_history"
}

func (s *HistoryPermissionsSignal) Check(ctx context.Context) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check common shell history files
	historyFiles := []string{
		".bash_history",
		".zsh_history",
	}

	for _, histFile := range historyFiles {
		fullPath := filepath.Join(homeDir, histFile)
		
		// Check if file exists
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			// File doesn't exist, skip
			continue
		}

		// Get file permissions
		mode := fileInfo.Mode().Perm()

		// Check if file is readable by group or others (not 600)
		// 0600 = owner read/write only
		// If group or others have any permissions, it's a problem
		if mode&0077 != 0 {
			// Group or others have permissions - this is bad
			return true
		}
	}

	return false
}

