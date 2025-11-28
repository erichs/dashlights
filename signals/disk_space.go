package signals

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
)

// DiskSpaceSignal checks for critical disk space
type DiskSpaceSignal struct {
	path        string
	percentUsed int
}

func NewDiskSpaceSignal() *DiskSpaceSignal {
	return &DiskSpaceSignal{}
}

func (s *DiskSpaceSignal) Name() string {
	return "Full Tank"
}

func (s *DiskSpaceSignal) Emoji() string {
	return "💾"
}

func (s *DiskSpaceSignal) Diagnostic() string {
	return fmt.Sprintf("%s is %d%% full", s.path, s.percentUsed)
}

func (s *DiskSpaceSignal) Remediation() string {
	return "Free up disk space to prevent logging and audit trail failures"
}

func (s *DiskSpaceSignal) Check(ctx context.Context) bool {
	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}

	// Check root volume
	if s.checkPath("/") {
		return true
	}

	// Check home volume (might be different from root)
	// This is a simplified check - could be enhanced
	return false
}

func (s *DiskSpaceSignal) checkPath(path string) bool {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return false
	}

	// Calculate percentage used
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	if total == 0 {
		return false
	}

	percentUsed := int((used * 100) / total)

	if percentUsed > 90 {
		s.path = path
		s.percentUsed = percentUsed
		return true
	}

	return false
}
