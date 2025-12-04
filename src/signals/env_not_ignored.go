package signals

import (
	"bufio"
	"context"
	"os"
	"strings"
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

	// Check if .env exists in current directory
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return false
	}

	// Check if .gitignore exists
	gitignoreFile, err := os.Open(".gitignore")
	if err != nil {
		// .gitignore doesn't exist, so .env is not ignored
		return true
	}
	defer gitignoreFile.Close()

	// Check if .env is in .gitignore
	scanner := bufio.NewScanner(gitignoreFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check for exact match or pattern match
		if line == ".env" || line == "*.env" || strings.Contains(line, ".env") {
			return false // .env is ignored, no problem
		}
	}

	// .env exists but is not in .gitignore
	return true
}
