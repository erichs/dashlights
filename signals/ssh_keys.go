package signals

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// SSHKeysSignal checks for SSH private keys with loose permissions
type SSHKeysSignal struct {
	foundKey    string
	foundPerms  os.FileMode
	diagnostic  string
	remediation string
}

func NewSSHKeysSignal() *SSHKeysSignal {
	return &SSHKeysSignal{}
}

func (s *SSHKeysSignal) Name() string {
	return "Open Door"
}

func (s *SSHKeysSignal) Emoji() string {
	return "🔑"
}

func (s *SSHKeysSignal) Diagnostic() string {
	return s.diagnostic
}

func (s *SSHKeysSignal) Remediation() string {
	return s.remediation
}

func (s *SSHKeysSignal) Check(ctx context.Context) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sshDir := filepath.Join(homeDir, ".ssh")

	// Check if .ssh directory exists
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return false
	}

	// Common private key filenames
	keyFiles := []string{
		"id_rsa",
		"id_dsa",
		"id_ecdsa",
		"id_ed25519",
		"id_ecdsa_sk",
		"id_ed25519_sk",
	}

	for _, keyFile := range keyFiles {
		keyPath := filepath.Join(sshDir, keyFile)
		info, err := os.Stat(keyPath)
		if err != nil {
			continue // File doesn't exist, skip
		}

		// Check if it's a regular file
		if !info.Mode().IsRegular() {
			continue
		}

		perms := info.Mode().Perm()

		// Private keys should be 0600 (owner read/write only)
		// Flag if permissions are more permissive
		if perms != 0600 {
			s.foundKey = keyPath
			s.foundPerms = perms
			s.diagnostic = filepath.Base(keyPath) + " has permissions " + formatPerms(perms) + " (should be 0600)"
			s.remediation = "chmod 600 " + keyPath
			return true
		}
	}

	return false
}

func formatPerms(mode os.FileMode) string {
	return fmt.Sprintf("0%o", mode&0777)
}
