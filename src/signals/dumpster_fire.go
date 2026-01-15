package signals

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erichs/dashlights/src/signals/internal/filestat"
)

// dumpsterFireBudget is the total time budget for scanning all hot zones.
// With parallel scanning, this is wall-clock time, not cumulative.
const dumpsterFireBudget = 8 * time.Millisecond

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

// dirScanResult holds results from scanning a single directory.
type dirScanResult struct {
	dir    string
	result filestat.ScanResult
	err    error
}

// Check scans hot-zone directories for sensitive-looking files.
// Directories are scanned in parallel with a global 8ms time budget.
// This is adaptive: fast systems scan more entries, slow systems scan fewer.
func (s *DumpsterFireSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_DUMPSTER_FIRE") != "" {
		return false
	}

	s.totalCount = 0
	s.dirCounts = make(map[string]int)
	s.foundPaths = nil

	// Create a time-budgeted context for all parallel scans
	scanCtx, cancel := context.WithTimeout(ctx, dumpsterFireBudget)
	defer cancel()

	patterns := filestat.DefaultSensitivePatterns()
	dirs := filestat.GetHotZoneDirectories()

	// Use time-based config: no entry limits, just the context deadline
	config := filestat.ScanConfig{
		MaxMatches: 10, // Still cap matches per directory (we've proven the point)
		MaxEntries: 0,  // No entry limit - use time budget instead
		Timeout:    0,  // No per-dir timeout - use global budget via context
	}

	// Launch parallel scans for all directories
	resultCh := make(chan dirScanResult, len(dirs))

	for _, dir := range dirs {
		go func(d string) {
			// Skip directories that don't exist
			if _, err := os.Stat(d); os.IsNotExist(err) {
				resultCh <- dirScanResult{d, filestat.ScanResult{}, err}
				return
			}

			result, err := patterns.ScanDirectory(scanCtx, d, config)
			resultCh <- dirScanResult{d, result, err}
		}(dir)
	}

	// Track unique files to avoid double-counting when $PWD overlaps with other dirs
	seenPaths := make(map[string]bool)

	// Collect results from all goroutines (with context timeout)
	for i := 0; i < len(dirs); i++ {
		select {
		case r := <-resultCh:
			if r.err != nil {
				continue // Skip directories we can't read
			}

			for _, match := range r.result.Matches {
				// Deduplicate paths (in case $PWD is ~/Downloads, etc.)
				if seenPaths[match.Path] {
					continue
				}
				seenPaths[match.Path] = true

				s.dirCounts[r.dir]++
				s.totalCount++
				s.foundPaths = append(s.foundPaths, match.Path)
			}
		case <-scanCtx.Done():
			// Time budget exhausted - return what we have so far
			return s.totalCount > 0
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
