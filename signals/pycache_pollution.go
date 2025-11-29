package signals

import (
	"context"
	"os"
	"path/filepath"
)

// PyCachePollutionSignal checks for __pycache__ directories that aren't properly ignored
// These should be in .gitignore to avoid polluting source control
type PyCachePollutionSignal struct {
	foundDirs []string
}

func NewPyCachePollutionSignal() Signal {
	return &PyCachePollutionSignal{}
}

func (s *PyCachePollutionSignal) Name() string {
	return "PyCache Pollution"
}

func (s *PyCachePollutionSignal) Emoji() string {
	return "🐍" // Snake (Python)
}

func (s *PyCachePollutionSignal) Diagnostic() string {
	if len(s.foundDirs) > 0 {
		return "__pycache__ directories found (not properly ignored in git)"
	}
	return "__pycache__ directories found (not properly ignored)"
}

func (s *PyCachePollutionSignal) Remediation() string {
	return "Add __pycache__/ to .gitignore and run: git rm -r --cached __pycache__"
}

func (s *PyCachePollutionSignal) Check(ctx context.Context) bool {
	s.foundDirs = []string{}

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// Not a git repo, no issue
		return false
	}

	// Walk the current directory looking for __pycache__ directories
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check for __pycache__ directories
		if info.IsDir() && info.Name() == "__pycache__" {
			// Check if this directory is tracked by git
			if isTrackedByGit(path) {
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
