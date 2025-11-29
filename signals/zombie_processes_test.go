package signals

import (
	"context"
	"runtime"
	"testing"
)

func TestZombieProcessesSignal_Name(t *testing.T) {
	signal := NewZombieProcessesSignal()
	if signal.Name() != "Zombie Apocalypse" {
		t.Errorf("Expected 'Zombie Apocalypse', got '%s'", signal.Name())
	}
}

func TestZombieProcessesSignal_Emoji(t *testing.T) {
	signal := NewZombieProcessesSignal()
	if signal.Emoji() != "🧟" {
		t.Errorf("Expected '🧟', got '%s'", signal.Emoji())
	}
}

func TestZombieProcessesSignal_Diagnostic(t *testing.T) {
	signal := NewZombieProcessesSignal()
	signal.count = 10
	expected := "Excessive zombie processes detected: 10"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestZombieProcessesSignal_Remediation(t *testing.T) {
	signal := NewZombieProcessesSignal()
	expected := "Investigate and fix parent processes not reaping children"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestZombieProcessesSignal_Check_NonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping non-Linux test on Linux")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// On non-Linux systems, should always return false
	if signal.Check(ctx) {
		t.Error("Expected false on non-Linux system")
	}
}

func TestZombieProcessesSignal_Check_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// Run the check - should not panic
	result := signal.Check(ctx)

	// Result depends on actual system state, but count should be set
	// We can't assert the result, but we can verify the count was set
	if result && signal.count <= 5 {
		t.Errorf("Signal returned true but count is %d (threshold is >5)", signal.count)
	}
	if !result && signal.count > 5 {
		t.Errorf("Signal returned false but count is %d (threshold is >5)", signal.count)
	}
}

func TestZombieProcessesSignal_Check_Threshold(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux system")
	}

	signal := NewZombieProcessesSignal()
	ctx := context.Background()

	// Run check
	signal.Check(ctx)

	// Verify threshold logic: >5 zombies triggers signal
	// The actual count depends on system state, but we can verify consistency
	if signal.count > 5 {
		// If we have more than 5 zombies, Check should have returned true
		// Run again to verify
		result := signal.Check(ctx)
		if !result && signal.count > 5 {
			t.Errorf("Expected true when zombie count is %d (>5)", signal.count)
		}
	}
}
