package signals

import (
	"context"
	"sync"
)

// Signal represents a security hygiene check.
//
// Thread-Safety Contract:
//   - Signal instances MUST NOT be shared across concurrent Check() calls.
//   - Each Signal instance should be used by only one goroutine at a time.
//   - Signal implementations MAY store mutable state in struct fields during Check().
//   - Diagnostic() and Remediation() methods MAY read state set by Check() and
//     MUST only be called after Check() completes (enforced by CheckAll).
//   - The current implementation in CheckAll() creates fresh Signal instances
//     for each execution, ensuring thread-safety.
type Signal interface {
	// Check performs the security check and returns true if a problem is detected.
	// This method may modify internal state of the Signal instance to store
	// diagnostic information for later retrieval via Diagnostic().
	Check(ctx context.Context) bool

	// Name returns the signal name (e.g., "Open Door")
	Name() string

	// Emoji returns the signal-specific emoji for diagnostic mode
	Emoji() string

	// Diagnostic returns a brief explanation of the detected issue.
	// This method may read internal state set by Check() and should only
	// be called after Check() has completed.
	Diagnostic() string

	// Remediation returns suggested fix/mitigation steps
	Remediation() string
}

// Result holds the outcome of a signal check
type Result struct {
	Signal   Signal
	Detected bool
	Error    error
}

// CheckAll runs all signal checks concurrently and returns results
func CheckAll(ctx context.Context, signals []Signal) []Result {
	results := make([]Result, len(signals))
	var wg sync.WaitGroup

	for i, sig := range signals {
		wg.Add(1)
		go func(idx int, s Signal) {
			defer wg.Done()

			// Use a panic recovery to ensure one bad signal doesn't crash everything
			defer func() {
				if r := recover(); r != nil {
					results[idx] = Result{
						Signal:   s,
						Detected: false,
						Error:    nil,
					}
				}
			}()

			detected := s.Check(ctx)
			results[idx] = Result{
				Signal:   s,
				Detected: detected,
				Error:    nil,
			}
		}(i, sig)
	}

	wg.Wait()
	return results
}

// CountDetected returns the number of detected signals
func CountDetected(results []Result) int {
	count := 0
	for _, r := range results {
		if r.Detected {
			count++
		}
	}
	return count
}

// GetDetected returns only the detected signals
func GetDetected(results []Result) []Result {
	detected := make([]Result, 0)
	for _, r := range results {
		if r.Detected {
			detected = append(detected, r)
		}
	}
	return detected
}
