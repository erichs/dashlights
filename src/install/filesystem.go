// Package install provides installation automation for dashlights.
// It handles shell prompt integration and AI agent hook installation.
package install

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Filesystem abstracts file operations for testability.
type Filesystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
	Exists(path string) bool
	MkdirAll(path string, perm os.FileMode) error
	Rename(src, dst string) error
	UserHomeDir() (string, error)
	Getenv(key string) string
}

// OSFilesystem implements Filesystem using real OS operations.
type OSFilesystem struct{}

// ReadFile reads the contents of a file.
func (f *OSFilesystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(path))
}

// WriteFile writes data to a file with the given permissions.
func (f *OSFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Stat returns file info for the given path.
func (f *OSFilesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Exists returns true if the file exists.
func (f *OSFilesystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MkdirAll creates a directory and all parent directories.
func (f *OSFilesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Rename renames (moves) a file.
func (f *OSFilesystem) Rename(src, dst string) error {
	return os.Rename(src, dst)
}

// UserHomeDir returns the user's home directory.
func (f *OSFilesystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// Getenv returns the value of an environment variable.
func (f *OSFilesystem) Getenv(key string) string {
	return os.Getenv(key)
}

// MockFilesystem implements Filesystem for testing.
type MockFilesystem struct {
	Files   map[string][]byte
	Modes   map[string]os.FileMode
	EnvVars map[string]string
	HomeDir string

	// Error simulation
	ReadFileErr  error
	WriteFileErr error
	StatErr      error
	MkdirAllErr  error
	RenameErr    error
	HomeDirErr   error
}

// NewMockFilesystem creates a new mock filesystem for testing.
func NewMockFilesystem() *MockFilesystem {
	return &MockFilesystem{
		Files:   make(map[string][]byte),
		Modes:   make(map[string]os.FileMode),
		EnvVars: make(map[string]string),
		HomeDir: "/home/testuser",
	}
}

// ReadFile reads from the mock filesystem.
func (f *MockFilesystem) ReadFile(path string) ([]byte, error) {
	if f.ReadFileErr != nil {
		return nil, f.ReadFileErr
	}
	content, ok := f.Files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return content, nil
}

// WriteFile writes to the mock filesystem.
func (f *MockFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if f.WriteFileErr != nil {
		return f.WriteFileErr
	}
	f.Files[path] = data
	f.Modes[path] = perm
	return nil
}

// Stat returns mock file info.
func (f *MockFilesystem) Stat(path string) (os.FileInfo, error) {
	if f.StatErr != nil {
		return nil, f.StatErr
	}
	content, ok := f.Files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	mode := f.Modes[path]
	if mode == 0 {
		mode = 0644
	}
	return &mockFileInfo{
		name: filepath.Base(path),
		size: int64(len(content)),
		mode: mode,
	}, nil
}

// Exists checks if a file exists in the mock filesystem.
func (f *MockFilesystem) Exists(path string) bool {
	_, ok := f.Files[path]
	return ok
}

// MkdirAll is a no-op in the mock filesystem.
func (f *MockFilesystem) MkdirAll(path string, perm os.FileMode) error {
	if f.MkdirAllErr != nil {
		return f.MkdirAllErr
	}
	return nil
}

// Rename moves a file in the mock filesystem.
func (f *MockFilesystem) Rename(src, dst string) error {
	if f.RenameErr != nil {
		return f.RenameErr
	}
	content, ok := f.Files[src]
	if !ok {
		return os.ErrNotExist
	}
	f.Files[dst] = content
	if mode, ok := f.Modes[src]; ok {
		f.Modes[dst] = mode
	}
	delete(f.Files, src)
	delete(f.Modes, src)
	return nil
}

// UserHomeDir returns the mock home directory.
func (f *MockFilesystem) UserHomeDir() (string, error) {
	if f.HomeDirErr != nil {
		return "", f.HomeDirErr
	}
	return f.HomeDir, nil
}

// Getenv returns a mock environment variable.
func (f *MockFilesystem) Getenv(key string) string {
	return f.EnvVars[key]
}

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *mockFileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi *mockFileInfo) Sys() interface{}   { return nil }
func (fi *mockFileInfo) ModTime() time.Time { return time.Time{} }
