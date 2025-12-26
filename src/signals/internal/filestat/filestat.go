// Package filestat provides utilities for filesystem operations
// focused on file pattern matching and stat-based detection.
package filestat

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Performance limits for sensitive file scanning.
// These cap worst-case behavior when scanning large directories.
const (
	// maxMatchesPerDir limits how many files to stat per directory.
	// After this many matches, we've proven the directory has issues.
	maxMatchesPerDir = 10

	// maxEntriesPerDir limits directory entries to process before giving up.
	// Handles pathological cases like /tmp with thousands of files.
	maxEntriesPerDir = 500

	// perDirTimeout is the maximum time budget for scanning a single directory.
	// With 4 hot zones, allows ~8ms total leaving 2ms buffer for 10ms budget.
	perDirTimeout = 2 * time.Millisecond
)

// ScanConfig contains configuration for directory scanning.
type ScanConfig struct {
	MaxMatches int           // Max matches to return (0 = unlimited)
	MaxEntries int           // Max entries to process (0 = unlimited)
	Timeout    time.Duration // Per-directory timeout (0 = no timeout)
}

// DefaultScanConfig returns conservative limits, future work may make this user-configurable
func DefaultScanConfig() ScanConfig {
	return ScanConfig{
		MaxMatches: maxMatchesPerDir,
		MaxEntries: maxEntriesPerDir,
		Timeout:    perDirTimeout,
	}
}

// ScanResult contains scan results and metadata.
type ScanResult struct {
	Matches   []MatchResult
	Truncated bool   // True if scan was limited by MaxMatches/MaxEntries/Timeout
	Reason    string // Why scan was truncated: "max_matches", "max_entries", or "timeout"
}

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
			"prod",       // Production data indicators (matched at word boundaries)
			"production", // Full word also matches (e.g., "production-dump.sql")
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

	// Check substrings with word-boundary awareness
	for _, substr := range p.Substrings {
		if containsAtWordBoundary(lowerName, substr) {
			return true
		}
	}

	return false
}

// containsAtWordBoundary checks if substr appears in s at word boundaries.
// Word boundaries are: start/end of string, or common delimiters (-_. and space).
// This prevents "prod" from matching "product", "produce", etc.
func containsAtWordBoundary(s, substr string) bool {
	idx := 0
	for {
		pos := strings.Index(s[idx:], substr)
		if pos == -1 {
			return false
		}
		pos += idx // Adjust to absolute position

		// Check if at word boundary
		atStart := pos == 0 || isDelimiter(s[pos-1])
		endPos := pos + len(substr)
		atEnd := endPos == len(s) || isDelimiter(s[endPos])

		if atStart && atEnd {
			return true
		}

		// Move past this occurrence and keep searching
		idx = pos + 1
		if idx >= len(s) {
			return false
		}
	}
}

// isDelimiter returns true if the byte is a common filename delimiter.
func isDelimiter(b byte) bool {
	return b == '-' || b == '_' || b == '.' || b == ' '
}

// ScanDirectory scans a directory for files matching the patterns.
// It returns only regular files (no directories, symlinks, etc.).
// This function is shallow - it does not recurse into subdirectories.
// It accepts a context for cancellation and a ScanConfig for limits.
func (p *SensitiveFilePatterns) ScanDirectory(ctx context.Context, dirPath string, config ScanConfig) (ScanResult, error) {
	result := ScanResult{}

	// Check context before starting
	select {
	case <-ctx.Done():
		result.Truncated = true
		result.Reason = "timeout"
		return result, nil
	default:
	}

	// Set up per-directory timeout if configured
	var scanCtx context.Context
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		scanCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	} else {
		scanCtx = ctx
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return result, err
	}

	entriesProcessed := 0

	for _, entry := range entries {
		// Check context cancellation
		select {
		case <-scanCtx.Done():
			result.Truncated = true
			result.Reason = "timeout"
			return result, nil
		default:
		}

		// Check entries limit
		if config.MaxEntries > 0 && entriesProcessed >= config.MaxEntries {
			result.Truncated = true
			result.Reason = "max_entries"
			return result, nil
		}
		entriesProcessed++

		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check if name matches patterns (fast, no syscall)
		if !p.MatchFile(entry.Name()) {
			continue
		}

		// Check matches limit BEFORE expensive stat
		if config.MaxMatches > 0 && len(result.Matches) >= config.MaxMatches {
			result.Truncated = true
			result.Reason = "max_matches"
			return result, nil
		}

		// Get file info for modification time (expensive stat syscall)
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Skip non-regular files (symlinks, devices, etc.)
		if !info.Mode().IsRegular() {
			continue
		}

		result.Matches = append(result.Matches, MatchResult{
			Path:    filepath.Join(dirPath, entry.Name()),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	return result, nil
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
