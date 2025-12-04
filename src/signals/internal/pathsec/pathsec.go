// Package pathsec provides path security utilities to prevent directory traversal attacks.
package pathsec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrTraversal is returned when a path contains directory traversal patterns.
var ErrTraversal = errors.New("directory traversal attempt detected")

// ErrPathEscape is returned when a joined path escapes the base directory.
var ErrPathEscape = errors.New("path escapes base directory")

// SafeJoinPath safely joins a base directory and filename, ensuring the result
// stays within the base directory. This prevents directory traversal attacks (G304).
//
// It performs the following validations:
//   - Cleans the filename to normalize it
//   - Rejects filenames containing path separators or ".." components
//   - Joins and cleans the full path
//   - Verifies the result stays within the base directory
//
// Returns the joined path and nil on success, or empty string and an error if validation fails.
func SafeJoinPath(baseDir, filename string) (string, error) {
	// Clean the filename
	filename = filepath.Clean(filename)

	// Reject any path components
	if strings.ContainsAny(filename, `/\`) || filename == ".." || strings.HasPrefix(filename, "..") {
		return "", ErrTraversal
	}

	// Join and clean the full path
	fullPath := filepath.Join(baseDir, filename)
	fullPath = filepath.Clean(fullPath)

	// Verify the result is still within the base directory
	if !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) && fullPath != baseDir {
		return "", ErrPathEscape
	}

	return fullPath, nil
}

// IsSafeName checks if a name is safe to use as a filename or path component.
// Returns true if the name:
//   - Does not contain ".." (directory traversal)
//   - Does not contain "/" or "\" (path separators)
//   - Is not empty
//
// This is useful for validating user-provided filenames before use.
func IsSafeName(name string) bool {
	if name == "" {
		return false
	}

	if strings.Contains(name, "..") {
		return false
	}

	if strings.ContainsAny(name, `/\`) {
		return false
	}

	return true
}

// IsValidPath validates that a path is safe to use.
// Returns true if the path:
//   - Is not empty
//   - Does not contain ".." (before or after cleaning)
//
// This is useful for validating paths read from configuration files.
func IsValidPath(path string) bool {
	if path == "" {
		return false
	}

	// Check for directory traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return false
	}

	// Clean the path and check again
	cleaned := filepath.Clean(path)
	return !strings.Contains(cleaned, "..")
}

// SafeJoinAndOpen safely joins a base directory and filename, then opens the file.
// This combines SafeJoinPath with os.Open and uses filepath.Clean for gosec G304.
//
// Returns the opened file and nil on success, or nil and an error if validation
// or opening fails. The caller is responsible for closing the file.
func SafeJoinAndOpen(baseDir, filename string) (*os.File, error) {
	path, err := SafeJoinPath(baseDir, filename)
	if err != nil {
		return nil, err
	}

	// filepath.Clean for gosec G304 - path is already validated by SafeJoinPath
	return os.Open(filepath.Clean(path))
}
