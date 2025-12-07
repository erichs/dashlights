package signals

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestLDPreloadSignal_Name(t *testing.T) {
	signal := NewLDPreloadSignal()
	if signal.Name() != "Trojan Horse" {
		t.Errorf("Expected 'Trojan Horse', got '%s'", signal.Name())
	}
}

func TestLDPreloadSignal_Emoji(t *testing.T) {
	signal := NewLDPreloadSignal()
	if signal.Emoji() != "🐴" {
		t.Errorf("Expected '🐴', got '%s'", signal.Emoji())
	}
}

func TestLDPreloadSignal_Diagnostic(t *testing.T) {
	signal := NewLDPreloadSignal()
	signal.varName = "LD_PRELOAD"
	signal.value = "/path/to/lib.so"
	expected := "LD_PRELOAD is set to: /path/to/lib.so"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestLDPreloadSignal_Remediation(t *testing.T) {
	signal := NewLDPreloadSignal()
	signal.varName = "LD_PRELOAD"
	expected := "Unset LD_PRELOAD unless intentionally debugging"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestLDPreloadSignal_Check_NoPreload(t *testing.T) {
	// Save and restore env vars
	oldLDPreload := os.Getenv("LD_PRELOAD")
	oldDyldInsert := os.Getenv("DYLD_INSERT_LIBRARIES")
	defer func() {
		if oldLDPreload != "" {
			os.Setenv("LD_PRELOAD", oldLDPreload)
		} else {
			os.Unsetenv("LD_PRELOAD")
		}
		if oldDyldInsert != "" {
			os.Setenv("DYLD_INSERT_LIBRARIES", oldDyldInsert)
		} else {
			os.Unsetenv("DYLD_INSERT_LIBRARIES")
		}
	}()

	// Clear both env vars
	os.Unsetenv("LD_PRELOAD")
	os.Unsetenv("DYLD_INSERT_LIBRARIES")

	signal := NewLDPreloadSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when no preload variables are set")
	}
}

func TestLDPreloadSignal_Check_LDPreloadSet_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	// Save and restore env var
	oldLDPreload := os.Getenv("LD_PRELOAD")
	defer func() {
		if oldLDPreload != "" {
			os.Setenv("LD_PRELOAD", oldLDPreload)
		} else {
			os.Unsetenv("LD_PRELOAD")
		}
	}()

	// Set LD_PRELOAD
	testValue := "/tmp/test.so"
	os.Setenv("LD_PRELOAD", testValue)

	signal := NewLDPreloadSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when LD_PRELOAD is set on Linux")
	}

	if signal.varName != "LD_PRELOAD" {
		t.Errorf("Expected varName 'LD_PRELOAD', got '%s'", signal.varName)
	}

	if signal.value != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, signal.value)
	}
}

func TestLDPreloadSignal_Check_DyldInsertSet_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS system")
	}

	// Save and restore env var
	oldDyldInsert := os.Getenv("DYLD_INSERT_LIBRARIES")
	defer func() {
		if oldDyldInsert != "" {
			os.Setenv("DYLD_INSERT_LIBRARIES", oldDyldInsert)
		} else {
			os.Unsetenv("DYLD_INSERT_LIBRARIES")
		}
	}()

	// Set DYLD_INSERT_LIBRARIES
	testValue := "/tmp/test.dylib"
	os.Setenv("DYLD_INSERT_LIBRARIES", testValue)

	signal := NewLDPreloadSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when DYLD_INSERT_LIBRARIES is set on macOS")
	}

	if signal.varName != "DYLD_INSERT_LIBRARIES" {
		t.Errorf("Expected varName 'DYLD_INSERT_LIBRARIES', got '%s'", signal.varName)
	}

	if signal.value != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, signal.value)
	}
}

func TestLDPreloadSignal_Check_WrongOSVariable(t *testing.T) {
	// Save and restore env vars
	oldLDPreload := os.Getenv("LD_PRELOAD")
	oldDyldInsert := os.Getenv("DYLD_INSERT_LIBRARIES")
	defer func() {
		if oldLDPreload != "" {
			os.Setenv("LD_PRELOAD", oldLDPreload)
		} else {
			os.Unsetenv("LD_PRELOAD")
		}
		if oldDyldInsert != "" {
			os.Setenv("DYLD_INSERT_LIBRARIES", oldDyldInsert)
		} else {
			os.Unsetenv("DYLD_INSERT_LIBRARIES")
		}
	}()

	signal := NewLDPreloadSignal()
	ctx := context.Background()

	switch runtime.GOOS {
	case "linux":
		// Set macOS variable on Linux - should not trigger
		os.Unsetenv("LD_PRELOAD")
		os.Setenv("DYLD_INSERT_LIBRARIES", "/tmp/test.dylib")

		if signal.Check(ctx) {
			t.Error("Expected false when DYLD_INSERT_LIBRARIES is set on Linux")
		}
	case "darwin":
		// Set Linux variable on macOS - should not trigger
		os.Setenv("LD_PRELOAD", "/tmp/test.so")
		os.Unsetenv("DYLD_INSERT_LIBRARIES")

		if signal.Check(ctx) {
			t.Error("Expected false when LD_PRELOAD is set on macOS")
		}
	}
}

func TestLDPreloadSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_LD_PRELOAD", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_LD_PRELOAD")

	signal := NewLDPreloadSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
