package gitutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsIgnored_NoGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// No .gitignore file exists
	if IsIgnored(".env") {
		t.Error("Expected false when .gitignore doesn't exist")
	}
}

func TestIsIgnored_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	os.WriteFile(".gitignore", []byte(".env\n"), 0644)

	if !IsIgnored(".env") {
		t.Error("Expected true for exact match")
	}
	if IsIgnored(".env.local") {
		t.Error("Expected false for non-matching filename")
	}
}

func TestIsIgnored_WildcardPattern(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	os.WriteFile(".gitignore", []byte("*.pem\n*.key\n"), 0644)

	tests := []struct {
		filename string
		expected bool
	}{
		{"cert.pem", true},
		{"private.key", true},
		{"cert.crt", false},
		{"file.txt", false},
	}

	for _, tt := range tests {
		if got := IsIgnored(tt.filename); got != tt.expected {
			t.Errorf("IsIgnored(%q) = %v, want %v", tt.filename, got, tt.expected)
		}
	}
}

func TestIsIgnored_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	gitignoreContent := `# This is a comment

node_modules/
# Another comment
.env
*.log
`
	os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)

	if !IsIgnored(".env") {
		t.Error("Expected true when .env is in .gitignore with comments")
	}
}

func TestIsIgnored_SubstringMatch(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Patterns like **/.env* should match .env via substring
	os.WriteFile(".gitignore", []byte("**/.env*\n"), 0644)

	if !IsIgnored(".env") {
		t.Error("Expected true when pattern contains .env")
	}
}

func TestIsIgnored_InvalidPattern(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Invalid pattern (malformed bracket expression) should be skipped
	os.WriteFile(".gitignore", []byte("[invalid\n"), 0644)

	// Should return false since the invalid pattern is skipped
	if IsIgnored("private.key") {
		t.Error("Expected false when pattern is invalid")
	}
}

func TestIsIgnoredIn_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom.ignore")
	os.WriteFile(customPath, []byte("*.pem\n"), 0644)

	if !IsIgnoredIn(customPath, "cert.pem") {
		t.Error("Expected true for custom gitignore path")
	}
	if IsIgnoredIn(customPath, "cert.key") {
		t.Error("Expected false for non-matching pattern in custom path")
	}
}

func TestIsIgnored_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	gitignoreContent := `# Key files
*.pem
*.key
*.p12
*.pfx
*.jks
*.keystore
`
	os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)

	keyFiles := []string{"cert.pem", "private.key", "keystore.jks", "cert.p12"}
	for _, f := range keyFiles {
		if !IsIgnored(f) {
			t.Errorf("Expected %s to be ignored", f)
		}
	}
}
