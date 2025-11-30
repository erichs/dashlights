package signals

import (
	"context"
	"os"
	"testing"
)

func TestHistoryDisabledSignal_Name(t *testing.T) {
	signal := NewHistoryDisabledSignal()
	if signal.Name() != "Blind Spot" {
		t.Errorf("Expected 'Blind Spot', got '%s'", signal.Name())
	}
}

func TestHistoryDisabledSignal_Emoji(t *testing.T) {
	signal := NewHistoryDisabledSignal()
	if signal.Emoji() != "🕶️" {
		t.Errorf("Expected '🕶️', got '%s'", signal.Emoji())
	}
}

func TestHistoryDisabledSignal_Diagnostic(t *testing.T) {
	signal := NewHistoryDisabledSignal()
	signal.reason = "Test reason"
	if signal.Diagnostic() != "Test reason" {
		t.Errorf("Expected 'Test reason', got '%s'", signal.Diagnostic())
	}
}

func TestHistoryDisabledSignal_Remediation(t *testing.T) {
	signal := NewHistoryDisabledSignal()
	expected := "Re-enable shell history for audit trail and incident response"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestHistoryDisabledSignal_Check_Normal(t *testing.T) {
	// Save and restore env vars
	oldHistfile := os.Getenv("HISTFILE")
	oldHistcontrol := os.Getenv("HISTCONTROL")
	defer func() {
		if oldHistfile != "" {
			os.Setenv("HISTFILE", oldHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
		if oldHistcontrol != "" {
			os.Setenv("HISTCONTROL", oldHistcontrol)
		} else {
			os.Unsetenv("HISTCONTROL")
		}
	}()

	// Clear env vars
	os.Unsetenv("HISTFILE")
	os.Unsetenv("HISTCONTROL")

	signal := NewHistoryDisabledSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when history is not disabled")
	}
}

func TestHistoryDisabledSignal_Check_HistfileDevNull(t *testing.T) {
	// Save and restore env var
	oldHistfile := os.Getenv("HISTFILE")
	defer func() {
		if oldHistfile != "" {
			os.Setenv("HISTFILE", oldHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
	}()

	// Set HISTFILE to /dev/null
	os.Setenv("HISTFILE", "/dev/null")

	signal := NewHistoryDisabledSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when HISTFILE is /dev/null")
	}

	expected := "HISTFILE set to /dev/null (history disabled)"
	if signal.reason != expected {
		t.Errorf("Expected reason '%s', got '%s'", expected, signal.reason)
	}
}

func TestHistoryDisabledSignal_Check_HistcontrolIgnorespace(t *testing.T) {
	// Save and restore env vars
	oldHistfile := os.Getenv("HISTFILE")
	oldHistcontrol := os.Getenv("HISTCONTROL")
	defer func() {
		if oldHistfile != "" {
			os.Setenv("HISTFILE", oldHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
		if oldHistcontrol != "" {
			os.Setenv("HISTCONTROL", oldHistcontrol)
		} else {
			os.Unsetenv("HISTCONTROL")
		}
	}()

	// Clear HISTFILE and set HISTCONTROL
	os.Unsetenv("HISTFILE")
	os.Setenv("HISTCONTROL", "ignorespace")

	signal := NewHistoryDisabledSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when HISTCONTROL is ignorespace")
	}

	expected := "HISTCONTROL set to 'ignorespace' (commands with leading space ignored)"
	if signal.reason != expected {
		t.Errorf("Expected reason '%s', got '%s'", expected, signal.reason)
	}
}

func TestHistoryDisabledSignal_Check_HistcontrolIgnoreboth(t *testing.T) {
	// Save and restore env vars
	oldHistfile := os.Getenv("HISTFILE")
	oldHistcontrol := os.Getenv("HISTCONTROL")
	defer func() {
		if oldHistfile != "" {
			os.Setenv("HISTFILE", oldHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
		if oldHistcontrol != "" {
			os.Setenv("HISTCONTROL", oldHistcontrol)
		} else {
			os.Unsetenv("HISTCONTROL")
		}
	}()

	// Clear HISTFILE and set HISTCONTROL
	os.Unsetenv("HISTFILE")
	os.Setenv("HISTCONTROL", "ignoreboth")

	signal := NewHistoryDisabledSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when HISTCONTROL is ignoreboth")
	}

	expected := "HISTCONTROL set to 'ignoreboth' (commands with leading space ignored)"
	if signal.reason != expected {
		t.Errorf("Expected reason '%s', got '%s'", expected, signal.reason)
	}
}

func TestHistoryDisabledSignal_Check_HistcontrolOther(t *testing.T) {
	// Save and restore env vars
	oldHistfile := os.Getenv("HISTFILE")
	oldHistcontrol := os.Getenv("HISTCONTROL")
	defer func() {
		if oldHistfile != "" {
			os.Setenv("HISTFILE", oldHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
		if oldHistcontrol != "" {
			os.Setenv("HISTCONTROL", oldHistcontrol)
		} else {
			os.Unsetenv("HISTCONTROL")
		}
	}()

	// Set HISTCONTROL to something else
	os.Unsetenv("HISTFILE")
	os.Setenv("HISTCONTROL", "ignoredups")

	signal := NewHistoryDisabledSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when HISTCONTROL is ignoredups (not a security issue)")
	}
}
