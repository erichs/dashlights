package signals

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPrivilegedPathSignal_Name(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	if signal.Name() != "Privileged Path" {
		t.Errorf("Expected 'Privileged Path', got '%s'", signal.Name())
	}
}

func TestPrivilegedPathSignal_Emoji(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	if signal.Emoji() != "💣" {
		t.Errorf("Expected '💣', got '%s'", signal.Emoji())
	}
}

func TestPrivilegedPathSignal_Diagnostic_NoFindings(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	expected := "Potentially dangerous entries detected in PATH"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Diagnostic_MultipleFindings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	// Create a temporary world-writable directory
	tmpDir, err := os.MkdirTemp("", "worldwritable")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Chmod(tmpDir, 0o777); err != nil {
		t.Skipf("unable to set world-writable permissions on temp dir: %v", err)
	}

	// Set PATH to include both '.' and the world-writable directory
	testPath := strings.Join([]string{".", tmpDir, origPath}, string(os.PathListSeparator))
	if err := os.Setenv("PATH", testPath); err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Fatal("Expected Check to detect multiple PATH issues")
	}

	got := signal.Diagnostic()
	if !strings.HasPrefix(got, "Multiple PATH issues detected:") {
		t.Errorf("Expected diagnostic to start with 'Multiple PATH issues detected:', got %q", got)
	}
	if !strings.Contains(got, "Current directory '.'") {
		t.Errorf("Expected diagnostic to mention current directory '.', got %q", got)
	}
	if !strings.Contains(got, "World-writable PATH entry") || !strings.Contains(got, tmpDir) {
		t.Errorf("Expected diagnostic to mention world-writable PATH entry and directory, got %q", got)
	}
}

func TestPrivilegedPathSignal_Remediation(t *testing.T) {
	signal := NewPrivilegedPathSignal()
	expected := "Remove '.' and world-writable or user bin directories from PATH, or move user bin directories after system paths like /usr/bin"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestPrivilegedPathSignal_Check_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	// On Windows, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on Windows")
	}
}

func TestPrivilegedPathSignal_Check_EmptyPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set empty PATH
	os.Setenv("PATH", "")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when PATH is empty")
	}
}

func TestPrivilegedPathSignal_Check_DotInPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set PATH with '.'
	os.Setenv("PATH", "/usr/bin:.:/usr/local/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when '.' is in PATH")
	}

	expected := "Current directory '.' found in PATH"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Check_EmptyEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set PATH with empty entry (::)
	os.Setenv("PATH", "/usr/bin::/usr/local/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when PATH has empty entry")
	}

	expected := "Empty PATH entry (::) found (implies current directory)"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}

func TestPrivilegedPathSignal_Check_SafePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	// Set safe PATH without '.'
	os.Setenv("PATH", "/usr/bin:/usr/local/bin:/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when PATH is safe")
	}
}

func TestPrivilegedPathSignal_Check_WorldWritablePathEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o777); err != nil {
		t.Skipf("unable to set world-writable permissions on temp dir: %v", err)
	}

	os.Setenv("PATH", dir)

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when PATH has world-writable directory")
	}

	diag := signal.Diagnostic()
	if !strings.Contains(diag, "World-writable PATH entry") || !strings.Contains(diag, dir) {
		t.Errorf("Expected diagnostic to mention world-writable PATH entry and directory, got %q", diag)
	}
}

func TestPrivilegedPathSignal_Check_WorldWritableUserBinBeforeSystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH and HOME
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	homeDir := t.TempDir()
	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	userBin := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(userBin, 0o777); err != nil {
		t.Fatalf("Failed to create user bin directory: %v", err)
	}

	// Ensure the directory is world-writable
	if err := os.Chmod(userBin, 0o777); err != nil {
		t.Skipf("unable to set world-writable permissions on user bin dir: %v", err)
	}

	os.Setenv("PATH", userBin+string(os.PathListSeparator)+"/usr/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Fatal("Expected true when world-writable user bin directory appears before system directories")
	}

	diag := signal.Diagnostic()
	if !strings.Contains(diag, "World-writable user PATH directory $HOME/bin appears before system directories") {
		t.Errorf("Expected diagnostic to mention world-writable user PATH directory before system directories, got %q", diag)
	}
	if strings.Contains(diag, "User PATH directory $HOME/bin appears before system directories") {
		t.Errorf("Did not expect separate user-bin-before-system message when already reported as world-writable, got %q", diag)
	}
}

func TestPrivilegedPathSignal_Check_UserBinBeforeSystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Save and restore PATH and HOME
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	homeDir := t.TempDir()
	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	userBin := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatalf("Failed to create user bin directory: %v", err)
	}

	os.Setenv("PATH", userBin+string(os.PathListSeparator)+"/usr/bin")

	signal := NewPrivilegedPathSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when user bin directory appears before system directories")
	}

	expected := "User PATH directory $HOME/bin appears before system directories"
	if got := signal.Diagnostic(); got != expected {
		t.Errorf("Expected diagnostic %q, got %q", expected, got)
	}
}
func TestBuildUserBinDirMap(t *testing.T) {
	t.Run("DefaultGOPATHWhenUnset", func(t *testing.T) {
		oldHome, hadHome := os.LookupEnv("HOME")
		oldGopath, hadGopath := os.LookupEnv("GOPATH")
		oldCargoHome, hadCargoHome := os.LookupEnv("CARGO_HOME")
		defer func() {
			if hadHome {
				_ = os.Setenv("HOME", oldHome)
			} else {
				_ = os.Unsetenv("HOME")
			}
			if hadGopath {
				_ = os.Setenv("GOPATH", oldGopath)
			} else {
				_ = os.Unsetenv("GOPATH")
			}
			if hadCargoHome {
				_ = os.Setenv("CARGO_HOME", oldCargoHome)
			} else {
				_ = os.Unsetenv("CARGO_HOME")
			}
		}()

		homeDir := t.TempDir()
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Fatalf("Failed to set HOME: %v", err)
		}
		if err := os.Unsetenv("GOPATH"); err != nil {
			t.Fatalf("Failed to unset GOPATH: %v", err)
		}
		if err := os.Unsetenv("CARGO_HOME"); err != nil {
			t.Fatalf("Failed to unset CARGO_HOME: %v", err)
		}

		dirs := buildUserBinDirMap()

		homeBin := filepath.Join(homeDir, "bin")
		if got, ok := dirs[homeBin]; !ok || got != "$HOME/bin" {
			t.Errorf("Expected %q for %q, got %q (present=%v)", "$HOME/bin", homeBin, got, ok)
		}

		homeLocalBin := filepath.Join(homeDir, ".local", "bin")
		if got, ok := dirs[homeLocalBin]; !ok || got != "$HOME/.local/bin" {
			t.Errorf("Expected %q for %q, got %q (present=%v)", "$HOME/.local/bin", homeLocalBin, got, ok)
		}

		homeCargoBin := filepath.Join(homeDir, ".cargo", "bin")
		if got, ok := dirs[homeCargoBin]; !ok || got != "$HOME/.cargo/bin" {
			t.Errorf("Expected %q for %q, got %q (present=%v)", "$HOME/.cargo/bin", homeCargoBin, got, ok)
		}

		defaultGopathBin := filepath.Join(homeDir, "go", "bin")
		if got, ok := dirs[defaultGopathBin]; !ok || got != "$GOPATH/bin" {
			t.Errorf("Expected default GOPATH bin %q for %q, got %q (present=%v)", "$GOPATH/bin", defaultGopathBin, got, ok)
		}
	})

	t.Run("MultipleGOPATHEntiresAndEmpty", func(t *testing.T) {
		oldHome, hadHome := os.LookupEnv("HOME")
		oldGopath, hadGopath := os.LookupEnv("GOPATH")
		oldCargoHome, hadCargoHome := os.LookupEnv("CARGO_HOME")
		defer func() {
			if hadHome {
				_ = os.Setenv("HOME", oldHome)
			} else {
				_ = os.Unsetenv("HOME")
			}
			if hadGopath {
				_ = os.Setenv("GOPATH", oldGopath)
			} else {
				_ = os.Unsetenv("GOPATH")
			}
			if hadCargoHome {
				_ = os.Setenv("CARGO_HOME", oldCargoHome)
			} else {
				_ = os.Unsetenv("CARGO_HOME")
			}
		}()

		homeDir := t.TempDir()
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Fatalf("Failed to set HOME: %v", err)
		}

		g1 := filepath.Join(homeDir, "go1")
		g2 := filepath.Join(homeDir, "go2")
		gopathEnv := strings.Join([]string{g1, "", g2}, string(os.PathListSeparator))
		if err := os.Setenv("GOPATH", gopathEnv); err != nil {
			t.Fatalf("Failed to set GOPATH: %v", err)
		}
		if err := os.Unsetenv("CARGO_HOME"); err != nil {
			t.Fatalf("Failed to unset CARGO_HOME: %v", err)
		}

		dirs := buildUserBinDirMap()

		g1Bin := filepath.Join(g1, "bin")
		if got, ok := dirs[g1Bin]; !ok || got != "$GOPATH/bin" {
			t.Errorf("Expected GOPATH bin %q for %q, got %q (present=%v)", "$GOPATH/bin", g1Bin, got, ok)
		}

		g2Bin := filepath.Join(g2, "bin")
		if got, ok := dirs[g2Bin]; !ok || got != "$GOPATH/bin" {
			t.Errorf("Expected GOPATH bin %q for %q, got %q (present=%v)", "$GOPATH/bin", g2Bin, got, ok)
		}

		// Empty GOPATH segment should not produce an entry.
		emptyBin := filepath.Join("", "bin")
		if _, ok := dirs[emptyBin]; ok {
			t.Errorf("Did not expect entry for empty GOPATH segment %q", emptyBin)
		}
	})

	t.Run("CustomCARGOHome", func(t *testing.T) {
		oldHome, hadHome := os.LookupEnv("HOME")
		oldGopath, hadGopath := os.LookupEnv("GOPATH")
		oldCargoHome, hadCargoHome := os.LookupEnv("CARGO_HOME")
		defer func() {
			if hadHome {
				_ = os.Setenv("HOME", oldHome)
			} else {
				_ = os.Unsetenv("HOME")
			}
			if hadGopath {
				_ = os.Setenv("GOPATH", oldGopath)
			} else {
				_ = os.Unsetenv("GOPATH")
			}
			if hadCargoHome {
				_ = os.Setenv("CARGO_HOME", oldCargoHome)
			} else {
				_ = os.Unsetenv("CARGO_HOME")
			}
		}()

		homeDir := t.TempDir()
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Fatalf("Failed to set HOME: %v", err)
		}
		if err := os.Unsetenv("GOPATH"); err != nil {
			t.Fatalf("Failed to unset GOPATH: %v", err)
		}

		cargoHome := filepath.Join(homeDir, "cargo-home")
		if err := os.Setenv("CARGO_HOME", cargoHome); err != nil {
			t.Fatalf("Failed to set CARGO_HOME: %v", err)
		}

		dirs := buildUserBinDirMap()

		homeCargoBin := filepath.Join(homeDir, ".cargo", "bin")
		if got, ok := dirs[homeCargoBin]; !ok || got != "$HOME/.cargo/bin" {
			t.Errorf("Expected home cargo bin %q for %q, got %q (present=%v)", "$HOME/.cargo/bin", homeCargoBin, got, ok)
		}

		cargoHomeBin := filepath.Join(cargoHome, "bin")
		if got, ok := dirs[cargoHomeBin]; !ok || got != "$CARGO_HOME/bin" {
			t.Errorf("Expected CARGO_HOME bin %q for %q, got %q (present=%v)", "$CARGO_HOME/bin", cargoHomeBin, got, ok)
		}
	})
}

func TestPrivilegedPathSignal_isSystemPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/bin", true},
		{"/sbin", true},
		{"/usr/bin", true},
		{"/usr/sbin", true},
		{"/usr/local/bin", true},
		{"/usr/local/sbin", true},
		{"/home/user/bin", false},
		{"/opt/bin", false},
		{".", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isSystemPath(tt.path)
		if result != tt.expected {
			t.Errorf("isSystemPath(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}
