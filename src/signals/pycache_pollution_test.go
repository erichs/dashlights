package signals

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPyCachePollutionSignal_NoGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create __pycache__ but no .git
	os.Mkdir("__pycache__", 0755)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when not in a git repository")
	}
}

func TestPyCachePollutionSignal_NoPyCache(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo without __pycache__
	exec.Command("git", "init").Run()

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no __pycache__ directories exist")
	}
}

func TestPyCachePollutionSignal_WithPyCache(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create __pycache__ with .pyc files
	os.Mkdir("__pycache__", 0755)
	os.WriteFile("__pycache__/module.cpython-39.pyc", []byte("fake pyc"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when __pycache__ with .pyc files exists")
	}

	// Verify diagnostic is populated
	diagnostic := signal.Diagnostic()
	if !strings.Contains(diagnostic, "__pycache__") {
		t.Errorf("Expected diagnostic to contain '__pycache__', got '%s'", diagnostic)
	}
}

func TestPyCachePollutionSignal_NestedPyCache(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create nested __pycache__
	os.MkdirAll("src/mypackage/__pycache__", 0755)
	os.WriteFile("src/mypackage/__pycache__/module.cpython-39.pyc", []byte("fake pyc"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when nested __pycache__ with .pyc files exists")
	}
}

func TestPyCachePollutionSignal_EmptyPyCache(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create empty __pycache__ (no .pyc files)
	os.Mkdir("__pycache__", 0755)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when __pycache__ is empty (no .pyc files)")
	}
}

func TestPyCachePollutionSignal_MultiplePyCacheDirs(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create multiple __pycache__ directories
	os.MkdirAll("package1/__pycache__", 0755)
	os.WriteFile("package1/__pycache__/mod1.cpython-39.pyc", []byte("fake"), 0644)

	os.MkdirAll("package2/__pycache__", 0755)
	os.WriteFile("package2/__pycache__/mod2.cpython-39.pyc", []byte("fake"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when multiple __pycache__ directories exist")
	}
}

func TestPyCachePollutionSignal_SkipsGitDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create __pycache__ inside .git (should be ignored)
	os.MkdirAll(".git/__pycache__", 0755)
	os.WriteFile(".git/__pycache__/test.pyc", []byte("fake"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false - should skip .git directory")
	}
}

func TestPyCachePollutionSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_PYCACHE_POLLUTION", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_PYCACHE_POLLUTION")

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
