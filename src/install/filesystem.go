// Package install provides installation automation for dashlights.
// It handles shell prompt integration and AI agent hook installation.
package install

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Filesystem abstracts file operations for testability.
type Filesystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error) // for symlink detection
	Exists(path string) bool
	MkdirAll(path string, perm os.FileMode) error
	Rename(src, dst string) error
	UserHomeDir() (string, error)
	Getenv(key string) string
	Executable() (string, error)    // returns path of running binary
	CopyFile(src, dst string) error // copy file preserving permissions
	Chmod(path string, mode os.FileMode) error
	SplitPath() []string         // splits $PATH by os.PathListSeparator
	IsWritable(path string) bool // checks if directory is writable
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

// Lstat returns file info without following symlinks.
func (f *OSFilesystem) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

// Executable returns the path of the running binary.
func (f *OSFilesystem) Executable() (string, error) {
	return os.Executable()
}

// CopyFile copies a file from src to dst, preserving permissions.
func (f *OSFilesystem) CopyFile(src, dst string) error {
	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(filepath.Clean(dst), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// Chmod changes the mode of a file.
func (f *OSFilesystem) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// SplitPath returns the directories in $PATH.
func (f *OSFilesystem) SplitPath() []string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}
	return filepath.SplitList(pathEnv)
}

// IsWritable checks if a directory is writable by the current user.
func (f *OSFilesystem) IsWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	// Try to create a temp file to test writability
	testFile := filepath.Join(filepath.Clean(path), ".dashlights-write-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	if err := file.Close(); err != nil {
		return false
	}
	if err := os.Remove(testFile); err != nil {
		// File was created but couldn't be removed - still writable
		return true
	}
	return true
}

// MockFilesystem implements Filesystem for testing.
type MockFilesystem struct {
	Files   map[string][]byte
	Modes   map[string]os.FileMode
	EnvVars map[string]string
	HomeDir string

	// Binary installation support
	ExecutablePath string            // path returned by Executable()
	Symlinks       map[string]string // path -> target (for symlink detection)
	WritableDirs   map[string]bool   // dir path -> writable
	PathEnv        string            // value for PATH (overrides EnvVars["PATH"])

	// Error simulation
	ReadFileErr   error
	WriteFileErr  error
	StatErr       error
	MkdirAllErr   error
	RenameErr     error
	HomeDirErr    error
	ExecutableErr error
	CopyFileErr   error
	ChmodErr      error
	LstatErr      error
}

// NewMockFilesystem creates a new mock filesystem for testing.
func NewMockFilesystem() *MockFilesystem {
	return &MockFilesystem{
		Files:          make(map[string][]byte),
		Modes:          make(map[string]os.FileMode),
		EnvVars:        make(map[string]string),
		HomeDir:        "/home/testuser",
		Symlinks:       make(map[string]string),
		WritableDirs:   make(map[string]bool),
		ExecutablePath: "/usr/local/bin/dashlights",
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

// Lstat returns mock file info, detecting symlinks.
func (f *MockFilesystem) Lstat(path string) (os.FileInfo, error) {
	if f.LstatErr != nil {
		return nil, f.LstatErr
	}
	// Check if it's a symlink
	if target, isSymlink := f.Symlinks[path]; isSymlink {
		return &mockFileInfo{
			name: filepath.Base(path),
			size: int64(len(target)),
			mode: os.ModeSymlink | 0777,
		}, nil
	}
	// Fall back to regular Stat behavior
	return f.Stat(path)
}

// Executable returns the mock executable path.
func (f *MockFilesystem) Executable() (string, error) {
	if f.ExecutableErr != nil {
		return "", f.ExecutableErr
	}
	return f.ExecutablePath, nil
}

// CopyFile copies a file in the mock filesystem.
func (f *MockFilesystem) CopyFile(src, dst string) error {
	if f.CopyFileErr != nil {
		return f.CopyFileErr
	}
	content, ok := f.Files[src]
	if !ok {
		return os.ErrNotExist
	}
	f.Files[dst] = make([]byte, len(content))
	copy(f.Files[dst], content)
	if mode, ok := f.Modes[src]; ok {
		f.Modes[dst] = mode
	} else {
		f.Modes[dst] = 0755 // default for binary
	}
	return nil
}

// Chmod changes the mode of a file in the mock filesystem.
func (f *MockFilesystem) Chmod(path string, mode os.FileMode) error {
	if f.ChmodErr != nil {
		return f.ChmodErr
	}
	if _, ok := f.Files[path]; !ok {
		return os.ErrNotExist
	}
	f.Modes[path] = mode
	return nil
}

// SplitPath returns the directories in the mock PATH.
func (f *MockFilesystem) SplitPath() []string {
	pathEnv := f.PathEnv
	if pathEnv == "" {
		pathEnv = f.EnvVars["PATH"]
	}
	if pathEnv == "" {
		return nil
	}
	return strings.Split(pathEnv, string(os.PathListSeparator))
}

// IsWritable checks if a directory is writable in the mock filesystem.
func (f *MockFilesystem) IsWritable(path string) bool {
	writable, ok := f.WritableDirs[path]
	return ok && writable
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
