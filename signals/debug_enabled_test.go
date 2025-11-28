package signals

import (
	"context"
	"os"
	"testing"
)

func TestDebugEnabledSignal_NoDebugVars(t *testing.T) {
	// Ensure no debug vars are set
	os.Unsetenv("DEBUG")
	os.Unsetenv("TRACE")
	os.Unsetenv("VERBOSE")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no debug variables are set")
	}
}

func TestDebugEnabledSignal_DEBUGSet(t *testing.T) {
	// Set DEBUG variable
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when DEBUG is set")
	}
}

func TestDebugEnabledSignal_TRACESet(t *testing.T) {
	// Set TRACE variable
	os.Setenv("TRACE", "1")
	defer os.Unsetenv("TRACE")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when TRACE is set")
	}
}

func TestDebugEnabledSignal_VERBOSESet(t *testing.T) {
	// Set VERBOSE variable
	os.Setenv("VERBOSE", "1")
	defer os.Unsetenv("VERBOSE")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when VERBOSE is set")
	}
}

func TestDebugEnabledSignal_DEBUGSetToFalse(t *testing.T) {
	// Set DEBUG to "false" - should still trigger because it's set
	os.Setenv("DEBUG", "false")
	defer os.Unsetenv("DEBUG")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when DEBUG is set (even to 'false')")
	}
}

func TestDebugEnabledSignal_DEBUGSetToEmpty(t *testing.T) {
	// Set DEBUG to empty string - should still trigger because it's set
	os.Setenv("DEBUG", "")
	defer os.Unsetenv("DEBUG")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when DEBUG is set (even to empty string)")
	}
}

func TestDebugEnabledSignal_MultipleSet(t *testing.T) {
	// Set multiple debug variables
	os.Setenv("DEBUG", "true")
	os.Setenv("TRACE", "1")
	os.Setenv("VERBOSE", "yes")
	defer func() {
		os.Unsetenv("DEBUG")
		os.Unsetenv("TRACE")
		os.Unsetenv("VERBOSE")
	}()

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when multiple debug variables are set")
	}
}

func TestDebugEnabledSignal_ValueDoesntMatter(t *testing.T) {
	// Test various values - all should trigger
	testCases := []struct {
		name  string
		value string
	}{
		{"true", "true"},
		{"false", "false"},
		{"1", "1"},
		{"0", "0"},
		{"yes", "yes"},
		{"no", "no"},
		{"empty", ""},
		{"random", "random_value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("DEBUG", tc.value)
			defer os.Unsetenv("DEBUG")

			signal := NewDebugEnabledSignal()
			ctx := context.Background()

			result := signal.Check(ctx)
			if !result {
				t.Errorf("Expected true when DEBUG is set to '%s'", tc.value)
			}
		})
	}
}

