package signals

import (
	"context"
	"testing"
	"time"
)

// mockSignal is a test implementation of Signal
type mockSignal struct {
	name         string
	emoji        string
	diagnostic   string
	remediation  string
	shouldDetect bool
	delay        time.Duration
}

func (m *mockSignal) Check(_ context.Context) bool {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.shouldDetect
}

func (m *mockSignal) Name() string        { return m.name }
func (m *mockSignal) Emoji() string       { return m.emoji }
func (m *mockSignal) Diagnostic() string  { return m.diagnostic }
func (m *mockSignal) Remediation() string { return m.remediation }

func TestCheckAll(t *testing.T) {
	signals := []Signal{
		&mockSignal{name: "test1", shouldDetect: true},
		&mockSignal{name: "test2", shouldDetect: false},
		&mockSignal{name: "test3", shouldDetect: true},
	}

	ctx := context.Background()
	results := CheckAll(ctx, signals)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if !results[0].Detected {
		t.Error("Expected first signal to be detected")
	}

	if results[1].Detected {
		t.Error("Expected second signal to not be detected")
	}

	if !results[2].Detected {
		t.Error("Expected third signal to be detected")
	}
}

func TestCheckAllConcurrency(t *testing.T) {
	// Create signals with delays to verify concurrent execution
	signals := []Signal{
		&mockSignal{name: "slow1", shouldDetect: true, delay: 10 * time.Millisecond},
		&mockSignal{name: "slow2", shouldDetect: true, delay: 10 * time.Millisecond},
		&mockSignal{name: "slow3", shouldDetect: true, delay: 10 * time.Millisecond},
	}

	ctx := context.Background()
	start := time.Now()
	results := CheckAll(ctx, signals)
	elapsed := time.Since(start)

	// If running concurrently, should take ~10ms, not ~30ms
	if elapsed > 25*time.Millisecond {
		t.Errorf("CheckAll took too long (%v), may not be running concurrently", elapsed)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestCountDetected(t *testing.T) {
	results := []Result{
		{Detected: true},
		{Detected: false},
		{Detected: true},
		{Detected: true},
	}

	count := CountDetected(results)
	if count != 3 {
		t.Errorf("Expected count of 3, got %d", count)
	}
}

func TestGetDetected(t *testing.T) {
	results := []Result{
		{Signal: &mockSignal{name: "sig1"}, Detected: true},
		{Signal: &mockSignal{name: "sig2"}, Detected: false},
		{Signal: &mockSignal{name: "sig3"}, Detected: true},
	}

	detected := GetDetected(results)
	if len(detected) != 2 {
		t.Errorf("Expected 2 detected signals, got %d", len(detected))
	}

	if detected[0].Signal.Name() != "sig1" {
		t.Errorf("Expected first detected signal to be 'sig1', got '%s'", detected[0].Signal.Name())
	}

	if detected[1].Signal.Name() != "sig3" {
		t.Errorf("Expected second detected signal to be 'sig3', got '%s'", detected[1].Signal.Name())
	}
}
