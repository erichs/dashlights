package signals

import (
	"context"
	"os"
	"testing"
)

func TestNakedCredentialsSignal_Name(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	if signal.Name() != "Naked Credential" {
		t.Errorf("Expected 'Naked Credential', got '%s'", signal.Name())
	}
}

func TestNakedCredentialsSignal_Emoji(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	if signal.Emoji() != "🩲" {
		t.Errorf("Expected '🩲', got '%s'", signal.Emoji())
	}
}

func TestNakedCredentialsSignal_Diagnostic_NoVars(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	signal.foundVars = []string{}
	expected := "Raw secrets detected in environment variables"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestNakedCredentialsSignal_Diagnostic_WithVars(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	signal.foundVars = []string{"AWS_SECRET_ACCESS_KEY", "GITHUB_TOKEN"}
	expected := "Raw secrets in environment: AWS_SECRET_ACCESS_KEY, GITHUB_TOKEN"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestNakedCredentialsSignal_Remediation(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	expected := "Use 1Password (op://), dotenvx (encrypted:), or other secret management tools"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestNakedCredentialsSignal_Check_NoSecrets(t *testing.T) {
	t.Helper()
	// Save and restore env vars
	oldAWS := os.Getenv("AWS_SECRET_ACCESS_KEY")
	oldGithub := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldAWS != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", oldAWS)
		} else {
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}
		if oldGithub != "" {
			os.Setenv("GITHUB_TOKEN", oldGithub)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Clear secret env vars
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("GITHUB_TOKEN")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	// May return true if other secrets exist in environment
	// Just verify it doesn't panic
	signal.Check(ctx)
}

func TestNakedCredentialsSignal_Check_AWSSecret(t *testing.T) {
	// Save and restore env var
	oldAWS := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer func() {
		if oldAWS != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", oldAWS)
		} else {
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}
	}()

	// Set AWS secret
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when AWS_SECRET_ACCESS_KEY is set")
	}

	found := false
	for _, v := range signal.foundVars {
		if v == "AWS_SECRET_ACCESS_KEY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected AWS_SECRET_ACCESS_KEY in foundVars")
	}
}

func TestNakedCredentialsSignal_Check_EmptyValue(t *testing.T) {
	// Save and restore env var
	oldGithub := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldGithub != "" {
			os.Setenv("GITHUB_TOKEN", oldGithub)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Set empty value
	os.Setenv("GITHUB_TOKEN", "")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	signal.Check(ctx)

	// Empty values should be skipped
	for _, v := range signal.foundVars {
		if v == "GITHUB_TOKEN" {
			t.Error("Expected GITHUB_TOKEN to be skipped when empty")
		}
	}
}

func TestNakedCredentialsSignal_Check_DashlightPrefix(t *testing.T) {
	// Save and restore env var
	oldDashlight := os.Getenv("DASHLIGHT_SECRET")
	defer func() {
		if oldDashlight != "" {
			os.Setenv("DASHLIGHT_SECRET", oldDashlight)
		} else {
			os.Unsetenv("DASHLIGHT_SECRET")
		}
	}()

	// Set DASHLIGHT_ prefixed var
	os.Setenv("DASHLIGHT_SECRET", "test-value")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	signal.Check(ctx)

	// DASHLIGHT_ variables should be skipped
	for _, v := range signal.foundVars {
		if v == "DASHLIGHT_SECRET" {
			t.Error("Expected DASHLIGHT_SECRET to be skipped")
		}
	}
}

func TestNakedCredentialsSignal_Check_FalsePositives(t *testing.T) {
	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	signal.Check(ctx)

	// Common false positives should be filtered
	falsePositives := []string{"PATH", "HOME", "SHELL", "TERM"}
	for _, fp := range falsePositives {
		for _, v := range signal.foundVars {
			if v == fp {
				t.Errorf("Expected %s to be filtered as false positive", fp)
			}
		}
	}
}

func TestNakedCredentialsSignal_Check_OnePasswordReference(t *testing.T) {
	// Save and restore env var
	oldAWS := os.Getenv("AWS_SECRET_ACCESS_KEY")
	defer func() {
		if oldAWS != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", oldAWS)
		} else {
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}
	}()

	// Set AWS secret with 1Password reference
	os.Setenv("AWS_SECRET_ACCESS_KEY", "op://vault/item/field")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	signal.Check(ctx)

	// 1Password references should be skipped
	for _, v := range signal.foundVars {
		if v == "AWS_SECRET_ACCESS_KEY" {
			t.Error("Expected AWS_SECRET_ACCESS_KEY with op:// prefix to be skipped")
		}
	}
}

func TestNakedCredentialsSignal_Check_DotenvxEncrypted(t *testing.T) {
	// Save and restore env var
	oldGithub := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldGithub != "" {
			os.Setenv("GITHUB_TOKEN", oldGithub)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Set GitHub token with dotenvx encryption
	os.Setenv("GITHUB_TOKEN", "encrypted:BEE4F...")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	signal.Check(ctx)

	// dotenvx encrypted values should be skipped
	for _, v := range signal.foundVars {
		if v == "GITHUB_TOKEN" {
			t.Error("Expected GITHUB_TOKEN with encrypted: prefix to be skipped")
		}
	}
}

func TestNakedCredentialsSignal_Check_RawSecretStillDetected(t *testing.T) {
	// Save and restore env var
	oldStripe := os.Getenv("STRIPE_SECRET_KEY")
	defer func() {
		if oldStripe != "" {
			os.Setenv("STRIPE_SECRET_KEY", oldStripe)
		} else {
			os.Unsetenv("STRIPE_SECRET_KEY")
		}
	}()

	// Set raw secret (not encrypted or referenced)
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_1234567890")

	signal := NewNakedCredentialsSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when raw secret is set")
	}

	found := false
	for _, v := range signal.foundVars {
		if v == "STRIPE_SECRET_KEY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected STRIPE_SECRET_KEY in foundVars when it's a raw secret")
	}
}
