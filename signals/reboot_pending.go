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

	if runtime.GOOS == "darwin" {
		// TODO: Check for macOS updates
		// check output of defaults read /Library/Preferences/com.apple.SoftwareUpdate.plist RecommendedUpdates
		// for lines containing 'DisplayName' and 'Security'
		if _, err := os.Stat("/Library/Preferences/com.apple.SoftwareUpdate.plist"); err == nil {

			return true
		}

	}

	// macOS doesn't have a simple file-based indicator
	// Could check for pending updates via softwareupdate, but that's slow

	return false
}
