package signals

import (
	"context"
	"os/exec"
	"runtime"
)

// SudoCachedSignal checks for active sudo timestamp
type SudoCachedSignal struct{}

func NewSudoCachedSignal() *SudoCachedSignal {
	return &SudoCachedSignal{}
}

func (s *SudoCachedSignal) Name() string {
	return "Hot Potato"
}

func (s *SudoCachedSignal) Emoji() string {
	return "⚡"
}

func (s *SudoCachedSignal) Diagnostic() string {
	return "Active sudo session detected (timestamp not expired)"
}

func (s *SudoCachedSignal) Remediation() string {
	return "Run 'sudo -k' before leaving terminal unattended"
}

func (s *SudoCachedSignal) Check(ctx context.Context) bool {
	// Only applicable on Unix-like systems
	if runtime.GOOS == "windows" {
		return false
	}
	
	// Try to run a command with sudo -n (non-interactive)
	// If it succeeds, we have a cached sudo session
	cmd := exec.CommandContext(ctx, "sudo", "-n", "true")
	err := cmd.Run()
	
	// If err is nil, sudo succeeded without password (cached)
	return err == nil
}

