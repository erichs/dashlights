package signals

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Performance limits to prevent runaway scans on large directory trees
const (
	maxInitPyDepth = 6   // Max directory depth to traverse
	maxInitPyDirs  = 500 // Max directories to visit before giving up
)

// MissingInitPySignal checks for Python packages missing __init__.py
// Directories with .py files but no __init__.py will cause import failures
type MissingInitPySignal struct {
	foundDirs []string
}

// NewMissingInitPySignal creates a MissingInitPySignal.
func NewMissingInitPySignal() Signal {
	return &MissingInitPySignal{}
}

// Name returns the human-readable name of the signal.
func (s *MissingInitPySignal) Name() string {
	return "Missing __init__.py"
}

// Emoji returns the emoji associated with the signal.
func (s *MissingInitPySignal) Emoji() string {
	return "📁" // Folder (missing file in package)
}

// Diagnostic returns a description of Python packages missing __init__.py.
func (s *MissingInitPySignal) Diagnostic() string {
	if len(s.foundDirs) > 0 {
		return "Python package directories missing __init__.py: " + s.foundDirs[0]
	}
	return "Python package directories missing __init__.py (imports will fail)"
}

// Remediation returns guidance on adding __init__.py files.
func (s *MissingInitPySignal) Remediation() string {
	return "Add __init__.py to package directories (can be empty)"
}

// Check walks the repository for Python-style package directories missing __init__.py.
func (s *MissingInitPySignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_MISSING_INIT_PY") != "" {
		return false
	}

	// Performance gate: skip if not in a Python project
	if !isPythonProject() {
		return false
	}

	s.foundDirs = []string{}
	dirsVisited := 0

	// Walk the current directory looking for Python packages
	// Use WalkDir for better performance (avoids extra Lstat calls)
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return filepath.SkipAll
		default:
		}

		if err != nil {
			return nil // Skip errors
		}

		// Only process directories
		if d.IsDir() {
			// Don't check the root directory itself, but continue walking
			// IMPORTANT: Check this BEFORE checking name, because "." starts with "."!
			if path == "." {
				return nil
			}

			// Check depth limit (count path separators)
			depth := strings.Count(path, string(filepath.Separator)) + 1
			if depth > maxInitPyDepth {
				return filepath.SkipDir
			}

			// Check directory count limit
			dirsVisited++
			if dirsVisited > maxInitPyDirs {
				return filepath.SkipAll
			}

			name := d.Name()

			// Skip hidden directories and common non-package directories
			if strings.HasPrefix(name, ".") ||
				name == "__pycache__" ||
				name == "venv" ||
				name == "env" ||
				name == "node_modules" ||
				name == "dist" ||
				name == "build" {
				return filepath.SkipDir
			}

			// Check if this directory looks like a Python package
			isPkg := isPythonPackage(path)
			hasInit := hasInitPy(path)
			if isPkg && !hasInit {
				s.foundDirs = append(s.foundDirs, path)
				// Early exit: we found a problem, no need to continue
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil {
		return false
	}

	return len(s.foundDirs) > 0
}

// isPythonProject checks if the current directory looks like a Python project.
// This is a performance gate to avoid scanning non-Python directories.
func isPythonProject() bool {
	// Check for common Python project markers
	markers := []string{"setup.py", "pyproject.toml", "requirements.txt"}
	for _, m := range markers {
		if _, err := os.Stat(m); err == nil {
			return true
		}
	}

	// Check for any .py files in the root directory
	entries, err := os.ReadDir(".")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".py") {
			return true
		}
	}

	return false
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
