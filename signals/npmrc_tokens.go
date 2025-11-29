package signals

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// NpmrcTokensSignal checks for auth tokens in .npmrc in the project root
// These should be in the user's home directory, not committed to the repo
type NpmrcTokensSignal struct {
	foundToken string
}

func NewNpmrcTokensSignal() Signal {
	return &NpmrcTokensSignal{}
}

func (s *NpmrcTokensSignal) Name() string {
	return "NPM RC Tokens"
}

func (s *NpmrcTokensSignal) Emoji() string {
	return "📦" // Package (npm)
}

func (s *NpmrcTokensSignal) Diagnostic() string {
	return ".npmrc contains auth tokens (should be in ~/.npmrc, not project root)"
}

func (s *NpmrcTokensSignal) Remediation() string {
	return "Move .npmrc to ~/.npmrc and add .npmrc to .gitignore"
}

func (s *NpmrcTokensSignal) Check(ctx context.Context) bool {
	// Check if .npmrc exists in current directory
	file, err := os.Open(".npmrc")
	if err != nil {
		// No .npmrc file in project root - good
		return false
	}
	defer file.Close()

	// Scan the first few lines for auth tokens
	scanner := bufio.NewScanner(file)
	lineCount := 0
	maxLines := 100 // Only scan first 100 lines for performance

	for scanner.Scan() && lineCount < maxLines {
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
