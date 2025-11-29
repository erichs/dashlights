package signals

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// MissingInitPySignal checks for Python packages missing __init__.py
// Directories with .py files but no __init__.py will cause import failures
type MissingInitPySignal struct {
	foundDirs []string
}

func NewMissingInitPySignal() Signal {
	return &MissingInitPySignal{}
}

func (s *MissingInitPySignal) Name() string {
	return "Missing __init__.py"
}

func (s *MissingInitPySignal) Emoji() string {
	return "📁" // Folder (missing file in package)
}

func (s *MissingInitPySignal) Diagnostic() string {
	if len(s.foundDirs) > 0 {
		return "Python package directories missing __init__.py: " + s.foundDirs[0]
	}
	return "Python package directories missing __init__.py (imports will fail)"
}

func (s *MissingInitPySignal) Remediation() string {
	return "Add __init__.py to package directories (can be empty)"
}

func (s *MissingInitPySignal) Check(ctx context.Context) bool {
	s.foundDirs = []string{}

	// Walk the current directory looking for Python packages
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common non-package directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") ||
				name == "__pycache__" ||
				name == "venv" ||
				name == "env" ||
				name == "node_modules" ||
				name == "dist" ||
				name == "build" {
				return filepath.SkipDir
			}

			// Skip the root directory itself
			if path == "." {
				return nil
			}

			// Check if this directory looks like a Python package
			isPkg := isPythonPackage(path)
			hasInit := hasInitPy(path)
			if isPkg && !hasInit {
				s.foundDirs = append(s.foundDirs, path)
			}
		}

		return nil
	})

	if err != nil {
		return false
	}

	return len(s.foundDirs) > 0
}

// isPythonPackage checks if a directory contains .py files
func isPythonPackage(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".py") {
			name := entry.Name()
			// Ignore test files (they don't need to be in a package)
			if strings.HasPrefix(name, "test_") {
				continue
			}
			// Ignore setup.py (it's a build script, not a module)
			if name == "setup.py" {
				continue
			}
			// Found a real Python module
			return true
		}
	}

	return false
}

// hasInitPy checks if a directory contains __init__.py
func hasInitPy(dir string) bool {
	initPath := filepath.Join(dir, "__init__.py")
	_, err := os.Stat(initPath)
	return err == nil
}
