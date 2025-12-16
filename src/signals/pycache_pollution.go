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
	maxPyCacheDepth = 6   // Max directory depth to traverse
	maxPyCacheDirs  = 500 // Max directories to visit before giving up
)

// PyCachePollutionSignal checks for __pycache__ directories that aren't properly ignored
// These should be in .gitignore to avoid polluting source control
type PyCachePollutionSignal struct {
	foundDirs []string
}

// NewPyCachePollutionSignal creates a PyCachePollutionSignal.
func NewPyCachePollutionSignal() Signal {
	return &PyCachePollutionSignal{}
}

// Name returns the human-readable name of the signal.
func (s *PyCachePollutionSignal) Name() string {
	return "PyCache Pollution"
}

// Emoji returns the emoji associated with the signal.
func (s *PyCachePollutionSignal) Emoji() string {
	return "🐍" // Snake (Python)
}

// Diagnostic returns a description of detected __pycache__ directories.
func (s *PyCachePollutionSignal) Diagnostic() string {
	if len(s.foundDirs) > 0 {
		return "__pycache__ directories found (not properly ignored in git)"
	}
	return "__pycache__ directories found (not properly ignored)"
}

// Remediation returns guidance on ignoring __pycache__ directories.
func (s *PyCachePollutionSignal) Remediation() string {
	return "Add __pycache__/ to .gitignore and run: git rm -r --cached __pycache__"
}

// Check walks the repository for tracked __pycache__ directories.
func (s *PyCachePollutionSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_PYCACHE_POLLUTION") != "" {
		return false
	}

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// Not a git repo, no issue
		return false
	}

	// Performance gate: skip if not in a Python project
	if !isPythonProjectForPyCache() {
		return false
	}

	s.foundDirs = []string{}
	dirsVisited := 0

	// Walk the current directory looking for __pycache__ directories
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

		if d.IsDir() {
			// Skip root directory
			if path == "." {
				return nil
			}

			// Check depth limit
			depth := strings.Count(path, string(filepath.Separator)) + 1
			if depth > maxPyCacheDepth {
				return filepath.SkipDir
			}

			// Check directory count limit
			dirsVisited++
			if dirsVisited > maxPyCacheDirs {
				return filepath.SkipAll
			}

			name := d.Name()

			// Skip .git directory
			if name == ".git" {
				return filepath.SkipDir
			}

			// Skip common non-Python directories
			if name == "node_modules" || name == "venv" || name == "env" ||
				strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			// Check for __pycache__ directories
			if name == "__pycache__" {
				// Check if this directory has .pyc files (pollution)
				if isTrackedByGit(path) {
					s.foundDirs = append(s.foundDirs, path)
					// Early exit: we found a problem, no need to continue
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil {
		return false
	}

	return len(s.foundDirs) > 0
}

// isPythonProjectForPyCache checks if the current directory looks like a Python project.
// This is a performance gate to avoid scanning non-Python directories.
func isPythonProjectForPyCache() bool {
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

// isTrackedByGit checks if a path is tracked by git
func isTrackedByGit(path string) bool {
	// Try to check if the directory or any files in it are tracked
	// We'll check if git ls-files shows anything in this directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	// If the directory has any .pyc files, it's likely being tracked
	// (or should be ignored but isn't)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".pyc" {
			// Found a .pyc file - this is pollution
			return true
		}
	}

	return false
}
