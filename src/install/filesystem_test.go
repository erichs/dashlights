package install

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestOSFilesystem tests the real filesystem implementation
func TestOSFilesystem_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	// Write test file
	err := os.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fs := &OSFilesystem{}
	data, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected content '%s', got '%s'", string(content), string(data))
	}
}

func TestOSFilesystem_ReadFile_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.txt")

	fs := &OSFilesystem{}
	_, err := fs.ReadFile(testFile)
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestOSFilesystem_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "write.txt")
	content := []byte("write test")

	fs := &OSFilesystem{}
	err := fs.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected content '%s', got '%s'", string(content), string(data))
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("Expected permissions 0644, got %o", info.Mode().Perm())
	}
}

func TestOSFilesystem_WriteFile_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "overwrite.txt")

	// Write initial content
	err := os.WriteFile(testFile, []byte("initial"), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Overwrite with new content
	newContent := []byte("overwritten")
	fs := &OSFilesystem{}
	err = fs.WriteFile(testFile, newContent, 0600)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Verify new content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != string(newContent) {
		t.Errorf("Expected content '%s', got '%s'", string(newContent), string(data))
	}
}

func TestOSFilesystem_Stat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "stat.txt")
	content := []byte("stat test")

	err := os.WriteFile(testFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fs := &OSFilesystem{}
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), info.Size())
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
	if info.IsDir() {
		t.Error("Expected file, not directory")
	}
}

func TestOSFilesystem_Stat_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	fs := &OSFilesystem{}
	info, err := fs.Stat(testDir)
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected directory")
	}
}

func TestOSFilesystem_Stat_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.txt")

	fs := &OSFilesystem{}
	_, err := fs.Stat(testFile)
	if err == nil {
		t.Error("Expected error when stat-ing non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestOSFilesystem_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "exists.txt")

	fs := &OSFilesystem{}

	// File doesn't exist yet
	if fs.Exists(testFile) {
		t.Error("Expected false for non-existent file")
	}

	// Create file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// File now exists
	if !fs.Exists(testFile) {
		t.Error("Expected true for existing file")
	}
}

func TestOSFilesystem_Exists_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")

	fs := &OSFilesystem{}

	// Directory doesn't exist yet
	if fs.Exists(testDir) {
		t.Error("Expected false for non-existent directory")
	}

	// Create directory
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Directory now exists
	if !fs.Exists(testDir) {
		t.Error("Expected true for existing directory")
	}
}

func TestOSFilesystem_MkdirAll(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "parent", "child", "grandchild")

	fs := &OSFilesystem{}
	err := fs.MkdirAll(testDir, 0755)
	if err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected directory")
	}
}

func TestOSFilesystem_MkdirAll_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "existing")

	// Create directory first
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// MkdirAll should succeed even if directory exists
	fs := &OSFilesystem{}
	err = fs.MkdirAll(testDir, 0755)
	if err != nil {
		t.Errorf("MkdirAll failed on existing directory: %v", err)
	}
}

func TestOSFilesystem_Rename(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := []byte("rename test")

	// Create source file
	err := os.WriteFile(srcFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	fs := &OSFilesystem{}
	err = fs.Rename(srcFile, dstFile)
	if err != nil {
		t.Errorf("Rename failed: %v", err)
	}

	// Verify source file no longer exists
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Expected source file to not exist after rename")
	}

	// Verify destination file exists with correct content
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected content '%s', got '%s'", string(content), string(data))
	}
}

func TestOSFilesystem_Rename_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "nonexistent.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	fs := &OSFilesystem{}
	err := fs.Rename(srcFile, dstFile)
	if err == nil {
		t.Error("Expected error when renaming non-existent file")
	}
}

func TestOSFilesystem_UserHomeDir(t *testing.T) {
	fs := &OSFilesystem{}
	home, err := fs.UserHomeDir()
	if err != nil {
		t.Errorf("UserHomeDir failed: %v", err)
	}
	if home == "" {
		t.Error("Expected non-empty home directory")
	}

	// Verify it matches os.UserHomeDir
	expectedHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir failed: %v", err)
	}
	if home != expectedHome {
		t.Errorf("Expected home '%s', got '%s'", expectedHome, home)
	}
}

func TestOSFilesystem_Getenv(t *testing.T) {
	testKey := "DASHLIGHTS_TEST_VAR"
	testValue := "test_value_123"

	// Set environment variable
	err := os.Setenv(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv(testKey)

	fs := &OSFilesystem{}
	value := fs.Getenv(testKey)
	if value != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, value)
	}
}

func TestOSFilesystem_Getenv_NonExistent(t *testing.T) {
	fs := &OSFilesystem{}
	value := fs.Getenv("DASHLIGHTS_NONEXISTENT_VAR")
	if value != "" {
		t.Errorf("Expected empty string for non-existent variable, got '%s'", value)
	}
}

// TestMockFilesystem tests the mock filesystem implementation
func TestNewMockFilesystem(t *testing.T) {
	fs := NewMockFilesystem()

	if fs.Files == nil {
		t.Error("Expected Files map to be initialized")
	}
	if fs.Modes == nil {
		t.Error("Expected Modes map to be initialized")
	}
	if fs.EnvVars == nil {
		t.Error("Expected EnvVars map to be initialized")
	}
	if fs.HomeDir != "/home/testuser" {
		t.Errorf("Expected default HomeDir '/home/testuser', got '%s'", fs.HomeDir)
	}
	if len(fs.Files) != 0 {
		t.Errorf("Expected empty Files map, got %d entries", len(fs.Files))
	}
}

func TestMockFilesystem_ReadFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		path        string
		expectData  []byte
		expectError bool
	}{
		{
			name: "read existing file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/test/file.txt"] = []byte("content")
			},
			path:        "/test/file.txt",
			expectData:  []byte("content"),
			expectError: false,
		},
		{
			name:        "read non-existent file",
			setup:       func(fs *MockFilesystem) {},
			path:        "/nonexistent.txt",
			expectData:  nil,
			expectError: true,
		},
		{
			name: "read with error simulation",
			setup: func(fs *MockFilesystem) {
				fs.ReadFileErr = errors.New("simulated error")
			},
			path:        "/any/path.txt",
			expectData:  nil,
			expectError: true,
		},
		{
			name: "read empty file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/empty.txt"] = []byte("")
			},
			path:        "/empty.txt",
			expectData:  []byte(""),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			data, err := fs.ReadFile(tt.path)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && string(data) != string(tt.expectData) {
				t.Errorf("Expected data '%s', got '%s'", string(tt.expectData), string(data))
			}
		})
	}
}

func TestMockFilesystem_ReadFile_ErrorIsNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	_, err := fs.ReadFile("/nonexistent.txt")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestMockFilesystem_WriteFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		path        string
		data        []byte
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "write new file",
			setup:       func(fs *MockFilesystem) {},
			path:        "/new/file.txt",
			data:        []byte("new content"),
			perm:        0644,
			expectError: false,
		},
		{
			name: "overwrite existing file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/existing.txt"] = []byte("old content")
				fs.Modes["/existing.txt"] = 0600
			},
			path:        "/existing.txt",
			data:        []byte("new content"),
			perm:        0644,
			expectError: false,
		},
		{
			name: "write with error simulation",
			setup: func(fs *MockFilesystem) {
				fs.WriteFileErr = errors.New("simulated error")
			},
			path:        "/any/path.txt",
			data:        []byte("content"),
			perm:        0644,
			expectError: true,
		},
		{
			name:        "write empty file",
			setup:       func(fs *MockFilesystem) {},
			path:        "/empty.txt",
			data:        []byte(""),
			perm:        0644,
			expectError: false,
		},
		{
			name:        "write with different permissions",
			setup:       func(fs *MockFilesystem) {},
			path:        "/file.txt",
			data:        []byte("content"),
			perm:        0755,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			err := fs.WriteFile(tt.path, tt.data, tt.perm)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				// Verify file was written
				if string(fs.Files[tt.path]) != string(tt.data) {
					t.Errorf("Expected data '%s', got '%s'", string(tt.data), string(fs.Files[tt.path]))
				}
				// Verify permissions were stored
				if fs.Modes[tt.path] != tt.perm {
					t.Errorf("Expected permissions %o, got %o", tt.perm, fs.Modes[tt.path])
				}
			}
		})
	}
}

func TestMockFilesystem_Stat(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		path        string
		expectError bool
		expectSize  int64
		expectMode  os.FileMode
	}{
		{
			name: "stat existing file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/test.txt"] = []byte("test content")
				fs.Modes["/test.txt"] = 0644
			},
			path:        "/test.txt",
			expectError: false,
			expectSize:  12, // len("test content")
			expectMode:  0644,
		},
		{
			name: "stat file with default mode",
			setup: func(fs *MockFilesystem) {
				fs.Files["/default.txt"] = []byte("content")
				// Don't set mode, should default to 0644
			},
			path:        "/default.txt",
			expectError: false,
			expectSize:  7, // len("content")
			expectMode:  0644,
		},
		{
			name:        "stat non-existent file",
			setup:       func(fs *MockFilesystem) {},
			path:        "/nonexistent.txt",
			expectError: true,
		},
		{
			name: "stat with error simulation",
			setup: func(fs *MockFilesystem) {
				fs.StatErr = errors.New("simulated error")
			},
			path:        "/any/path.txt",
			expectError: true,
		},
		{
			name: "stat empty file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/empty.txt"] = []byte("")
				fs.Modes["/empty.txt"] = 0600
			},
			path:        "/empty.txt",
			expectError: false,
			expectSize:  0,
			expectMode:  0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			info, err := fs.Stat(tt.path)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if info.Size() != tt.expectSize {
					t.Errorf("Expected size %d, got %d", tt.expectSize, info.Size())
				}
				if info.Mode() != tt.expectMode {
					t.Errorf("Expected mode %o, got %o", tt.expectMode, info.Mode())
				}
				// Verify name is basename
				expectedName := filepath.Base(tt.path)
				if info.Name() != expectedName {
					t.Errorf("Expected name '%s', got '%s'", expectedName, info.Name())
				}
			}
		})
	}
}

func TestMockFilesystem_Stat_ErrorIsNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	_, err := fs.Stat("/nonexistent.txt")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestMockFilesystem_Exists(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*MockFilesystem)
		path   string
		expect bool
	}{
		{
			name: "file exists",
			setup: func(fs *MockFilesystem) {
				fs.Files["/exists.txt"] = []byte("content")
			},
			path:   "/exists.txt",
			expect: true,
		},
		{
			name:   "file does not exist",
			setup:  func(fs *MockFilesystem) {},
			path:   "/nonexistent.txt",
			expect: false,
		},
		{
			name: "empty file exists",
			setup: func(fs *MockFilesystem) {
				fs.Files["/empty.txt"] = []byte("")
			},
			path:   "/empty.txt",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			exists := fs.Exists(tt.path)
			if exists != tt.expect {
				t.Errorf("Expected %v, got %v", tt.expect, exists)
			}
		})
	}
}

func TestMockFilesystem_MkdirAll(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		path        string
		perm        os.FileMode
		expectError bool
	}{
		{
			name:        "create directory",
			setup:       func(fs *MockFilesystem) {},
			path:        "/test/dir",
			perm:        0755,
			expectError: false,
		},
		{
			name: "create with error simulation",
			setup: func(fs *MockFilesystem) {
				fs.MkdirAllErr = errors.New("simulated error")
			},
			path:        "/any/dir",
			perm:        0755,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			err := fs.MkdirAll(tt.path, tt.perm)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMockFilesystem_Rename(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		src         string
		dst         string
		expectError bool
	}{
		{
			name: "rename existing file",
			setup: func(fs *MockFilesystem) {
				fs.Files["/old.txt"] = []byte("content")
				fs.Modes["/old.txt"] = 0644
			},
			src:         "/old.txt",
			dst:         "/new.txt",
			expectError: false,
		},
		{
			name:        "rename non-existent file",
			setup:       func(fs *MockFilesystem) {},
			src:         "/nonexistent.txt",
			dst:         "/new.txt",
			expectError: true,
		},
		{
			name: "rename with error simulation",
			setup: func(fs *MockFilesystem) {
				fs.RenameErr = errors.New("simulated error")
			},
			src:         "/any.txt",
			dst:         "/other.txt",
			expectError: true,
		},
		{
			name: "rename preserves content and mode",
			setup: func(fs *MockFilesystem) {
				fs.Files["/src.txt"] = []byte("test content")
				fs.Modes["/src.txt"] = 0600
			},
			src:         "/src.txt",
			dst:         "/dst.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			// Save original content and mode if source exists
			var originalContent []byte
			var originalMode os.FileMode
			if data, ok := fs.Files[tt.src]; ok {
				originalContent = data
				originalMode = fs.Modes[tt.src]
			}

			err := fs.Rename(tt.src, tt.dst)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				// Verify source no longer exists
				if _, ok := fs.Files[tt.src]; ok {
					t.Error("Expected source file to be removed")
				}
				if _, ok := fs.Modes[tt.src]; ok {
					t.Error("Expected source mode to be removed")
				}
				// Verify destination has correct content and mode
				if string(fs.Files[tt.dst]) != string(originalContent) {
					t.Errorf("Expected content '%s', got '%s'", string(originalContent), string(fs.Files[tt.dst]))
				}
				if originalMode != 0 && fs.Modes[tt.dst] != originalMode {
					t.Errorf("Expected mode %o, got %o", originalMode, fs.Modes[tt.dst])
				}
			}
		})
	}
}

func TestMockFilesystem_Rename_ErrorIsNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	err := fs.Rename("/nonexistent.txt", "/new.txt")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestMockFilesystem_UserHomeDir(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockFilesystem)
		expectHome  string
		expectError bool
	}{
		{
			name:        "default home directory",
			setup:       func(fs *MockFilesystem) {},
			expectHome:  "/home/testuser",
			expectError: false,
		},
		{
			name: "custom home directory",
			setup: func(fs *MockFilesystem) {
				fs.HomeDir = "/custom/home"
			},
			expectHome:  "/custom/home",
			expectError: false,
		},
		{
			name: "error simulation",
			setup: func(fs *MockFilesystem) {
				fs.HomeDirErr = errors.New("simulated error")
			},
			expectHome:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			home, err := fs.UserHomeDir()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && home != tt.expectHome {
				t.Errorf("Expected home '%s', got '%s'", tt.expectHome, home)
			}
		})
	}
}

func TestMockFilesystem_Getenv(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*MockFilesystem)
		key    string
		expect string
	}{
		{
			name: "get existing variable",
			setup: func(fs *MockFilesystem) {
				fs.EnvVars["TEST_VAR"] = "test_value"
			},
			key:    "TEST_VAR",
			expect: "test_value",
		},
		{
			name:   "get non-existent variable",
			setup:  func(fs *MockFilesystem) {},
			key:    "NONEXISTENT",
			expect: "",
		},
		{
			name: "get empty variable",
			setup: func(fs *MockFilesystem) {
				fs.EnvVars["EMPTY"] = ""
			},
			key:    "EMPTY",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setup(fs)

			value := fs.Getenv(tt.key)
			if value != tt.expect {
				t.Errorf("Expected value '%s', got '%s'", tt.expect, value)
			}
		})
	}
}

// TestMockFileInfo tests the mockFileInfo implementation
func TestMockFileInfo_Name(t *testing.T) {
	fi := &mockFileInfo{name: "test.txt"}
	if fi.Name() != "test.txt" {
		t.Errorf("Expected name 'test.txt', got '%s'", fi.Name())
	}
}

func TestMockFileInfo_Size(t *testing.T) {
	tests := []struct {
		name   string
		size   int64
		expect int64
	}{
		{"zero size", 0, 0},
		{"positive size", 123, 123},
		{"large size", 1048576, 1048576},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := &mockFileInfo{size: tt.size}
			if fi.Size() != tt.expect {
				t.Errorf("Expected size %d, got %d", tt.expect, fi.Size())
			}
		})
	}
}

func TestMockFileInfo_Mode(t *testing.T) {
	tests := []struct {
		name   string
		mode   fs.FileMode
		expect fs.FileMode
	}{
		{"regular file 0644", 0644, 0644},
		{"regular file 0600", 0600, 0600},
		{"executable 0755", 0755, 0755},
		{"directory", fs.ModeDir | 0755, fs.ModeDir | 0755},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := &mockFileInfo{mode: tt.mode}
			if fi.Mode() != tt.expect {
				t.Errorf("Expected mode %o, got %o", tt.expect, fi.Mode())
			}
		})
	}
}

func TestMockFileInfo_IsDir(t *testing.T) {
	tests := []struct {
		name   string
		mode   fs.FileMode
		expect bool
	}{
		{"regular file", 0644, false},
		{"directory", fs.ModeDir | 0755, true},
		{"executable file", 0755, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi := &mockFileInfo{mode: tt.mode}
			if fi.IsDir() != tt.expect {
				t.Errorf("Expected IsDir() %v, got %v", tt.expect, fi.IsDir())
			}
		})
	}
}

func TestMockFileInfo_ModTime(t *testing.T) {
	fi := &mockFileInfo{}
	modTime := fi.ModTime()
	// Should return zero time
	if !modTime.IsZero() {
		t.Errorf("Expected zero time, got %v", modTime)
	}
	if modTime != (time.Time{}) {
		t.Errorf("Expected time.Time{}, got %v", modTime)
	}
}

func TestMockFileInfo_Sys(t *testing.T) {
	fi := &mockFileInfo{}
	sys := fi.Sys()
	if sys != nil {
		t.Errorf("Expected nil, got %v", sys)
	}
}

// Integration tests combining multiple operations
func TestMockFilesystem_Integration_ReadWriteSequence(t *testing.T) {
	fs := NewMockFilesystem()
	path := "/test/file.txt"
	content1 := []byte("first content")
	content2 := []byte("second content")

	// Write first content
	err := fs.WriteFile(path, content1, 0644)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Read first content
	data, err := fs.ReadFile(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != string(content1) {
		t.Errorf("Expected '%s', got '%s'", string(content1), string(data))
	}

	// Overwrite with second content
	err = fs.WriteFile(path, content2, 0600)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// Read second content
	data, err = fs.ReadFile(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != string(content2) {
		t.Errorf("Expected '%s', got '%s'", string(content2), string(data))
	}

	// Verify mode was updated
	info, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Mode() != 0600 {
		t.Errorf("Expected mode 0600, got %o", info.Mode())
	}
}

func TestMockFilesystem_Integration_RenameAndRead(t *testing.T) {
	fs := NewMockFilesystem()
	srcPath := "/src.txt"
	dstPath := "/dst.txt"
	content := []byte("rename test content")

	// Write initial file
	err := fs.WriteFile(srcPath, content, 0644)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Rename
	err = fs.Rename(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Verify source doesn't exist
	if fs.Exists(srcPath) {
		t.Error("Source should not exist after rename")
	}

	// Verify destination exists and has correct content
	if !fs.Exists(dstPath) {
		t.Error("Destination should exist after rename")
	}

	data, err := fs.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected '%s', got '%s'", string(content), string(data))
	}
}

func TestMockFilesystem_Integration_MultipleFiles(t *testing.T) {
	fs := NewMockFilesystem()

	files := map[string]string{
		"/file1.txt":     "content 1",
		"/file2.txt":     "content 2",
		"/dir/file3.txt": "content 3",
	}

	// Write all files
	for path, content := range files {
		err := fs.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Write failed for %s: %v", path, err)
		}
	}

	// Verify all files exist and have correct content
	for path, expectedContent := range files {
		if !fs.Exists(path) {
			t.Errorf("File %s should exist", path)
		}

		data, err := fs.ReadFile(path)
		if err != nil {
			t.Errorf("Read failed for %s: %v", path, err)
		}
		if string(data) != expectedContent {
			t.Errorf("For %s: expected '%s', got '%s'", path, expectedContent, string(data))
		}
	}
}
