package install

import (
	"fmt"
	"os"
	"time"
)

// BackupResult contains information about a backup operation.
type BackupResult struct {
	OriginalPath string
	BackupPath   string
	Created      bool
}

// BackupManager handles backup file operations.
type BackupManager struct {
	fs Filesystem
}

// NewBackupManager creates a new BackupManager with the given filesystem.
func NewBackupManager(fs Filesystem) *BackupManager {
	return &BackupManager{fs: fs}
}

// CreateBackup creates a backup of the specified file.
// If a backup already exists, a timestamped backup is created.
// Returns BackupResult with Created=false if the file doesn't exist.
func (b *BackupManager) CreateBackup(filePath string) (*BackupResult, error) {
	backupPath := filePath + ".dashlights-backup"

	// If backup already exists, use timestamp
	if b.fs.Exists(backupPath) {
		backupPath = fmt.Sprintf("%s.dashlights-backup-%d", filePath, time.Now().Unix())
	}

	// Get original file info to preserve permissions
	info, err := b.fs.Stat(filePath)
	if os.IsNotExist(err) {
		return &BackupResult{Created: false}, nil // Nothing to backup
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	content, err := b.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Preserve original file mode
	if err := b.fs.WriteFile(backupPath, content, info.Mode()); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	return &BackupResult{
		OriginalPath: filePath,
		BackupPath:   backupPath,
		Created:      true,
	}, nil
}
