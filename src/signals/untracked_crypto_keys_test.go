package signals

import (
	"context"
	"os"
	"testing"
)

func TestUntrackedCryptoKeysSignal_Name(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	if signal.Name() != "Dead Letter" {
		t.Errorf("Expected 'Dead Letter', got '%s'", signal.Name())
	}
}

func TestUntrackedCryptoKeysSignal_Emoji(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	if signal.Emoji() != "🗝️" {
		t.Errorf("Expected '🗝️', got '%s'", signal.Emoji())
	}
}

func TestUntrackedCryptoKeysSignal_Diagnostic_NoKeys(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	signal.foundKeys = []string{}
	expected := "Cryptographic keys found not in .gitignore"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUntrackedCryptoKeysSignal_Diagnostic_WithKeys(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	signal.foundKeys = []string{"private.key", "cert.pem"}
	expected := "Unignored key: private.key"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUntrackedCryptoKeysSignal_Remediation(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	expected := "Add key files to .gitignore to prevent accidental commit"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestUntrackedCryptoKeysSignal_Check_NoKeys(t *testing.T) {
	// Create a temp directory without key files
	tmpDir := t.TempDir()

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when no key files exist")
	}
}

func TestUntrackedCryptoKeysSignal_Check_KeyInGitignore(t *testing.T) {
	// Create a temp directory with key file and .gitignore
	tmpDir := t.TempDir()

	// Create key file
	if err := os.WriteFile(tmpDir+"/private.key", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .gitignore with key file
	if err := os.WriteFile(tmpDir+"/.gitignore", []byte("private.key\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when key file is in .gitignore")
	}
}

func TestUntrackedCryptoKeysSignal_Check_KeyNotInGitignore(t *testing.T) {
	// Create a temp directory with key file but no .gitignore
	tmpDir := t.TempDir()

	// Create key file
	if err := os.WriteFile(tmpDir+"/private.key", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when key file is not in .gitignore")
	}

	if len(signal.foundKeys) == 0 {
		t.Error("Expected foundKeys to contain private.key")
	}
}

func TestUntrackedCryptoKeysSignal_Check_MultipleKeyExtensions(t *testing.T) {
	// Create a temp directory with multiple key files
	tmpDir := t.TempDir()

	// Create various key files
	keyFiles := []string{"cert.pem", "private.key", "keystore.jks", "cert.p12"}
	for _, f := range keyFiles {
		if err := os.WriteFile(tmpDir+"/"+f, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when multiple key files are not in .gitignore")
	}

	if len(signal.foundKeys) == 0 {
		t.Error("Expected foundKeys to contain key files")
	}
}

func TestUntrackedCryptoKeysSignal_Check_InvalidGitignorePattern(t *testing.T) {
	// Create a temp directory with key file and .gitignore with invalid pattern
	tmpDir := t.TempDir()

	// Create key file
	if err := os.WriteFile(tmpDir+"/private.key", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .gitignore with invalid pattern (malformed bracket expression)
	// This should be skipped and the key should be detected as unignored
	if err := os.WriteFile(tmpDir+"/.gitignore", []byte("[invalid\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	// Should detect the key since the invalid pattern is skipped
	if !signal.Check(ctx) {
		t.Error("Expected true when key file is not matched by invalid pattern")
	}

	if len(signal.foundKeys) == 0 {
		t.Error("Expected foundKeys to contain private.key")
	}
}

func TestUntrackedCryptoKeysSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_UNTRACKED_CRYPTO_KEYS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_UNTRACKED_CRYPTO_KEYS")

	signal := NewUntrackedCryptoKeysSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
