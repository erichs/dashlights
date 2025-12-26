package signals

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	ps "github.com/mitchellh/go-ps"
	"howett.net/plist"

	"github.com/erichs/dashlights/src/signals/internal/fileutil"
	"github.com/erichs/dashlights/src/signals/internal/homedirutil"
)

// appConfig defines the process name and plist settings for each terminal app.
type appConfig struct {
	processName string
	plistFile   string // plist filename in ~/Library/Preferences/
	keyName     string
	displayName string
}

// terminalApps lists the terminal applications to check for Secure Keyboard Entry.
var terminalApps = []appConfig{
	{"Terminal", "com.apple.Terminal.plist", "SecureKeyboardEntry", "Terminal.app"},
	{"iTerm2", "com.googlecode.iterm2.plist", "Secure Input", "iTerm2"},
	{"ghostty", "com.mitchellh.ghostty.plist", "SecureInput", "Ghostty"},
}

// SecureKeyboardSignal detects when Terminal.app, iTerm2, or Ghostty is running
// without Secure Keyboard Entry enabled, which could allow keyloggers to
// intercept sensitive input like passwords and tokens.
type SecureKeyboardSignal struct {
	insecureApps []string // Which apps triggered (for Diagnostic)

	// processLister allows injecting a mock for testing.
	processLister func() ([]ps.Process, error)
	// plistReader allows injecting a mock for testing.
	// Returns true if the setting is enabled, false otherwise.
	plistReader func(ctx context.Context, plistFile, keyName string) (bool, error)
}

// NewSecureKeyboardSignal creates a SecureKeyboardSignal.
func NewSecureKeyboardSignal() *SecureKeyboardSignal {
	s := &SecureKeyboardSignal{}
	s.processLister = ps.Processes
	s.plistReader = s.readPlistKey
	return s
}

// Name returns the human-readable name of the signal.
func (s *SecureKeyboardSignal) Name() string {
	return "Insecure Terminal"
}

// Emoji returns the emoji associated with the signal.
func (s *SecureKeyboardSignal) Emoji() string {
	return "⌨️"
}

// Diagnostic returns a description of the detected issue.
func (s *SecureKeyboardSignal) Diagnostic() string {
	if len(s.insecureApps) == 1 {
		return fmt.Sprintf("%s is running without Secure Keyboard Entry - keystrokes may be intercepted", s.insecureApps[0])
	}
	return fmt.Sprintf("%s are running without Secure Keyboard Entry - keystrokes may be intercepted",
		strings.Join(s.insecureApps, " and "))
}

// Remediation returns guidance on how to fix the issue.
func (s *SecureKeyboardSignal) Remediation() string {
	var steps []string
	for _, app := range s.insecureApps {
		if app == "Terminal.app" {
			steps = append(steps, "Terminal menu > Secure Keyboard Entry")
		} else if app == "Ghostty" {
			steps = append(steps, "Ghostty menu > Secure Keyboard Entry")
		} else {
			steps = append(steps, "iTerm2 menu > Secure Keyboard Entry")
		}
	}
	return "Enable via " + strings.Join(steps, "; ")
}

// Check detects if a terminal is running without Secure Keyboard Entry enabled.
func (s *SecureKeyboardSignal) Check(ctx context.Context) bool {
	// macOS only
	if runtime.GOOS != "darwin" {
		return false
	}

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_SECURE_KEYBOARD") != "" {
		return false
	}

	// Get all running processes once (efficient)
	processes, err := s.processLister()
	if err != nil {
		return false
	}

	// Build set of running process names
	running := make(map[string]bool)
	for _, p := range processes {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		running[p.Executable()] = true
	}

	// Reset insecureApps for fresh check
	s.insecureApps = nil

	// Check each terminal app
	for _, app := range terminalApps {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if running[app.processName] {
			enabled, err := s.plistReader(ctx, app.plistFile, app.keyName)
			if err != nil {
				// Can't read plist - assume safe
				continue
			}
			if !enabled {
				s.insecureApps = append(s.insecureApps, app.displayName)
			}
		}
	}

	return len(s.insecureApps) > 0
}

// readPlistKey reads a preference key from the plist file.
// The plist is located at ~/Library/Preferences/{plistFile}.
func (s *SecureKeyboardSignal) readPlistKey(ctx context.Context, plistFile, keyName string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	path, err := homedirutil.SafeHomePath("Library", "Preferences", plistFile)
	if err != nil {
		return false, err
	}

	return s.readPlistKeyFromPath(ctx, path, keyName)
}

// readPlistKeyFromPath reads a preference key from a plist file at the given path.
// This is separated from readPlistKey to allow testing with temp files.
func (s *SecureKeyboardSignal) readPlistKeyFromPath(ctx context.Context, path, keyName string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	data, err := fileutil.ReadFileLimited(path, 512*1024)
	if err != nil {
		return false, err
	}
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	var prefs map[string]interface{}
	if _, err := plist.Unmarshal(data, &prefs); err != nil {
		return false, err
	}

	// Check the key - handles both bool and int (0/1) values
	if val, ok := prefs[keyName]; ok {
		switch v := val.(type) {
		case bool:
			return v, nil
		case int64:
			return v != 0, nil
		case uint64:
			return v != 0, nil
		}
	}

	// Key not set = disabled (default is OFF)
	return false, nil
}
