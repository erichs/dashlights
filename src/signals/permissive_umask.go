package signals

import (
	"context"
	"fmt"
	"sync"
	"syscall"
)

var (
	// umaskMutex serializes umask checks to prevent race conditions.
	// syscall.Umask() is a process-wide operation that temporarily modifies
	// the umask, so concurrent calls would interfere with each other.
	umaskMutex sync.Mutex
)

// PermissiveUmaskSignal checks for overly permissive umask
type PermissiveUmaskSignal struct {
	currentUmask int
}

func NewPermissiveUmaskSignal() *PermissiveUmaskSignal {
	return &PermissiveUmaskSignal{}
}

func (s *PermissiveUmaskSignal) Name() string {
	return "Loose Cannon"
}

func (s *PermissiveUmaskSignal) Emoji() string {
	return "😷"
}

func (s *PermissiveUmaskSignal) Diagnostic() string {
	return "Permissive umask detected: " + formatUmask(s.currentUmask)
}

func (s *PermissiveUmaskSignal) Remediation() string {
	return "Set umask to 0022 or 0027 for better security"
}

func (s *PermissiveUmaskSignal) Check(ctx context.Context) bool {
	// Serialize umask checks to prevent race conditions across concurrent goroutines.
	// syscall.Umask() modifies process-wide state, so we must ensure only one
	// goroutine accesses it at a time.
	umaskMutex.Lock()
	defer umaskMutex.Unlock()

	// Get current umask (note: this temporarily changes it)
	oldUmask := syscall.Umask(0)
	syscall.Umask(oldUmask) // Restore immediately

	s.currentUmask = oldUmask

	// Flag if umask is 0000 (world-writable) or 0002 (group-writable for new files)
	// Most secure would be 0077, but 0022 is common and acceptable
	if oldUmask == 0000 || oldUmask == 0002 {
		return true
	}

	return false
}

func formatUmask(umask int) string {
	return fmt.Sprintf("0%03o", umask)
}
