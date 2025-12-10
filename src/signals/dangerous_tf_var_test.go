package signals

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestDangerousTFVarSignal_NoTFVars(t *testing.T) {
	// Ensure no TF_VAR_ variables are set
	for _, env := range os.Environ() {
		if len(env) >= 7 && env[:7] == "TF_VAR_" {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) > 0 {
				os.Unsetenv(parts[0])
			}
		}
	}

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no TF_VAR_ variables are set")
	}
}

func TestDangerousTFVarSignal_SafeTFVar(t *testing.T) {
	// Set a safe TF_VAR variable
	os.Setenv("TF_VAR_region", "us-west-2")
	defer os.Unsetenv("TF_VAR_region")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false for safe TF_VAR_region")
	}
}

func TestDangerousTFVarSignal_AccessKey(t *testing.T) {
	// Set TF_VAR_access_key
	os.Setenv("TF_VAR_access_key", "AKIAIOSFODNN7EXAMPLE")
	defer os.Unsetenv("TF_VAR_access_key")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_access_key")
	}

	if len(signal.(*DangerousTFVarSignal).foundVars) == 0 {
		t.Error("Expected foundVars to contain the dangerous variable")
	}
}

func TestDangerousTFVarSignal_SecretKey(t *testing.T) {
	// Set TF_VAR_secret_key
	os.Setenv("TF_VAR_secret_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	defer os.Unsetenv("TF_VAR_secret_key")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_secret_key")
	}
}

func TestDangerousTFVarSignal_Password(t *testing.T) {
	// Set TF_VAR_db_password
	os.Setenv("TF_VAR_db_password", "supersecret123")
	defer os.Unsetenv("TF_VAR_db_password")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_db_password")
	}
}

func TestDangerousTFVarSignal_Token(t *testing.T) {
	// Set TF_VAR_api_token
	os.Setenv("TF_VAR_api_token", "ghp_1234567890abcdef")
	defer os.Unsetenv("TF_VAR_api_token")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_api_token")
	}
}

func TestDangerousTFVarSignal_APIKey(t *testing.T) {
	// Set TF_VAR_stripe_api_key
	os.Setenv("TF_VAR_stripe_api_key", "sk_test_1234567890")
	defer os.Unsetenv("TF_VAR_stripe_api_key")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_stripe_api_key")
	}
}

func TestDangerousTFVarSignal_PrivateKey(t *testing.T) {
	// Set TF_VAR_ssh_private_key
	os.Setenv("TF_VAR_ssh_private_key", "-----BEGIN RSA PRIVATE KEY-----")
	defer os.Unsetenv("TF_VAR_ssh_private_key")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_ssh_private_key")
	}
}

func TestDangerousTFVarSignal_CaseInsensitive(t *testing.T) {
	// Set TF_VAR_AWS_ACCESS_KEY (uppercase)
	os.Setenv("TF_VAR_AWS_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE")
	defer os.Unsetenv("TF_VAR_AWS_ACCESS_KEY")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for TF_VAR_AWS_ACCESS_KEY (case insensitive)")
	}
}

func TestDangerousTFVarSignal_Metadata(t *testing.T) {
	signal := NewDangerousTFVarSignal()

	if signal.Name() == "" {
		t.Error("Name should not be empty")
	}

	if signal.Emoji() == "" {
		t.Error("Emoji should not be empty")
	}

	if signal.Diagnostic() == "" {
		t.Error("Diagnostic should not be empty")
	}

	if signal.Remediation() == "" {
		t.Error("Remediation should not be empty")
	}
}

func TestDangerousTFVarSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_DANGEROUS_TF_VAR", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_DANGEROUS_TF_VAR")

	signal := NewDangerousTFVarSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
