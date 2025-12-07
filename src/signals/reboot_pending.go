package signals

import (
	"context"
	"os"
	"runtime"
)

// RebootPendingSignal checks for pending reboot (unactivated patches)
type RebootPendingSignal struct{}

// NewRebootPendingSignal creates a RebootPendingSignal.
func NewRebootPendingSignal() *RebootPendingSignal {
	return &RebootPendingSignal{}
}

// Name returns the human-readable name of the signal.
func (s *RebootPendingSignal) Name() string {
	return "Reboot Pending"
}

// Emoji returns the emoji associated with the signal.
func (s *RebootPendingSignal) Emoji() string {
	return "♻️"
}

// Diagnostic returns a description of the pending reboot condition.
func (s *RebootPendingSignal) Diagnostic() string {
	return "System reboot required to activate security patches"
}

// Remediation returns guidance on how to resolve the pending reboot.
func (s *RebootPendingSignal) Remediation() string {
	return "Reboot system to activate installed kernel patches"
}

// Check determines whether a system reboot is pending.
func (s *RebootPendingSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_REBOOT_PENDING") != "" {
		return false
	}

	// Check for Debian/Ubuntu reboot-required flag
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/var/run/reboot-required"); err == nil {
			return true
		}
	}

	// TODO (macOS): Check for pending macOS updates that require reboot

	return false
}
