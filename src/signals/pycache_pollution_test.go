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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create multiple __pycache__ directories
	// Note: with early exit optimization, we detect the first one and stop
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

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

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

// Tests for performance optimizations

func TestPyCachePollutionSignal_NonPythonProject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// No Python project markers - just some random directories
	os.MkdirAll("src/lib", 0755)
	os.WriteFile("src/main.go", []byte("package main"), 0644)

	// Create __pycache__ anyway (shouldn't be checked due to project gate)
	os.Mkdir("__pycache__", 0755)
	os.WriteFile("__pycache__/module.cpython-39.pyc", []byte("fake pyc"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when not in a Python project")
	}
}

func TestPyCachePollutionSignal_DepthLimit(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create a Python project marker
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a deeply nested __pycache__ (beyond maxPyCacheDepth=6)
	deepPath := "a/b/c/d/e/f/g/h/__pycache__"
	os.MkdirAll(deepPath, 0755)
	os.WriteFile(deepPath+"/module.cpython-39.pyc", []byte("fake"), 0644)

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	// The deeply nested __pycache__ should be skipped due to depth limit
	if result {
		t.Error("Expected false - should skip __pycache__ beyond depth limit")
	}
}

func TestPyCachePollutionSignal_EarlyExit(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a git repo
	exec.Command("git", "init").Run()

	// Create a Python project marker
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create many __pycache__ directories - with early exit, we should find the first and stop
	for i := 0; i < 10; i++ {
		dir := "pkg" + string(rune('a'+i)) + "/__pycache__"
		os.MkdirAll(dir, 0755)
		os.WriteFile(dir+"/module.cpython-39.pyc", []byte("fake"), 0644)
	}

	signal := NewPyCachePollutionSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when __pycache__ directories exist")
	}

	// With early exit, we should have found exactly one
	s := signal.(*PyCachePollutionSignal)
	if len(s.foundDirs) != 1 {
		t.Errorf("Expected exactly 1 found dir with early exit, got %d", len(s.foundDirs))
	}
}
