package signals

import (
	"context"
	"os"
	"testing"
)

func TestEnvNotIgnoredSignal_Name(t *testing.T) {
	signal := NewEnvNotIgnoredSignal()
	if signal.Name() != "Unignored Secret" {
		t.Errorf("Expected 'Unignored Secret', got '%s'", signal.Name())
	}
}

func TestEnvNotIgnoredSignal_Emoji(t *testing.T) {
	signal := NewEnvNotIgnoredSignal()
	if signal.Emoji() != "📝" {
		t.Errorf("Expected '📝', got '%s'", signal.Emoji())
	}
}

func TestEnvNotIgnoredSignal_Diagnostic(t *testing.T) {
	signal := NewEnvNotIgnoredSignal()
	expected := ".env file exists but is not listed in .gitignore"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestEnvNotIgnoredSignal_Remediation(t *testing.T) {
	signal := NewEnvNotIgnoredSignal()
	expected := "Add '.env' to .gitignore to prevent accidental commit"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestEnvNotIgnoredSignal_Check_NoEnvFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// No .env file exists
	if signal.Check(ctx) {
		t.Error("Expected false when .env doesn't exist")
	}
}

func TestEnvNotIgnoredSignal_Check_EnvExistsNoGitignore(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env exists but no .gitignore
	if !signal.Check(ctx) {
		t.Error("Expected true when .env exists but .gitignore doesn't")
	}
}

func TestEnvNotIgnoredSignal_Check_EnvInGitignore(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	// Create .gitignore with .env
	os.WriteFile(".gitignore", []byte(".env\n"), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env is in .gitignore
	if signal.Check(ctx) {
		t.Error("Expected false when .env is in .gitignore")
	}
}

func TestEnvNotIgnoredSignal_Check_EnvNotInGitignore(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	// Create .gitignore without .env
	os.WriteFile(".gitignore", []byte("node_modules/\n*.log\n"), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env exists but not in .gitignore
	if !signal.Check(ctx) {
		t.Error("Expected true when .env exists but not in .gitignore")
	}
}

func TestEnvNotIgnoredSignal_Check_WildcardPattern(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	// Create .gitignore with wildcard pattern
	os.WriteFile(".gitignore", []byte("*.env\n"), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env is covered by *.env pattern
	if signal.Check(ctx) {
		t.Error("Expected false when .env is covered by *.env pattern")
	}
}

func TestEnvNotIgnoredSignal_Check_CommentsAndEmptyLines(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	// Create .gitignore with comments and empty lines
	gitignoreContent := `# This is a comment

node_modules/
# Another comment
.env
*.log
`
	os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env is in .gitignore (after comments)
	if signal.Check(ctx) {
		t.Error("Expected false when .env is in .gitignore with comments")
	}
}

func TestEnvNotIgnoredSignal_Check_SubstringMatch(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Create .env file
	os.WriteFile(".env", []byte("SECRET=value"), 0644)

	// Create .gitignore with pattern containing .env
	os.WriteFile(".gitignore", []byte("**/.env*\n"), 0644)

	signal := NewEnvNotIgnoredSignal()
	ctx := context.Background()

	// .env is covered by pattern containing .env
	if signal.Check(ctx) {
		t.Error("Expected false when .env is covered by pattern containing .env")
	}
}
