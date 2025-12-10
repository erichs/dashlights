// Package filestat provides utilities for filesystem operations
// focused on file pattern matching and stat-based detection.
package filestat

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SensitiveFilePatterns defines patterns for detecting sensitive files.
// These are designed for fast name-only matching (no content scanning).
type SensitiveFilePatterns struct {
	// Extensions to match (e.g., ".sql", ".db")
	Extensions []string
	// Prefixes to match (e.g., "dump-", "backup-")
	Prefixes []string
	// Substrings to match anywhere in filename (e.g., "prod")
	Substrings []string
}

// DefaultSensitivePatterns returns patterns for common sensitive file types.
func DefaultSensitivePatterns() SensitiveFilePatterns {
	return SensitiveFilePatterns{
		// Database files and backups
		Extensions: []string{
			".sql",      // SQL dumps
			".sqlite",   // SQLite databases
			".db",       // Generic database files
			".bak",      // Backup files
			".har",      // HTTP Archive (contains full requests/responses)
			".pcap",     // Packet captures
			".keychain", // macOS keychain files
			".pem",      // PEM certificates/keys
			".pfx",      // PKCS#12 files
			".jks",      // Java KeyStores
		},
		Prefixes: []string{
			"dump-",   // Common dump prefix
			"backup-", // Common backup prefix
		},
		Substrings: []string{
			"prod", // Production data indicators
		},
	}
}

// MatchResult contains information about a matched file.
type MatchResult struct {
	Path    string
	ModTime time.Time
	Size    int64
}

// MatchFile checks if a filename matches the sensitive file patterns.
// This performs name-only matching for performance.
func (p *SensitiveFilePatterns) MatchFile(name string) bool {
	lowerName := strings.ToLower(name)

	// Check extensions
	for _, ext := range p.Extensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range p.Prefixes {
		if strings.HasPrefix(lowerName, prefix) {
			return true
		}
	}

	// Check substrings
	for _, substr := range p.Substrings {
		if strings.Contains(lowerName, substr) {
			return true
		}
	}

	return false
}

// ScanDirectory scans a directory for files matching the patterns.
// It returns only regular files (no directories, symlinks, etc.).
// This function is shallow - it does not recurse into subdirectories.
func (p *SensitiveFilePatterns) ScanDirectory(dirPath string) ([]MatchResult, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var results []MatchResult
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check if name matches patterns
		if !p.MatchFile(entry.Name()) {
			continue
		}

		// Get file info for modification time
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Skip non-regular files (symlinks, devices, etc.)
		if !info.Mode().IsRegular() {
			continue
		}

		results = append(results, MatchResult{
			Path:    filepath.Join(dirPath, entry.Name()),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	return results, nil
}

// IsOlderThan checks if a file's modification time is older than the threshold.
func (m *MatchResult) IsOlderThan(threshold time.Duration) bool {
	return time.Since(m.ModTime) > threshold
}

// GetHotZoneDirectories returns the standard directories to scan for data sprawl.
// These are "junk drawer" locations where sensitive data often accumulates.
func GetHotZoneDirectories() []string {
	var dirs []string

	// Get home directory
	home, err := os.UserHomeDir()
	if err == nil {
		// ~/Downloads
		dirs = append(dirs, filepath.Join(home, "Downloads"))
		// ~/Desktop
		dirs = append(dirs, filepath.Join(home, "Desktop"))
	}

	// Current working directory
	cwd, err := os.Getwd()
	if err == nil {
		dirs = append(dirs, cwd)
	}

	// /tmp
	dirs = append(dirs, "/tmp")

	return dirs
}
