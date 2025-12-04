// Package homedirutil provides safe utilities for working with home directory paths.
package homedirutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrTraversalAttempt is returned when a path contains directory traversal patterns.
var ErrTraversalAttempt = errors.New("directory traversal attempt detected")

// ErrRelativePath is returned when a home directory path is not absolute.
var ErrRelativePath = errors.New("home directory must be an absolute path")

// ErrNoHomeDir is returned when the home directory cannot be determined.
var ErrNoHomeDir = errors.New("could not determine home directory")

// SafeHomePath returns a safe, sanitized path under the user's home directory.
// It performs the following validations:
//   - Gets the user's home directory via os.UserHomeDir()
//   - Rejects home directories containing ".." (directory traversal)
//   - Cleans the path to resolve any . or .. components
//   - Ensures the home directory is absolute
//   - Joins the provided path components under the sanitized home directory
//
// Returns the full path and nil error on success, or empty string and an error
// if validation fails.
func SafeHomePath(pathComponents ...string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", ErrNoHomeDir
	}

	return SafeHomePathFrom(homeDir, pathComponents...)
}

// SafeHomePathFrom performs the same sanitization as SafeHomePath but accepts
// a pre-determined home directory. This is useful for testing or when the
// home directory is obtained through other means.
func SafeHomePathFrom(homeDir string, pathComponents ...string) (string, error) {
	// Validate that homeDir doesn't contain suspicious patterns
	if strings.Contains(homeDir, "..") {
		return "", ErrTraversalAttempt
	}

	// Clean the path to resolve any . or .. components
	sanitizedHome := filepath.Clean(homeDir)

	// Ensure the sanitized path is absolute (home directories should always be absolute)
	if !filepath.IsAbs(sanitizedHome) {
		return "", ErrRelativePath
	}

	// Join the path components under the sanitized home directory
	fullPath := filepath.Join(append([]string{sanitizedHome}, pathComponents...)...)

	return fullPath, nil
}
