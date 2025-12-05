package signals

import (
	"context"
	"os"
)

// DebugEnabledSignal detects debug/trace/verbose environment variables
// These can spam logs and leak sensitive data in production
type DebugEnabledSignal struct{}

// NewDebugEnabledSignal creates a DebugEnabledSignal.
func NewDebugEnabledSignal() Signal {
	return &DebugEnabledSignal{}
}

// Name returns the human-readable name of the signal.
func (s *DebugEnabledSignal) Name() string {
	return "Debug Mode Enabled"
}

// Emoji returns the emoji associated with the signal.
func (s *DebugEnabledSignal) Emoji() string {
	return "🐛" // Bug emoji
}

// Diagnostic returns a description of detected debug environment variables.
func (s *DebugEnabledSignal) Diagnostic() string {
	return "Debug/trace/verbose environment variables are set (can spam logs and leak data)"
}

// Remediation returns guidance on disabling debug environment variables.
func (s *DebugEnabledSignal) Remediation() string {
	return "Unset DEBUG, TRACE, and VERBOSE environment variables in production"
}

// Check reports whether common debug environment variables are set.
func (s *DebugEnabledSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check for common debug environment variables
	// The actual value doesn't matter - just that they are set
	debugVars := []string{
		"DEBUG",
		"TRACE",
		"VERBOSE",
	}

	for _, varName := range debugVars {
		if _, exists := os.LookupEnv(varName); exists {
			return true
		}
	}

	return false
}
