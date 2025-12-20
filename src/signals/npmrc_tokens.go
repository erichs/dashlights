package signals

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/fileutil"
)

// NpmrcTokensSignal checks for auth tokens in .npmrc in the project root
// These should be in the user's home directory, not committed to the repo
type NpmrcTokensSignal struct {
	foundToken string
}

// NewNpmrcTokensSignal creates an NpmrcTokensSignal.
func NewNpmrcTokensSignal() Signal {
	return &NpmrcTokensSignal{}
}

// Name returns the human-readable name of the signal.
func (s *NpmrcTokensSignal) Name() string {
	return "NPM RC Tokens"
}

// Emoji returns the emoji associated with the signal.
func (s *NpmrcTokensSignal) Emoji() string {
	return "📦" // Package (npm)
}

// Diagnostic returns a description of the detected npm token configuration.
func (s *NpmrcTokensSignal) Diagnostic() string {
	return ".npmrc contains auth tokens (should be in ~/.npmrc, not project root)"
}

// Remediation returns guidance on how to handle npm auth tokens safely.
func (s *NpmrcTokensSignal) Remediation() string {
	return "Move .npmrc to ~/.npmrc and add .npmrc to .gitignore"
}

// Check inspects the project .npmrc for embedded auth tokens.
func (s *NpmrcTokensSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_NPMRC_TOKENS") != "" {
		return false
	}

	// Skip if we're in the home directory - .npmrc with tokens is expected there
	homeDir, err := os.UserHomeDir()
	if err == nil {
		cwd, cwdErr := os.Getwd()
		if cwdErr == nil && cwd == homeDir {
			return false
		}
	}

	// Check if .npmrc exists in current directory
	const maxNpmrcBytes = 128 * 1024

	data, err := fileutil.ReadFileLimitedString(".npmrc", maxNpmrcBytes)
	if err != nil {
		// No .npmrc file in project root - good
		return false
	}

	// Scan the first few lines for auth tokens
	scanner := bufio.NewScanner(strings.NewReader(data))
	lineCount := 0
	maxLines := 100 // Only scan first 100 lines for performance

	for scanner.Scan() && lineCount < maxLines {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for auth tokens
		if strings.Contains(line, "_auth=") ||
			strings.Contains(line, "_authToken=") ||
			strings.Contains(line, "//registry.npmjs.org/:_authToken") {
			s.foundToken = line
			return true
		}
	}

	return false
}
