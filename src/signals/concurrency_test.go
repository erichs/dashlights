package signals

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestConcurrentCheckAll verifies that multiple concurrent CheckAll executions
// don't interfere with each other (simulating multiple terminal prompts).
func TestConcurrentCheckAll(t *testing.T) {
	const numConcurrent = 10
	var wg sync.WaitGroup

	// Channel to collect results from all concurrent executions
	resultsChan := make(chan []Result, numConcurrent)
	errorsChan := make(chan error, numConcurrent)

	// Launch multiple concurrent CheckAll executions
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()

			// Each execution gets fresh signals
			signals := GetAllSignals()
			ctx := context.Background()

			// Run the checks
			results := CheckAll(ctx, signals)

			// Verify we got results for all signals
			if len(results) != len(signals) {
				errorsChan <- nil
				return
			}

			resultsChan <- results
		}(i)
	}

	// Wait for all executions to complete
	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Verify all executions completed successfully
	resultsCount := 0
	for range resultsChan {
		resultsCount++
	}

	if resultsCount != numConcurrent {
		t.Errorf("Expected %d successful executions, got %d", numConcurrent, resultsCount)
	}
}

// TestConcurrentUmaskCheck specifically tests the umask signal for race conditions
func TestConcurrentUmaskCheck(t *testing.T) {
	const numConcurrent = 50
	var wg sync.WaitGroup

	results := make([]bool, numConcurrent)

	// Launch many concurrent umask checks
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			signal := NewPermissiveUmaskSignal()
			ctx := context.Background()
			results[idx] = signal.Check(ctx)
		}(i)
	}

	wg.Wait()

	// All results should be consistent (same umask value)
	firstResult := results[0]
	for i, result := range results {
		if result != firstResult {
			t.Errorf("Inconsistent umask result at index %d: got %v, expected %v", i, result, firstResult)
		}
	}
}

// TestConcurrentTimeDriftCheck tests the time drift signal for file collision issues
func TestConcurrentTimeDriftCheck(t *testing.T) {
	const numConcurrent = 20
	var wg sync.WaitGroup

	errors := make([]error, numConcurrent)
	results := make([]bool, numConcurrent)

	// Launch many concurrent time drift checks
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Use panic recovery to catch any issues
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in concurrent time drift check %d: %v", idx, r)
				}
			}()

			signal := NewTimeDriftSignal()
			ctx := context.Background()
			results[idx] = signal.Check(ctx)
		}(i)
	}

	wg.Wait()

	// Verify no errors occurred
	for i, err := range errors {
		if err != nil {
			t.Errorf("Error in concurrent time drift check %d: %v", i, err)
		}
	}
}

// TestConcurrentSignalStateIsolation verifies that signal instances don't share state
func TestConcurrentSignalStateIsolation(t *testing.T) {
	const numConcurrent = 10
	var wg sync.WaitGroup

	// Test with a signal that stores mutable state
	diagnostics := make([]string, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each goroutine gets its own signal instance
			signal := NewPermissiveUmaskSignal()
			ctx := context.Background()

			// Run check (which may modify internal state)
			signal.Check(ctx)

			// Read diagnostic (which reads internal state)
			diagnostics[idx] = signal.Diagnostic()

			// Small delay to increase chance of race if one exists
			time.Sleep(time.Millisecond)
		}(i)
	}

	wg.Wait()

	// All diagnostics should be valid (non-empty)
	for i, diag := range diagnostics {
		if diag == "" {
			t.Errorf("Empty diagnostic at index %d", i)
		}
	}
}
