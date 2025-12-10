package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSSHKeysSignal_Name(t *testing.T) {
	signal := NewSSHKeysSignal()
	if signal.Name() != "Open Door" {
		t.Errorf("Expected 'Open Door', got '%s'", signal.Name())
	}
}

func TestSSHKeysSignal_Emoji(t *testing.T) {
	signal := NewSSHKeysSignal()
	if signal.Emoji() != "🔑" {
		t.Errorf("Expected '🔑', got '%s'", signal.Emoji())
	}
}

func TestSSHKeysSignal_Check_NoSSHDir(t *testing.T) {
	// Create a temp directory without .ssh
	tmpDir := t.TempDir()

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewSSHKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when .ssh directory doesn't exist")
	}
}

func TestSSHKeysSignal_Check_SecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory with .ssh and secure key
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	keyFile := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(keyFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewSSHKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when SSH key has secure permissions (0600)")
	}
}

func TestSSHKeysSignal_Check_InsecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory with .ssh and insecure key
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	keyFile := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(keyFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewSSHKeysSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when SSH key has insecure permissions (0644)")
	}

	// Verify diagnostic and remediation are populated
	if signal.Diagnostic() == "" {
		t.Error("Expected Diagnostic() to be populated after Check()")
	}

	if signal.Remediation() == "" {
		t.Error("Expected Remediation() to be populated after Check()")
	}
}

func TestSSHKeysSignal_Check_MultipleKeys(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	// Create a temp directory with .ssh and multiple keys
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	// One secure, one insecure
	secureKey := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(secureKey, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	insecureKey := filepath.Join(sshDir, "id_ed25519")
	if err := os.WriteFile(insecureKey, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewSSHKeysSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when at least one SSH key has insecure permissions")
	}
}

func TestSSHKeysSignal_formatPerms(t *testing.T) {
	tests := []struct {
		mode     os.FileMode
		expected string
	}{
		{0600, "0600"},
		{0644, "0644"},
		{0666, "0666"},
		{0700, "0700"},
		{0755, "0755"},
		{0777, "0777"},
	}

	for _, tt := range tests {
		result := formatPerms(tt.mode)
		if result != tt.expected {
			t.Errorf("formatPerms(%o) = %s, expected %s", tt.mode, result, tt.expected)
		}
	}
}

func TestSSHKeysSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_SSH_KEYS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_SSH_KEYS")

	signal := NewSSHKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
