package signals

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// EnvNotIgnoredSignal checks if .env exists but isn't in .gitignore
type EnvNotIgnoredSignal struct{}

func NewEnvNotIgnoredSignal() *EnvNotIgnoredSignal {
	return &EnvNotIgnoredSignal{}
}

func (s *EnvNotIgnoredSignal) Name() string {
	return "Unignored Secret"
}

func (s *EnvNotIgnoredSignal) Emoji() string {
	return "📝"
}

func (s *EnvNotIgnoredSignal) Diagnostic() string {
	return ".env file exists but is not listed in .gitignore"
}

func (s *EnvNotIgnoredSignal) Remediation() string {
	return "Add '.env' to .gitignore to prevent accidental commit"
}

func (s *EnvNotIgnoredSignal) Check(ctx context.Context) bool {
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
