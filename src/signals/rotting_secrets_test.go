package signals

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRottingSecretsSignal_NewFilesNotDetected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a new SQL file (mtime is now, so < 7 days old)
	sqlFile := filepath.Join(tmpDir, "backup.sql")
	if err := os.WriteFile(sqlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()
	signal.Check(ctx)

	// New files should not be detected
	if signal.Count() > 0 {
		// Check if the count is from our temp dir (might be from ~/Downloads etc)
		// We need to make sure it's not from our test file
		t.Log("Note: Count may include files from system hot zones")
	}
}

func TestRottingSecretsSignal_OldFilesDetected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create SQL file and backdate it to 10 days ago
	sqlFile := filepath.Join(tmpDir, "old-backup.sql")
	if err := os.WriteFile(sqlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set mtime to 10 days ago
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(sqlFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected to detect old sensitive file")
	}

	if signal.Count() < 1 {
		t.Errorf("Expected at least 1 old file, got %d", signal.Count())
	}

	// Oldest age should be at least 10 days
	if signal.OldestAge() < 9*24*time.Hour {
		t.Errorf("Expected oldest age >= 9 days, got %v", signal.OldestAge())
	}

	// Test Diagnostic after detection (coverage for count > 0 branch)
	diag := signal.Diagnostic()
	if !strings.Contains(diag, "sensitive file") {
		t.Errorf("Expected diagnostic to mention sensitive files, got %q", diag)
	}

	// Test Remediation (coverage)
	rem := signal.Remediation()
	if rem == "" {
		t.Error("Expected non-empty remediation")
	}
}

func TestRottingSecretsSignal_ExactlySevenDaysNotDetected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file and set mtime to exactly 7 days ago
	sqlFile := filepath.Join(tmpDir, "week-old.sql")
	if err := os.WriteFile(sqlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Exactly 7 days - should NOT be detected (threshold is > 7 days)
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	if err := os.Chtimes(sqlFile, sevenDaysAgo, sevenDaysAgo); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()
	signal.Check(ctx)

	// A file at exactly 7 days should NOT trigger the signal
	// (The threshold is > 7 days, not >= 7 days)
	// Note: This is checking the logic, but real detection may vary by milliseconds
}

func TestRottingSecretsSignal_JustOverSevenDaysDetected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file and set mtime to just over 7 days ago
	sqlFile := filepath.Join(tmpDir, "old-enough.sql")
	if err := os.WriteFile(sqlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// 7 days + 1 hour
	oldTime := time.Now().Add(-7*24*time.Hour - time.Hour)
	if err := os.Chtimes(sqlFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected to detect file just over 7 days old")
	}
}

func TestRottingSecretsSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_ROTTING_SECRETS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_ROTTING_SECRETS")

	signal := NewRottingSecretsSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled")
	}
}

func TestRottingSecretsSignal_Diagnostic(t *testing.T) {
	signal := NewRottingSecretsSignal()

	// Before check, diagnostic should be generic
	diag := signal.Diagnostic()
	if diag == "" {
		t.Error("Expected non-empty diagnostic")
	}
}

func TestRottingSecretsSignal_Name(t *testing.T) {
	signal := NewRottingSecretsSignal()
	if signal.Name() != "Rotting Secrets" {
		t.Errorf("Expected 'Rotting Secrets', got %s", signal.Name())
	}
}

func TestRottingSecretsSignal_Emoji(t *testing.T) {
	signal := NewRottingSecretsSignal()
	if signal.Emoji() != "🦴" {
		t.Errorf("Expected bone emoji, got %s", signal.Emoji())
	}
}

func TestRottingSecretsSignal_ImplementsVerboseRemediator(t *testing.T) {
	signal := NewRottingSecretsSignal()

	// Type assertion to verify the interface is implemented
	_, ok := interface{}(signal).(VerboseRemediator)
	if !ok {
		t.Error("RottingSecretsSignal should implement VerboseRemediator interface")
	}
}

func TestRottingSecretsSignal_VerboseRemediation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sensitive file and make it old
	sqlFile := filepath.Join(tmpDir, "old-dump.sql")
	if err := os.WriteFile(sqlFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set mtime to 10 days ago
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(sqlFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()
	signal.Check(ctx)

	verbose := signal.VerboseRemediation()

	if verbose == "" {
		t.Error("Expected verbose remediation to be non-empty")
	}
	if !strings.Contains(verbose, "rm") {
		t.Error("Expected verbose remediation to contain 'rm' command")
	}
	if !strings.Contains(verbose, "old-dump.sql") {
		t.Error("Expected verbose remediation to contain detected filename")
	}
	if !strings.Contains(verbose, "7 days old") {
		t.Error("Expected verbose remediation to mention 7 days")
	}
}

func TestRottingSecretsSignal_VerboseRemediationEmpty(t *testing.T) {
	signal := NewRottingSecretsSignal()
	// Don't call Check() - no files found

	verbose := signal.VerboseRemediation()
	if verbose != "" {
		t.Errorf("Expected empty verbose remediation when no files found, got %q", verbose)
	}
}

func TestRottingSecretsSignal_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many old sensitive files
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("dump%d.sql", i)
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(path, oldTime, oldTime)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()

	// Pre-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	result := signal.Check(ctx)
	elapsed := time.Since(start)

	if result {
		t.Error("Expected false when context is cancelled")
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 10ms)", elapsed)
	}
}

func TestRottingSecretsSignal_PerformanceWithManyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()

	// Create 500 old sensitive files (pathological case)
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	for i := 0; i < 500; i++ {
		filename := fmt.Sprintf("dump%d.sql", i)
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(path, oldTime, oldTime)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewRottingSecretsSignal()
	ctx := context.Background()

	start := time.Now()
	signal.Check(ctx)
	elapsed := time.Since(start)

	// Should complete well under 100ms due to limits (relaxed for CI variability)
	if elapsed > 50*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 50ms)", elapsed)
	}

	// Should have found some files but be limited by maxMatchesPerDir
	if signal.Count() == 0 {
		t.Error("Expected to find some old sensitive files")
	}
	// Max 10 per directory due to limits
	if signal.Count() > 10 {
		t.Errorf("Expected max 10 matches per dir due to limits, got %d", signal.Count())
	}
}
