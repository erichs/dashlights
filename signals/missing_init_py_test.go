package signals

import (
	"context"
	"os"
	"testing"
)

func TestMissingInitPySignal_NoPythonFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

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

func TestMissingInitPySignal_MissingInitPy(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

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

	// Create multiple packages missing __init__.py
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
