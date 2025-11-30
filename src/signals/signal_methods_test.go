package signals

import (
	"testing"
)

// This file tests the Name(), Emoji(), Diagnostic(), and Remediation() methods
// for signals that don't have dedicated test files covering these methods.

func TestAllSignals_NameNotEmpty(t *testing.T) {
	signals := GetAllSignals()
	for _, sig := range signals {
		name := sig.Name()
		if name == "" {
			t.Errorf("Signal %T has empty Name()", sig)
		}
	}
}

func TestAllSignals_EmojiNotEmpty(t *testing.T) {
	signals := GetAllSignals()
	for _, sig := range signals {
		emoji := sig.Emoji()
		if emoji == "" {
			t.Errorf("Signal %T has empty Emoji()", sig)
		}
	}
}

// Note: Diagnostic() and Remediation() may be empty until Check() is called
// for some signals, so we don't test them in the generic test.

// Individual signal method tests for signals without dedicated test files

func TestAWSAliasHijackSignal_Methods(t *testing.T) {
	signal := NewAWSAliasHijackSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestCargoPathDepsSignal_Methods(t *testing.T) {
	signal := NewCargoPathDepsSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestDanglingSymlinksSignal_Methods(t *testing.T) {
	signal := NewDanglingSymlinksSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestDiskSpaceSignal_Methods(t *testing.T) {
	signal := NewDiskSpaceSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestGoReplaceSignal_Methods(t *testing.T) {
	signal := NewGoReplaceSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestMissingInitPySignal_Methods(t *testing.T) {
	signal := NewMissingInitPySignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestProdPanicSignal_Methods(t *testing.T) {
	signal := NewProdPanicSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestPyCachePollutionSignal_Methods(t *testing.T) {
	signal := NewPyCachePollutionSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestRootOwnedHomeSignal_Methods(t *testing.T) {
	signal := NewRootOwnedHomeSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestSnapshotDependencySignal_Methods(t *testing.T) {
	signal := NewSnapshotDependencySignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestSSHAgentBloatSignal_Methods(t *testing.T) {
	signal := NewSSHAgentBloatSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestSSHKeysSignal_Methods(t *testing.T) {
	signal := NewSSHKeysSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
}

func TestTimeDriftSignal_Methods(t *testing.T) {
	signal := NewTimeDriftSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Diagnostic() == "" {
		t.Error("Diagnostic() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestUntrackedCryptoKeysSignal_Methods(t *testing.T) {
	signal := NewUntrackedCryptoKeysSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}

func TestWorldWritableConfigSignal_Methods(t *testing.T) {
	signal := NewWorldWritableConfigSignal()
	if signal.Name() == "" {
		t.Error("Name() should not be empty")
	}
	if signal.Emoji() == "" {
		t.Error("Emoji() should not be empty")
	}
	if signal.Remediation() == "" {
		t.Error("Remediation() should not be empty")
	}
}
