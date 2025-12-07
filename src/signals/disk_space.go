package signals

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// DiskSpaceSignal checks for critical disk space
type DiskSpaceSignal struct {
	path        string
	percentUsed int
}

// NewDiskSpaceSignal creates a DiskSpaceSignal.
func NewDiskSpaceSignal() *DiskSpaceSignal {
	return &DiskSpaceSignal{}
}

// Name returns the human-readable name of the signal.
func (s *DiskSpaceSignal) Name() string {
	return "Full Tank"
}

// Emoji returns the emoji associated with the signal.
func (s *DiskSpaceSignal) Emoji() string {
	return "💾"
}

// Diagnostic returns a description of the detected disk space issue.
func (s *DiskSpaceSignal) Diagnostic() string {
	return fmt.Sprintf("%s is %d%% full", s.path, s.percentUsed)
}

// Remediation returns guidance on how to resolve the disk space issue.
func (s *DiskSpaceSignal) Remediation() string {
	return "Free up disk space to prevent logging and audit trail failures"
}

// Check evaluates disk usage and reports whether disk space is critically low.
func (s *DiskSpaceSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_DISK_SPACE") != "" {
		return false
	}

	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}

	// Check root volume
	if s.checkPath("/") {
		return true
	}

	return false
}

func (s *DiskSpaceSignal) checkPath(path string) bool {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return false
	}

	// Validate Bsize is positive before conversion to prevent integer overflow
	if stat.Bsize <= 0 {
		return false
	}

	// Safe conversion: we've validated Bsize is positive
	blockSize := uint64(stat.Bsize)

	// Calculate percentage used
	total := stat.Blocks * blockSize
	free := stat.Bfree * blockSize
	used := total - free

	if total == 0 {
		return false
	}

	// Calculate percentage (result will be 0-100, safe for int conversion)
	percentUsed64 := (used * 100) / total

	// Validate the value is within int range before conversion (G115)
	// For a percentage, this should always be 0-100, but we validate to be safe
	if percentUsed64 > 100 {
		// Cap at 100% if somehow we get a value over 100
		s.path = path
		s.percentUsed = 100
		return true
	}

	// Safe conversion: percentUsed64 is guaranteed to be 0-100
	percentUsed := int(percentUsed64)

	if percentUsed > 90 {
		s.path = path
		s.percentUsed = percentUsed
		return true
	}

	return false
}
