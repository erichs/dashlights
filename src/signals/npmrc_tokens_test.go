package signals

import (
	"context"
	"os"
	"testing"
)

func TestNpmrcTokensSignal_Name(t *testing.T) {
	signal := NewNpmrcTokensSignal()
	if signal.Name() != "NPM RC Tokens" {
		t.Errorf("Expected 'NPM RC Tokens', got '%s'", signal.Name())
	}
}

func TestNpmrcTokensSignal_Emoji(t *testing.T) {
	signal := NewNpmrcTokensSignal()
	if signal.Emoji() != "📦" {
		t.Errorf("Expected '📦', got '%s'", signal.Emoji())
	}
}

func TestNpmrcTokensSignal_Diagnostic(t *testing.T) {
	signal := NewNpmrcTokensSignal()
	expected := ".npmrc contains auth tokens (should be in ~/.npmrc, not project root)"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestNpmrcTokensSignal_Remediation(t *testing.T) {
	signal := NewNpmrcTokensSignal()
	expected := "Move .npmrc to ~/.npmrc and add .npmrc to .gitignore"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestNpmrcTokensSignal_NoNpmrc(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when .npmrc doesn't exist")
	}
}

func TestNpmrcTokensSignal_NoTokens(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc without auth tokens
	npmrc := `registry=https://registry.npmjs.org/
save-exact=true
package-lock=false
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when .npmrc has no auth tokens")
	}
}

func TestNpmrcTokensSignal_WithAuthToken(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc with auth token
	npmrc := `registry=https://registry.npmjs.org/
//registry.npmjs.org/:_authToken=npm_1234567890abcdef
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .npmrc contains _authToken")
	}
}

func TestNpmrcTokensSignal_WithAuth(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc with _auth
	npmrc := `registry=https://registry.npmjs.org/
_auth=dXNlcm5hbWU6cGFzc3dvcmQ=
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .npmrc contains _auth")
	}
}

func TestNpmrcTokensSignal_ScopedRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc with scoped registry auth
	npmrc := `@myorg:registry=https://npm.pkg.github.com
//npm.pkg.github.com/:_authToken=ghp_1234567890
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .npmrc contains scoped registry auth token")
	}
}

func TestNpmrcTokensSignal_CommentedToken(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc with commented auth token
	npmrc := `registry=https://registry.npmjs.org/
# //registry.npmjs.org/:_authToken=npm_1234567890
save-exact=true
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when auth token is commented out")
	}
}

func TestNpmrcTokensSignal_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create empty .npmrc
	os.WriteFile(".npmrc", []byte(""), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when .npmrc is empty")
	}
}

func TestNpmrcTokensSignal_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .npmrc with only comments
	npmrc := `# NPM configuration
; This is a comment
# registry=https://registry.npmjs.org/
`
	os.WriteFile(".npmrc", []byte(npmrc), 0644)

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when .npmrc has only comments")
	}
}

func TestNpmrcTokensSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_NPMRC_TOKENS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_NPMRC_TOKENS")

	signal := NewNpmrcTokensSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
