package signals

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	ps "github.com/mitchellh/go-ps"
)

// mockProcess implements ps.Process for testing.
type mockProcess struct {
	executable string
	pid        int
	ppid       int
}

func (m *mockProcess) Pid() int           { return m.pid }
func (m *mockProcess) PPid() int          { return m.ppid }
func (m *mockProcess) Executable() string { return m.executable }

func TestSecureKeyboardSignal_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping non-darwin test on darwin")
	}

	signal := NewSecureKeyboardSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false on non-darwin platform")
	}
}

func TestSecureKeyboardSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_SECURE_KEYBOARD", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_SECURE_KEYBOARD")

	signal := NewSecureKeyboardSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}

func TestSecureKeyboardSignal_NoTerminalsRunning(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "some_other_process", pid: 1},
			&mockProcess{executable: "bash", pid: 2},
		}, nil
	}

	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when no terminals are running")
	}
}

func TestSecureKeyboardSignal_TerminalRunning_SKEEnabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "Terminal", pid: 100},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		// SKE is enabled
		return true, nil
	}

	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when Terminal is running with SKE enabled")
	}
}

func TestSecureKeyboardSignal_TerminalRunning_SKEDisabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "Terminal", pid: 100},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		// SKE is disabled
		return false, nil
	}

	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when Terminal is running with SKE disabled")
	}

	if len(signal.insecureApps) != 1 || signal.insecureApps[0] != "Terminal.app" {
		t.Errorf("Expected insecureApps=['Terminal.app'], got %v", signal.insecureApps)
	}

	// Verify metadata
	if signal.Name() != "Insecure Terminal" {
		t.Errorf("Expected Name()='Insecure Terminal', got '%s'", signal.Name())
	}
	if signal.Emoji() != "⌨️" {
		t.Errorf("Unexpected emoji: '%s'", signal.Emoji())
	}
	if signal.Diagnostic() == "" {
		t.Error("Expected non-empty diagnostic")
	}
	if signal.Remediation() == "" {
		t.Error("Expected non-empty remediation")
	}
}

func TestSecureKeyboardSignal_iTerm2Running_SKEEnabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "iTerm2", pid: 200},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		return true, nil
	}

	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when iTerm2 is running with SKE enabled")
	}
}

func TestSecureKeyboardSignal_iTerm2Running_SKEDisabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "iTerm2", pid: 200},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		return false, nil
	}

	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when iTerm2 is running with SKE disabled")
	}

	if len(signal.insecureApps) != 1 || signal.insecureApps[0] != "iTerm2" {
		t.Errorf("Expected insecureApps=['iTerm2'], got %v", signal.insecureApps)
	}

	// Verify iTerm2-specific remediation
	remediation := signal.Remediation()
	if remediation != "Enable via iTerm2 menu > Secure Keyboard Entry" {
		t.Errorf("Expected iTerm2-specific remediation, got '%s'", remediation)
	}
}

func TestSecureKeyboardSignal_GhosttyRunning_SKEEnabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "ghostty", pid: 300},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		if plistFile != "com.mitchellh.ghostty.plist" {
			t.Fatalf("Unexpected plistFile: %s", plistFile)
		}
		if keyName != "SecureInput" {
			t.Fatalf("Unexpected keyName: %s", keyName)
		}
		return true, nil
	}

	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when Ghostty is running with SKE enabled")
	}
}

func TestSecureKeyboardSignal_GhosttyRunning_SKEDisabled(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "ghostty", pid: 300},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		if plistFile != "com.mitchellh.ghostty.plist" {
			t.Fatalf("Unexpected plistFile: %s", plistFile)
		}
		if keyName != "SecureInput" {
			t.Fatalf("Unexpected keyName: %s", keyName)
		}
		return false, nil
	}

	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when Ghostty is running with SKE disabled")
	}

	if len(signal.insecureApps) != 1 || signal.insecureApps[0] != "Ghostty" {
		t.Errorf("Expected insecureApps=['Ghostty'], got %v", signal.insecureApps)
	}

	remediation := signal.Remediation()
	if remediation != "Enable via Ghostty menu > Secure Keyboard Entry" {
		t.Errorf("Expected Ghostty-specific remediation, got '%s'", remediation)
	}
}

func TestSecureKeyboardSignal_BothRunning_TerminalInsecure(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "Terminal", pid: 100},
			&mockProcess{executable: "iTerm2", pid: 200},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		// Terminal has SKE disabled, iTerm2 has SKE enabled
		if plistFile == "com.apple.Terminal.plist" {
			return false, nil
		}
		return true, nil
	}

	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when Terminal is insecure")
	}

	// Only Terminal should be flagged
	if len(signal.insecureApps) != 1 || signal.insecureApps[0] != "Terminal.app" {
		t.Errorf("Expected insecureApps=['Terminal.app'], got %v", signal.insecureApps)
	}
}

func TestSecureKeyboardSignal_BothRunning_BothInsecure(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "Terminal", pid: 100},
			&mockProcess{executable: "iTerm2", pid: 200},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		// Both have SKE disabled
		return false, nil
	}

	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when both terminals are insecure")
	}

	// Both should be flagged
	if len(signal.insecureApps) != 2 {
		t.Errorf("Expected 2 insecure apps, got %d: %v", len(signal.insecureApps), signal.insecureApps)
	}

	// Verify diagnostic mentions both
	diag := signal.Diagnostic()
	if !strings.Contains(diag, "Terminal.app") || !strings.Contains(diag, "iTerm2") {
		t.Errorf("Expected diagnostic to mention both apps, got: %s", diag)
	}

	// Verify remediation mentions both
	rem := signal.Remediation()
	if !strings.Contains(rem, "Terminal") || !strings.Contains(rem, "iTerm2") {
		t.Errorf("Expected remediation to mention both apps, got: %s", rem)
	}
}

func TestSecureKeyboardSignal_ProcessListError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return nil, errors.New("process list error")
	}

	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when process listing fails")
	}
}

func TestSecureKeyboardSignal_PlistReadError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		return []ps.Process{
			&mockProcess{executable: "Terminal", pid: 100},
		}, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		return false, errors.New("plist read error")
	}

	ctx := context.Background()

	// Should not signal when plist can't be read (assume safe)
	if signal.Check(ctx) {
		t.Error("Expected false when plist read fails")
	}
}

func TestSecureKeyboardSignal_ContextCancellation_Processes(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()

	callCount := 0
	signal.processLister = func() ([]ps.Process, error) {
		// Return many processes to trigger iteration
		processes := make([]ps.Process, 100)
		for i := range processes {
			processes[i] = &mockProcess{executable: "proc", pid: i}
		}
		return processes, nil
	}
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		callCount++
		return false, nil
	}

	// Create a pre-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should exit immediately
	result := signal.Check(ctx)

	if result {
		t.Error("Expected false when context is cancelled")
	}

	// plistReader should never be called because we exit during process iteration
	if callCount > 0 {
		t.Errorf("Expected plistReader not to be called, was called %d times", callCount)
	}
}

func TestSecureKeyboardSignal_ContextCancellation_Apps(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	signal := NewSecureKeyboardSignal()
	signal.processLister = func() ([]ps.Process, error) {
		// Return just 1 process so we get past process loop
		return []ps.Process{}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := signal.Check(ctx)

	if result {
		t.Error("Expected false when context is cancelled")
	}
}

func TestSecureKeyboardSignal_Metadata(t *testing.T) {
	signal := NewSecureKeyboardSignal()

	if signal.Name() != "Insecure Terminal" {
		t.Errorf("Unexpected Name: %s", signal.Name())
	}

	if signal.Emoji() != "⌨️" {
		t.Errorf("Unexpected Emoji: %s", signal.Emoji())
	}

	// Set insecureApps for Terminal only
	signal.insecureApps = []string{"Terminal.app"}
	diag := signal.Diagnostic()
	if diag == "" {
		t.Error("Expected non-empty diagnostic")
	}
	if signal.Remediation() != "Enable via Terminal menu > Secure Keyboard Entry" {
		t.Errorf("Unexpected Terminal remediation: %s", signal.Remediation())
	}

	// Set insecureApps for iTerm2 only
	signal.insecureApps = []string{"iTerm2"}
	if signal.Remediation() != "Enable via iTerm2 menu > Secure Keyboard Entry" {
		t.Errorf("Unexpected iTerm2 remediation: %s", signal.Remediation())
	}

	// Set insecureApps for Ghostty only
	signal.insecureApps = []string{"Ghostty"}
	if signal.Remediation() != "Enable via Ghostty menu > Secure Keyboard Entry" {
		t.Errorf("Unexpected Ghostty remediation: %s", signal.Remediation())
	}

	// Set insecureApps for both
	signal.insecureApps = []string{"Terminal.app", "iTerm2"}
	diag = signal.Diagnostic()
	if !strings.Contains(diag, "Terminal.app and iTerm2") {
		t.Errorf("Expected diagnostic to mention both apps, got: %s", diag)
	}
	rem := signal.Remediation()
	if !strings.Contains(rem, "Terminal") || !strings.Contains(rem, "iTerm2") {
		t.Errorf("Expected remediation to mention both, got: %s", rem)
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_MissingKey(t *testing.T) {
	// Create a temp plist with no SecureKeyboardEntry key
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write a minimal plist without the key
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>SomeOtherKey</key>
	<string>value</string>
</dict>
</plist>`
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	// Override plistReader to use our temp path
	signal.plistReader = func(ctx context.Context, plistFile, keyName string) (bool, error) {
		return signal.readPlistKeyFromPath(ctx, plistPath, keyName)
	}

	enabled, err := signal.plistReader(context.Background(), "", "SecureKeyboardEntry")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Missing key = disabled (default is OFF)
	if enabled {
		t.Error("Expected false when key is missing")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_BoolTrue(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write plist with SecureKeyboardEntry = true
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>SecureKeyboardEntry</key>
	<true/>
</dict>
</plist>`
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	enabled, err := signal.readPlistKeyFromPath(context.Background(), plistPath, "SecureKeyboardEntry")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected true when key is set to true")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_BoolFalse(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write plist with SecureKeyboardEntry = false
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>SecureKeyboardEntry</key>
	<false/>
</dict>
</plist>`
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	enabled, err := signal.readPlistKeyFromPath(context.Background(), plistPath, "SecureKeyboardEntry")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if enabled {
		t.Error("Expected false when key is set to false")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_Integer1(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write plist with SecureKeyboardEntry = 1 (integer)
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>SecureKeyboardEntry</key>
	<integer>1</integer>
</dict>
</plist>`
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	enabled, err := signal.readPlistKeyFromPath(context.Background(), plistPath, "SecureKeyboardEntry")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected true when key is set to integer 1")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_Integer0(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write plist with SecureKeyboardEntry = 0 (integer)
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>SecureKeyboardEntry</key>
	<integer>0</integer>
</dict>
</plist>`
	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	enabled, err := signal.readPlistKeyFromPath(context.Background(), plistPath, "SecureKeyboardEntry")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if enabled {
		t.Error("Expected false when key is set to integer 0")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_MalformedPlist(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/com.apple.Terminal.plist"

	// Write invalid plist
	err := os.WriteFile(plistPath, []byte("not a valid plist"), 0644)
	if err != nil {
		t.Fatalf("Failed to write plist: %v", err)
	}

	signal := NewSecureKeyboardSignal()
	_, err = signal.readPlistKeyFromPath(context.Background(), plistPath, "SecureKeyboardEntry")
	if err == nil {
		t.Error("Expected error for malformed plist")
	}
}

func TestSecureKeyboardSignal_ReadPlistKey_FileNotFound(t *testing.T) {
	signal := NewSecureKeyboardSignal()
	_, err := signal.readPlistKeyFromPath(context.Background(), "/nonexistent/path/file.plist", "SecureKeyboardEntry")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestSecureKeyboardSignal_RealProcessCheck(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping darwin-specific test on non-darwin")
	}

	// Integration test: actually enumerate processes and read plist
	signal := NewSecureKeyboardSignal()
	ctx := context.Background()

	// This should not panic or error
	_ = signal.Check(ctx)

	// Verify metadata is always available
	if signal.Name() == "" {
		t.Error("Expected non-empty name")
	}
	if signal.Emoji() == "" {
		t.Error("Expected non-empty emoji")
	}
}
