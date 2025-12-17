package signals

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erichs/dashlights/src/signals/internal/filestat"
)

// RottingSecretsAgeThreshold is the default age threshold for "rotting" files.
// Files older than this are considered forgotten high-value artifacts.
const RottingSecretsAgeThreshold = 7 * 24 * time.Hour // 7 days

// RottingSecretsSignal detects long-lived sensitive files that may have been forgotten.
// It identifies the same files as DumpsterFireSignal but only flags those older than
// the age threshold. These are likely "forgotten" sensitive files that should be cleaned up.
type RottingSecretsSignal struct {
	count      int
	oldestAge  time.Duration
	foundPaths []string // Store paths for verbose remediation
}

// NewRottingSecretsSignal creates a RottingSecretsSignal.
func NewRottingSecretsSignal() *RottingSecretsSignal {
	return &RottingSecretsSignal{}
}

// Name returns the human-readable name of the signal.
func (s *RottingSecretsSignal) Name() string {
	return "Rotting Secrets"
}

// Emoji returns the emoji associated with the signal.
func (s *RottingSecretsSignal) Emoji() string {
	return "🦴" // Bone emoji - represents old/decaying artifacts
}

// Diagnostic returns a description of detected old sensitive files.
func (s *RottingSecretsSignal) Diagnostic() string {
	if s.count == 0 {
		return "Old sensitive files detected in common directories"
	}
	days := int(s.oldestAge.Hours() / 24)
	return fmt.Sprintf("%d sensitive file(s) older than 7 days (oldest: %d days)", s.count, days)
}

// Remediation returns guidance on handling old sensitive files.
func (s *RottingSecretsSignal) Remediation() string {
	return "Clean up old database dumps, logs, and key files - or move to secure storage"
}

// Check scans hot-zone directories for old sensitive files.
func (s *RottingSecretsSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_ROTTING_SECRETS") != "" {
		return false
	}

	s.count = 0
	s.oldestAge = 0
	s.foundPaths = nil

	patterns := filestat.DefaultSensitivePatterns()
	dirs := filestat.GetHotZoneDirectories()
	config := filestat.DefaultScanConfig()

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

		result, err := patterns.ScanDirectory(ctx, dir, config)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, match := range result.Matches {
			// Deduplicate paths
			if seenPaths[match.Path] {
				continue
			}
			seenPaths[match.Path] = true

			// Only count files older than threshold
			if !match.IsOlderThan(RottingSecretsAgeThreshold) {
				continue
			}

			s.count++
			s.foundPaths = append(s.foundPaths, match.Path)

			// Track oldest file age
			age := time.Since(match.ModTime)
			if age > s.oldestAge {
				s.oldestAge = age
			}
		}
	}

	return s.count > 0
}

// Count returns the number of old sensitive files found.
func (s *RottingSecretsSignal) Count() int {
	return s.count
}

// OldestAge returns the age of the oldest sensitive file found.
func (s *RottingSecretsSignal) OldestAge() time.Duration {
	return s.oldestAge
}

// VerboseRemediation returns specific rm commands for the detected old files.
func (s *RottingSecretsSignal) VerboseRemediation() string {
	if len(s.foundPaths) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("These files are over 7 days old - review and remove:\n\n")

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
