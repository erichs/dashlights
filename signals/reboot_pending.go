package signals

import (
	"context"
	"os"
	"runtime"
)

// RebootPendingSignal checks for pending reboot (unactivated patches)
type RebootPendingSignal struct{}

func NewRebootPendingSignal() *RebootPendingSignal {
	return &RebootPendingSignal{}
}

func (s *RebootPendingSignal) Name() string {
	return "Reboot Pending"
}

func (s *RebootPendingSignal) Emoji() string {
	return "♻️"
}

func (s *RebootPendingSignal) Diagnostic() string {
	return "System reboot required to activate security patches"
}

func (s *RebootPendingSignal) Remediation() string {
	return "Reboot system to activate installed kernel patches"
}

func (s *RebootPendingSignal) Check(ctx context.Context) bool {
	// Check for Debian/Ubuntu reboot-required flag
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/var/run/reboot-required"); err == nil {
			return true
		}
	}

	// TODO (macOS): Check for pending macOS updates that require reboot

	return false
}
