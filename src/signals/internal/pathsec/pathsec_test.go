package pathsec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoinPath(t *testing.T) {
	tests := []struct {
		name      string
		baseDir   string
		filename  string
		wantErr   bool
		errType   error
		checkPath bool // if true, verify path is within baseDir
	}{
		{
			name:      "valid simple filename",
			baseDir:   "/base/dir",
			filename:  "file.txt",
			wantErr:   false,
			checkPath: true,
		},
		{
			name:     "directory traversal with ..",
			baseDir:  "/base/dir",
			filename: "../etc/passwd",
			wantErr:  true,
			errType:  ErrTraversal,
		},
		{
			name:     "directory traversal .. only",
			baseDir:  "/base/dir",
			filename: "..",
			wantErr:  true,
			errType:  ErrTraversal,
		},
		{
			name:     "path with forward slash",
			baseDir:  "/base/dir",
			filename: "subdir/file.txt",
			wantErr:  true,
			errType:  ErrTraversal,
		},
		{
			name:     "path with backslash",
			baseDir:  "/base/dir",
			filename: "subdir\\file.txt",
			wantErr:  true,
			errType:  ErrTraversal,
		},
		{
			name:      "filename with dots is OK",
			baseDir:   "/base/dir",
			filename:  "file.tar.gz",
			wantErr:   false,
			checkPath: true,
		},
		{
			name:      "hidden file is OK",
			baseDir:   "/base/dir",
			filename:  ".hidden",
			wantErr:   false,
			checkPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := SafeJoinPath(tt.baseDir, tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SafeJoinPath() expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("SafeJoinPath() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("SafeJoinPath() unexpected error: %v", err)
				return
			}

			if tt.checkPath {
				// Verify path is within base directory
				if !strings.HasPrefix(path, tt.baseDir) {
					t.Errorf("SafeJoinPath() path %q not within base %q", path, tt.baseDir)
				}
			}
		})
	}
}

func TestIsSafeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple filename", "file.txt", true},
		{"hidden file", ".gitignore", true},
		{"dots in filename", "file.tar.gz", true},
		{"empty string", "", false},
		{"directory traversal", "..", false},
		{"traversal in path", "../etc", false},
		{"forward slash", "dir/file", false},
		{"backslash", "dir\\file", false},
		{"embedded traversal", "foo/../bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeName(tt.input); got != tt.expected {
				t.Errorf("IsSafeName(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple path", ".git/hooks", true},
		{"absolute path", "/usr/local/bin", true},
		{"relative path", "src/signals", true},
		{"empty string", "", false},
		{"traversal ..", "..", false},
		{"traversal in path", "../etc/passwd", false},
		{"embedded traversal", "/foo/../bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPath(tt.input); got != tt.expected {
				t.Errorf("IsValidPath(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSafeJoinAndOpen(t *testing.T) {
	// Create a temp directory with a test file
	tmpDir := t.TempDir()
	testFile := "test.txt"
	testPath := filepath.Join(tmpDir, testFile)
	if err := os.WriteFile(testPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("valid file opens successfully", func(t *testing.T) {
		file, err := SafeJoinAndOpen(tmpDir, testFile)
		if err != nil {
			t.Errorf("SafeJoinAndOpen() unexpected error: %v", err)
			return
		}
		file.Close()
	})

	t.Run("traversal attempt fails", func(t *testing.T) {
		_, err := SafeJoinAndOpen(tmpDir, "../etc/passwd")
		if err == nil {
			t.Errorf("SafeJoinAndOpen() expected error for traversal attempt")
		}
	})

	t.Run("nonexistent file fails", func(t *testing.T) {
		_, err := SafeJoinAndOpen(tmpDir, "nonexistent.txt")
		if err == nil {
			t.Errorf("SafeJoinAndOpen() expected error for nonexistent file")
		}
	})
}
