package signals

import (
	"context"
	"os"
)

// RootLoginSignal checks if the user is logged in as root
type RootLoginSignal struct{}

// NewRootLoginSignal creates a RootLoginSignal.
func NewRootLoginSignal() *RootLoginSignal {
	return &RootLoginSignal{}
}

// Name returns the human-readable name of the signal.
func (s *RootLoginSignal) Name() string {
	return "Danger Zone"
}

// Emoji returns the emoji associated with the signal.
func (s *RootLoginSignal) Emoji() string {
	return "👑"
}

// Diagnostic returns a description of running as root.
func (s *RootLoginSignal) Diagnostic() string {
	return "Running as root user (UID 0)"
}

// Remediation returns guidance on avoiding running as root.
func (s *RootLoginSignal) Remediation() string {
	return "Use a non-root account and run privileged commands with sudo instead"
}

// Check reports whether the effective user ID is root.
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
