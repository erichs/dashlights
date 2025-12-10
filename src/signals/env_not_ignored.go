package signals

import (
	"context"
	"os"

	"github.com/erichs/dashlights/src/signals/internal/gitutil"
)

// EnvNotIgnoredSignal checks if .env exists but isn't in .gitignore
type EnvNotIgnoredSignal struct{}

// NewEnvNotIgnoredSignal creates an EnvNotIgnoredSignal.
func NewEnvNotIgnoredSignal() *EnvNotIgnoredSignal {
	return &EnvNotIgnoredSignal{}
}

// Name returns the human-readable name of the signal.
func (s *EnvNotIgnoredSignal) Name() string {
	return "Unignored Secret"
}

// Emoji returns the emoji associated with the signal.
func (s *EnvNotIgnoredSignal) Emoji() string {
	return "📝"
}

// Diagnostic returns a description of the .env ignore misconfiguration.
func (s *EnvNotIgnoredSignal) Diagnostic() string {
	return ".env file exists but is not listed in .gitignore"
}

// Remediation returns guidance on how to ensure .env is ignored by git.
func (s *EnvNotIgnoredSignal) Remediation() string {
	return "Add '.env' to .gitignore to prevent accidental commit"
}

// Check verifies that an existing .env file is properly ignored by git.
func (s *EnvNotIgnoredSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_ENV_NOT_IGNORED") != "" {
		return false
	}

	// Check if .env exists in current directory
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return false
	}

	// Check if .gitignore exists and contains .env
	if gitutil.IsIgnored(".env") {
		return false // .env is ignored, no problem
	}

	// .env exists but is not in .gitignore
	return true
}
