package signals

import (
	"context"
	"os"
)

// RootLoginSignal checks if the user is logged in as root
type RootLoginSignal struct{}

func NewRootLoginSignal() *RootLoginSignal {
	return &RootLoginSignal{}
}

func (s *RootLoginSignal) Name() string {
	return "Danger Zone"
}

func (s *RootLoginSignal) Emoji() string {
	return "👑"
}

func (s *RootLoginSignal) Diagnostic() string {
	return "Running as root user (UID 0)"
}

func (s *RootLoginSignal) Remediation() string {
	return "Use a non-root account and run privileged commands with sudo instead"
}

func (s *RootLoginSignal) Check(ctx context.Context) bool {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// Check if effective user ID is 0 (root)
	return os.Geteuid() == 0
}
