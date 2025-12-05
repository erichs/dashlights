// Package gitutil provides shared utilities for git-related operations.
package gitutil

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IsIgnored checks if a filename is covered by a pattern in .gitignore.
// It reads the .gitignore file from the current working directory and checks
// if the given filename matches any non-comment, non-empty line.
//
// Pattern matching rules:
//   - Exact match: "filename" matches filename
//   - Wildcard patterns using filepath.Match (e.g., "*.pem" matches "cert.pem")
//   - Substring match: patterns containing the filename (e.g., "**/.env*" matches ".env")
//
// Returns true if the file is ignored, false otherwise.
// If .gitignore doesn't exist or cannot be read, returns false.
func IsIgnored(filename string) bool {
	return IsIgnoredIn(".gitignore", filename)
}

// IsIgnoredIn checks if a filename is covered by a pattern in the specified gitignore file.
// This is useful for testing or when checking against a specific gitignore file path.
func IsIgnoredIn(gitignorePath, filename string) bool {
	// filepath.Clean for gosec G304
	file, err := os.Open(filepath.Clean(gitignorePath))
	if err != nil {
		return false // No .gitignore means not ignored
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for exact match
		if line == filename {
			return true
		}

		// Check for wildcard patterns (only if pattern contains *)
		if strings.Contains(line, "*") {
			matched, err := filepath.Match(line, filename)
			if err != nil {
				// Invalid pattern, skip it
				continue
			}
			if matched {
				return true
			}
		}

		// Check for substring match (for patterns like **/.env* or .env)
		// This handles cases where the pattern contains the filename
		if strings.Contains(line, filename) {
			return true
		}
	}

	return false
}
