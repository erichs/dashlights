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

func TestDumpsterFireSignal_NoSensitiveFiles(t *testing.T) {
	// Create a temp directory with only normal files
	tmpDir := t.TempDir()

	normalFiles := []string{"readme.txt", "main.go", "config.yaml"}
	for _, f := range normalFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Change to temp directory
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()

	// The signal may still detect files in ~/Downloads, ~/Desktop, or /tmp
	// so we can't guarantee false, but totalCount for tmpDir should be 0
	signal.Check(ctx)

	// Check that our temp dir doesn't have matches
	if count := signal.GetCounts()[tmpDir]; count != 0 {
		t.Errorf("Expected 0 sensitive files in temp dir, got %d", count)
	}
}

func TestDumpsterFireSignal_DetectsSQLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create SQL dump file
	if err := os.WriteFile(filepath.Join(tmpDir, "backup.sql"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()

	// Run check
	detected := signal.Check(ctx)

	// Should have found the SQL file
	// Use resolved CWD path since macOS /tmp resolves to /private/tmp
	cwd, _ := os.Getwd()
	if count := signal.GetCounts()[cwd]; count != 1 {
		t.Errorf("Expected 1 SQL file in %s, got %d (detected=%v, counts=%v)", cwd, count, detected, signal.GetCounts())
	}

	// Test Diagnostic after detection (coverage for totalCount > 0 branch)
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

func TestDumpsterFireSignal_DetectsMultipleTypes(t *testing.T) {
	tmpDir := t.TempDir()

	sensitiveFiles := []string{
		"dump.sql",
		"data.sqlite",
		"app.db",
		"network.pcap",
		"request.har",
		"cert.pem",
	}
	for _, f := range sensitiveFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()
	signal.Check(ctx)

	cwd, _ := os.Getwd()
	if count := signal.GetCounts()[cwd]; count != 6 {
		t.Errorf("Expected 6 sensitive files, got %d", count)
	}
}

func TestDumpsterFireSignal_DetectsProdPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	// File with "prod" in name
	if err := os.WriteFile(filepath.Join(tmpDir, "prod-data.csv"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()
	signal.Check(ctx)

	cwd, _ := os.Getwd()
	if count := signal.GetCounts()[cwd]; count != 1 {
		t.Errorf("Expected 1 prod file, got %d", count)
	}
}

func TestDumpsterFireSignal_DetectsDumpPrefix(t *testing.T) {
	tmpDir := t.TempDir()

	// File with dump- prefix
	if err := os.WriteFile(filepath.Join(tmpDir, "dump-2024-01-01.tar.gz"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()
	signal.Check(ctx)

	cwd, _ := os.Getwd()
	if count := signal.GetCounts()[cwd]; count != 1 {
		t.Errorf("Expected 1 dump file, got %d", count)
	}
}

func TestDumpsterFireSignal_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with .sql extension (should be skipped)
	subDir := filepath.Join(tmpDir, "subdir.sql")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()
	signal.Check(ctx)

	cwd, _ := os.Getwd()
	if count := signal.GetCounts()[cwd]; count != 0 {
		t.Errorf("Expected 0 (directories should be skipped), got %d", count)
	}
}

func TestDumpsterFireSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_DUMPSTER_FIRE", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_DUMPSTER_FIRE")

	signal := NewDumpsterFireSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled")
	}
}

func TestDumpsterFireSignal_ImplementsVerboseRemediator(t *testing.T) {
	signal := NewDumpsterFireSignal()

	// Type assertion to verify the interface is implemented
	_, ok := interface{}(signal).(VerboseRemediator)
	if !ok {
		t.Error("DumpsterFireSignal should implement VerboseRemediator interface")
	}
}

func TestDumpsterFireSignal_VerboseRemediation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sensitive file
	if err := os.WriteFile(filepath.Join(tmpDir, "backup.sql"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()
	signal.Check(ctx)

	verbose := signal.VerboseRemediation()

	if verbose == "" {
		t.Error("Expected verbose remediation to be non-empty")
	}
	if !strings.Contains(verbose, "rm") {
		t.Error("Expected verbose remediation to contain 'rm' command")
	}
	if !strings.Contains(verbose, "backup.sql") {
		t.Error("Expected verbose remediation to contain detected filename")
	}
}

func TestDumpsterFireSignal_VerboseRemediationEmpty(t *testing.T) {
	signal := NewDumpsterFireSignal()
	// Don't call Check() - no files found

	verbose := signal.VerboseRemediation()
	if verbose != "" {
		t.Errorf("Expected empty verbose remediation when no files found, got %q", verbose)
	}
}

func TestDumpsterFireSignal_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many sensitive files
	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("dump%d.sql", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()

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

func TestDumpsterFireSignal_PerformanceWithManyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()

	// Create 500 sensitive files (pathological case)
	for i := 0; i < 500; i++ {
		filename := fmt.Sprintf("dump%d.sql", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	signal := NewDumpsterFireSignal()
	ctx := context.Background()

	start := time.Now()
	signal.Check(ctx)
	elapsed := time.Since(start)

	// Should complete well under 100ms due to limits (relaxed for CI variability)
	if elapsed > 50*time.Millisecond {
		t.Errorf("Check took too long: %v (expected < 50ms)", elapsed)
	}

	// Should have found some files but be limited by maxMatchesPerDir
	if signal.TotalCount() == 0 {
		t.Error("Expected to find some sensitive files")
	}
	// Max 10 per directory due to limits
	cwd, _ := os.Getwd()
	if signal.GetCounts()[cwd] > 10 {
		t.Errorf("Expected max 10 matches per dir due to limits, got %d", signal.GetCounts()[cwd])
	}
}
