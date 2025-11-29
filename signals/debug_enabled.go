package signals

import (
	"context"
	"os"
)

// DebugEnabledSignal detects debug/trace/verbose environment variables
// These can spam logs and leak sensitive data in production
type DebugEnabledSignal struct{}

func NewDebugEnabledSignal() Signal {
	return &DebugEnabledSignal{}
}

func (s *DebugEnabledSignal) Name() string {
	return "Debug Mode Enabled"
}

func (s *DebugEnabledSignal) Emoji() string {
	return "🐛" // Bug emoji
}

func (s *DebugEnabledSignal) Diagnostic() string {
	return "Debug/trace/verbose environment variables are set (can spam logs and leak data)"
}

func (s *DebugEnabledSignal) Remediation() string {
	return "Unset DEBUG, TRACE, and VERBOSE environment variables in production"
}

func (s *DebugEnabledSignal) Check(ctx context.Context) bool {
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
