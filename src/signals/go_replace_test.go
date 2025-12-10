package signals

import (
	"context"
	"os"
	"testing"
)

func TestGoReplaceSignal_NoGoMod(t *testing.T) {
	// Test in a directory without go.mod
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when go.mod doesn't exist")
	}
}

func TestGoReplaceSignal_NoReplace(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod without replace directives
	goMod := `module example.com/myproject

go 1.22

require (
	github.com/foo/bar v1.2.3
	github.com/baz/qux v2.0.0
)
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when go.mod has no replace directives")
	}
}

func TestGoReplaceSignal_WithReplace(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod with replace directive
	goMod := `module example.com/myproject

go 1.22

require (
	github.com/foo/bar v1.2.3
)

replace github.com/foo/bar => ../local/bar
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when go.mod has replace directive")
	}

	// Check diagnostic message
	diag := signal.Diagnostic()
	if !contains(diag, "github.com/foo/bar") {
		t.Errorf("Expected diagnostic to mention replaced module, got: %s", diag)
	}
}

func TestGoReplaceSignal_MultipleReplaces(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod with multiple replace directives
	goMod := `module example.com/myproject

go 1.22

replace (
	github.com/foo/bar => ../local/bar
	github.com/baz/qux => /home/user/projects/qux
)
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when go.mod has replace directives")
	}
}

func TestGoReplaceSignal_CommentedReplace(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod with commented replace directive
	goMod := `module example.com/myproject

go 1.22

// replace github.com/foo/bar => ../local/bar

require (
	github.com/foo/bar v1.2.3
)
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when replace directive is commented out")
	}
}

func TestGoReplaceSignal_ReplaceInline(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a go.mod with inline replace
	goMod := `module example.com/myproject

go 1.22

replace github.com/foo/bar => github.com/fork/bar v1.0.0
`
	os.WriteFile("go.mod", []byte(goMod), 0644)

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when go.mod has replace directive (even for remote fork)")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGoReplaceSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_GO_REPLACE", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_GO_REPLACE")

	signal := NewGoReplaceSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
