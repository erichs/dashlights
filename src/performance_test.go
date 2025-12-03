//go:build integration
// +build integration

package main

import (
	"os/exec"
	"testing"
	"time"
)

// TestPerformanceThreshold verifies that the dashlights binary completes
// execution in 10 milliseconds or less. This is a critical performance
// requirement to ensure the tool can be used in shell prompts without
// noticeable delay.
func TestPerformanceThreshold(t *testing.T) {
	const maxDuration = 10 * time.Millisecond
	const numRuns = 10   // Run multiple times to get a more reliable measurement
	const warmupRuns = 3 // Discard first few runs (macOS code signing overhead)

	// Build the binary first
	// Using -buildvcs=false to avoid VCS errors in network-isolated environments
	// Build from parent directory (repo root) where go.mod lives
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", "src/dashlights_test", "./src")
	buildCmd.Dir = ".."
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, buildOutput)
	}
	defer exec.Command("rm", "dashlights_test").Run()

	// Warmup runs to get past macOS Gatekeeper/code signing overhead
	for i := 0; i < warmupRuns; i++ {
		cmd := exec.Command("./dashlights_test")
		cmd.Run()
	}

	var totalDuration time.Duration
	var maxRunDuration time.Duration
	var minRunDuration time.Duration = time.Hour // Start with a large value

	for i := 0; i < numRuns; i++ {
		start := time.Now()
		cmd := exec.Command("./dashlights_test")
		err := cmd.Run()
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Run %d: Binary execution failed: %v", i+1, err)
		}

		totalDuration += elapsed
		if elapsed > maxRunDuration {
			maxRunDuration = elapsed
		}
		if elapsed < minRunDuration {
			minRunDuration = elapsed
		}
	}

	avgDuration := totalDuration / time.Duration(numRuns)

	t.Logf("Performance Results (after %d warmup runs):", warmupRuns)
	t.Logf("  Minimum: %v", minRunDuration)
	t.Logf("  Average: %v", avgDuration)
	t.Logf("  Maximum: %v", maxRunDuration)
	t.Logf("  Threshold: %v", maxDuration)

	// Use minimum time as the best indicator of actual performance
	// (excludes OS scheduling noise)
	if minRunDuration > maxDuration {
		t.Errorf("❌ PERFORMANCE THRESHOLD EXCEEDED!\n"+
			"Minimum execution time: %v (threshold: %v)\n"+
			"Average execution time: %v\n"+
			"The binary is %.1fx slower than required.",
			minRunDuration, maxDuration, avgDuration,
			float64(minRunDuration)/float64(maxDuration))
	} else {
		t.Logf("✅ Performance threshold met! Binary completes in %v minimum (%.1f%% of threshold)",
			minRunDuration, float64(minRunDuration)/float64(maxDuration)*100)
	}
}
