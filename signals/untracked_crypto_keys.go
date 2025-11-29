package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// UntrackedCryptoKeysSignal checks for crypto keys not in .gitignore
type UntrackedCryptoKeysSignal struct {
	foundKeys []string
}

func NewUntrackedCryptoKeysSignal() *UntrackedCryptoKeysSignal {
	return &UntrackedCryptoKeysSignal{}
}

func (s *UntrackedCryptoKeysSignal) Name() string {
	return "Dead Letter"
}

func (s *UntrackedCryptoKeysSignal) Emoji() string {
	return "🗝️"
}

func (s *UntrackedCryptoKeysSignal) Diagnostic() string {
	if len(s.foundKeys) == 0 {
		return "Cryptographic keys found not in .gitignore"
	}
	return "Unignored key: " + s.foundKeys[0]
}

func (s *UntrackedCryptoKeysSignal) Remediation() string {
	return "Add key files to .gitignore to prevent accidental commit"
}

func (s *UntrackedCryptoKeysSignal) Check(ctx context.Context) bool {
	s.foundKeys = []string{}

	// Key file extensions to look for
	keyExtensions := []string{".pem", ".key", ".p12", ".pfx", ".jks", ".keystore"}

	// Find key files in current directory
	entries, err := os.ReadDir(".")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range keyExtensions {
			if strings.HasSuffix(name, ext) {
				// Found a key file, check if it's in .gitignore
				if !isInGitignore(name) {
					s.foundKeys = append(s.foundKeys, name)
				}
				break
			}
		}
	}

	return len(s.foundKeys) > 0
}

func isInGitignore(filename string) bool {
	gitignoreFile, err := os.Open(".gitignore")
	if err != nil {
		return false // No .gitignore means not ignored
	}
	defer gitignoreFile.Close()

	scanner := bufio.NewScanner(gitignoreFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for exact match or pattern match
		if line == filename {
			return true
		}

		// Check for wildcard patterns
		if strings.Contains(line, "*") {
			matched, _ := filepath.Match(line, filename)
			if matched {
				return true
			}
		}
	}

	return false
}
