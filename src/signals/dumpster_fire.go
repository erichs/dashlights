package signals

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/filestat"
)

// DumpsterFireSignal detects sensitive-looking files in user "hot zones"
// where data sprawl commonly accumulates: Downloads, Desktop, $PWD, and /tmp.
// This is a coarse-grained check using name-only pattern matching for performance.
type DumpsterFireSignal struct {
	totalCount int
	dirCounts  map[string]int
	foundPaths []string // Store paths for verbose remediation
}

// NewDumpsterFireSignal creates a DumpsterFireSignal.
func NewDumpsterFireSignal() *DumpsterFireSignal {
	return &DumpsterFireSignal{
		dirCounts: make(map[string]int),
	}
}

// Name returns the human-readable name of the signal.
func (s *DumpsterFireSignal) Name() string {
	return "Dumpster Fire"
}

// Emoji returns the emoji associated with the signal.
func (s *DumpsterFireSignal) Emoji() string {
	return "🗑️" // Wastebasket emoji
}

// Diagnostic returns a description of detected sensitive files.
func (s *DumpsterFireSignal) Diagnostic() string {
	if s.totalCount == 0 {
		return "Sensitive files detected in common directories"
	}
	return fmt.Sprintf("%d sensitive file(s) in hot zones (Downloads, Desktop, $PWD, /tmp)", s.totalCount)
}

// Remediation returns guidance on handling sensitive file sprawl.
func (s *DumpsterFireSignal) Remediation() string {
	return "Review and remove/secure database dumps, logs, and key files from these locations"
}

// Check scans hot-zone directories for sensitive-looking files.
func (s *DumpsterFireSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_DUMPSTER_FIRE") != "" {
		return false
	}

	s.totalCount = 0
	s.dirCounts = make(map[string]int)
	s.foundPaths = nil

	patterns := filestat.DefaultSensitivePatterns()
	dirs := filestat.GetHotZoneDirectories()

	// Track unique files to avoid double-counting when $PWD overlaps with other dirs
	seenPaths := make(map[string]bool)

	for _, dir := range dirs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return false
		default:
		}

		// Skip directories that don't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		matches, err := patterns.ScanDirectory(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, match := range matches {
			// Deduplicate paths (in case $PWD is ~/Downloads, etc.)
			if seenPaths[match.Path] {
				continue
			}
			seenPaths[match.Path] = true

			s.dirCounts[dir]++
			s.totalCount++
			s.foundPaths = append(s.foundPaths, match.Path)
		}
	}

	return s.totalCount > 0
}

// GetCounts returns the counts of sensitive files by directory.
// This is useful for detailed reporting.
func (s *DumpsterFireSignal) GetCounts() map[string]int {
	return s.dirCounts
}

// TotalCount returns the total number of sensitive files found.
func (s *DumpsterFireSignal) TotalCount() int {
	return s.totalCount
}

// VerboseRemediation returns specific rm commands for the detected files.
func (s *DumpsterFireSignal) VerboseRemediation() string {
	if len(s.foundPaths) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Review and remove these files:\n\n")

	for _, path := range s.foundPaths {
		sb.WriteString(fmt.Sprintf("   rm %q\n", path))
	}

	// Only show combined command if multiple files
	if len(s.foundPaths) > 1 {
		sb.WriteString("\nOr remove all at once (DANGEROUS - review first!):\n\n   rm")
		for _, path := range s.foundPaths {
			sb.WriteString(fmt.Sprintf(" %q", path))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
