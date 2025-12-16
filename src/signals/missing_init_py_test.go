package signals

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestMissingInitPySignal_NoPythonFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a directory with no Python files
	os.Mkdir("mydir", 0755)
	os.WriteFile("mydir/readme.txt", []byte("hello"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no Python files exist")
	}
}

func TestMissingInitPySignal_WithInitPy(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a proper Python package
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/__init__.py", []byte(""), 0644)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when __init__.py exists")
	}
}

func TestMissingInitPySignal_Diagnostic_NoFoundDirs(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	// Check with no Python files - should not trigger
	signal.Check(ctx)

	// Diagnostic should still work even when not triggered
	diagnostic := signal.Diagnostic()
	if diagnostic == "" {
		t.Error("Expected Diagnostic() to return a non-empty string")
	}
}

func TestMissingInitPySignal_Diagnostic_WithFoundDirs(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a package missing __init__.py
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when __init__.py is missing")
	}

	// Diagnostic should contain the directory name
	diagnostic := signal.Diagnostic()
	if !strings.Contains(diagnostic, "mypackage") {
		t.Errorf("Expected diagnostic to contain 'mypackage', got '%s'", diagnostic)
	}
}

func TestMissingInitPySignal_MissingInitPy(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a directory with Python files but no __init__.py
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when Python package missing __init__.py")
	}
}

func TestMissingInitPySignal_NestedPackages(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create nested packages, one missing __init__.py
	os.MkdirAll("mypackage/subpackage", 0755)
	os.WriteFile("mypackage/__init__.py", []byte(""), 0644)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)
	os.WriteFile("mypackage/subpackage/module.py", []byte("def bar(): pass"), 0644)
	// Missing: mypackage/subpackage/__init__.py

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when nested package missing __init__.py")
	}
}

func TestMissingInitPySignal_IgnoresTestFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a directory with only test files (should not require __init__.py)
	os.Mkdir("tests", 0755)
	os.WriteFile("tests/test_foo.py", []byte("def test_foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when directory only has test files")
	}
}

func TestMissingInitPySignal_IgnoresSetupPy(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create root directory with only setup.py
	os.WriteFile("setup.py", []byte("from setuptools import setup"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when only setup.py exists at root")
	}
}

func TestMissingInitPySignal_SkipsVenv(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create venv directory with Python files (should be ignored)
	os.MkdirAll("venv/lib/python3.9", 0755)
	os.WriteFile("venv/lib/python3.9/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false - should skip venv directory")
	}
}

func TestMissingInitPySignal_SkipsPycache(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create __pycache__ directory (should be ignored)
	os.Mkdir("__pycache__", 0755)
	os.WriteFile("__pycache__/module.cpython-39.pyc", []byte("fake"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false - should skip __pycache__ directory")
	}
}

func TestMissingInitPySignal_MultipleMissing(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker so the signal runs
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create multiple packages missing __init__.py
	// Note: with early exit optimization, we detect the first one and stop
	os.Mkdir("package1", 0755)
	os.WriteFile("package1/module1.py", []byte("def foo(): pass"), 0644)

	os.Mkdir("package2", 0755)
	os.WriteFile("package2/module2.py", []byte("def bar(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when multiple packages missing __init__.py")
	}
}

func TestMissingInitPySignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_MISSING_INIT_PY", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_MISSING_INIT_PY")

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}

// Tests for performance optimizations

func TestMissingInitPySignal_NonPythonProject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No Python project markers - just some random directories
	os.MkdirAll("src/lib", 0755)
	os.WriteFile("src/main.go", []byte("package main"), 0644)
	os.WriteFile("src/lib/utils.go", []byte("package lib"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when not in a Python project")
	}
}

func TestMissingInitPySignal_ProjectDetection_SetupPy(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// setup.py is a project marker
	os.WriteFile("setup.py", []byte("from setuptools import setup"), 0644)
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true - setup.py should trigger project detection")
	}
}

func TestMissingInitPySignal_ProjectDetection_Pyproject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// pyproject.toml is a project marker
	os.WriteFile("pyproject.toml", []byte("[project]\nname = \"foo\""), 0644)
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true - pyproject.toml should trigger project detection")
	}
}

func TestMissingInitPySignal_ProjectDetection_RootPyFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// A .py file in root is a project marker
	os.WriteFile("app.py", []byte("print('hello')"), 0644)
	os.Mkdir("mypackage", 0755)
	os.WriteFile("mypackage/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true - .py file in root should trigger project detection")
	}
}

func TestMissingInitPySignal_DepthLimit(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create a deeply nested directory (beyond maxInitPyDepth=6)
	deepPath := "a/b/c/d/e/f/g/h/i/j"
	os.MkdirAll(deepPath, 0755)
	os.WriteFile(deepPath+"/module.py", []byte("def foo(): pass"), 0644)

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	// The deeply nested directory should be skipped due to depth limit
	if result {
		t.Error("Expected false - should skip directories beyond depth limit")
	}
}

func TestMissingInitPySignal_EarlyExit(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create a Python project marker
	os.WriteFile("requirements.txt", []byte(""), 0644)

	// Create many packages - with early exit, we should find the first and stop
	for i := 0; i < 100; i++ {
		dir := "pkg" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		os.Mkdir(dir, 0755)
		os.WriteFile(dir+"/module.py", []byte("def foo(): pass"), 0644)
	}

	signal := NewMissingInitPySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when packages missing __init__.py")
	}

	// With early exit, we should have found exactly one
	s := signal.(*MissingInitPySignal)
	if len(s.foundDirs) != 1 {
		t.Errorf("Expected exactly 1 found dir with early exit, got %d", len(s.foundDirs))
	}
}
