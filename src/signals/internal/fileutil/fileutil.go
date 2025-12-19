// Package fileutil provides bounded file read helpers for signal checks.
package fileutil

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	// ErrFileTooLarge is returned when a file exceeds the allowed size.
	ErrFileTooLarge = errors.New("file too large")
	// ErrNotRegular is returned when a path is not a regular file.
	ErrNotRegular = errors.New("not a regular file")
)

// ReadFileLimited reads at most maxBytes from a regular file.
// It rejects non-regular files and enforces the byte limit to prevent OOMs.
func ReadFileLimited(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, ErrFileTooLarge
	}

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if info, err := file.Stat(); err == nil {
		if !info.Mode().IsRegular() {
			return nil, ErrNotRegular
		}
		if info.Size() > maxBytes {
			return nil, ErrFileTooLarge
		}
	}

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrFileTooLarge
	}

	return data, nil
}

// ReadFileLimitedString returns a limited file read as a string.
func ReadFileLimitedString(path string, maxBytes int64) (string, error) {
	data, err := ReadFileLimited(path, maxBytes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
