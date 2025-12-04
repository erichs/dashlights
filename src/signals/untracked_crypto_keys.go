package signals

import (
	"context"
	"os"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/gitutil"
)

// UntrackedCryptoKeysSignal checks for crypto keys not in .gitignore
type UntrackedCryptoKeysSignal struct {
	foundKeys []string
}

// NewUntrackedCryptoKeysSignal creates an UntrackedCryptoKeysSignal.
func NewUntrackedCryptoKeysSignal() *UntrackedCryptoKeysSignal {
	return &UntrackedCryptoKeysSignal{}
}

// Name returns the human-readable name of the signal.
func (s *UntrackedCryptoKeysSignal) Name() string {
	return "Dead Letter"
}

// Emoji returns the emoji associated with the signal.
func (s *UntrackedCryptoKeysSignal) Emoji() string {
	return "🗝️"
}

// Diagnostic returns a description of the detected untracked crypto keys.
func (s *UntrackedCryptoKeysSignal) Diagnostic() string {
	if len(s.foundKeys) == 0 {
		return "Cryptographic keys found not in .gitignore"
	}
	return "Unignored key: " + s.foundKeys[0]
}

// Remediation returns guidance on how to keep crypto keys out of source control.
func (s *UntrackedCryptoKeysSignal) Remediation() string {
	return "Add key files to .gitignore to prevent accidental commit"
}

// Check searches for key-like files in the current directory that are not ignored by git.
func (s *UntrackedCryptoKeysSignal) Check(ctx context.Context) bool {
	_ = ctx

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
				if !gitutil.IsIgnored(name) {
					s.foundKeys = append(s.foundKeys, name)
				}
				break
			}
		}
	}

	return len(s.foundKeys) > 0
}
