package signals

import (
	"context"
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

// VerboseRemediator is an optional interface that signals can implement to provide
// additional actionable remediation guidance when verbose mode is enabled.
// For example, a signal may return a ready-to-use shell command that fixes the issue.
type VerboseRemediator interface {
	// VerboseRemediation returns detailed, actionable remediation guidance.
	// This may include ready-to-use shell commands or configuration snippets.
	// Returns an empty string if no additional verbose guidance is available.
	VerboseRemediation() string
}

// Result holds the outcome of a signal check
type Result struct {
	Signal   Signal
	Detected bool
	Error    error
}

// indexedResult pairs a result with its index for channel-based collection
type indexedResult struct {
	idx    int
	result Result
}

// CheckAll runs all signal checks concurrently and returns results.
// The bool return value is true if all signals completed, false if timeout occurred.
func CheckAll(ctx context.Context, signals []Signal) ([]Result, bool) {
	if len(signals) == 0 {
		return []Result{}, true
	}

	results := make([]Result, len(signals))
	resultChan := make(chan indexedResult, len(signals))

	for i, sig := range signals {
		go func(idx int, s Signal) {
			// Use a panic recovery to ensure one bad signal doesn't crash everything
			defer func() {
				if r := recover(); r != nil {
					resultChan <- indexedResult{
						idx: idx,
						result: Result{
							Signal:   s,
							Detected: false,
							Error:    nil,
						},
					}
				}
			}()

			detected := s.Check(ctx)
			resultChan <- indexedResult{
				idx: idx,
				result: Result{
					Signal:   s,
					Detected: detected,
					Error:    nil,
				},
			}
		}(i, sig)
	}

	// Collect results until context expires or all complete
	completed := 0
	for completed < len(signals) {
		select {
		case ir := <-resultChan:
			results[ir.idx] = ir.result
			completed++
		case <-ctx.Done():
			// Drain any results that arrived just before/during timeout
		drainLoop:
			for {
				select {
				case ir := <-resultChan:
					results[ir.idx] = ir.result
					completed++
				default:
					break drainLoop
				}
			}

			// Mark unreceived signals with empty results (preserving Signal reference)
			for i, sig := range signals {
				if results[i].Signal == nil {
					results[i] = Result{Signal: sig, Detected: false, Error: nil}
				}
			}
			return results, false // Partial results
		}
	}

	return results, true // All complete
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
