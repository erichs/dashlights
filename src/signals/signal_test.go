package signals

import (
	"context"
	"os"
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

// TestSignalDisableEnvVar tests that signals can be disabled via environment variables
func TestSignalDisableEnvVar(t *testing.T) {
	testCases := []struct {
		name       string
		envVar     string
		newSignal  func() Signal
		setupCheck func() // Optional setup to ensure signal would normally trigger
		cleanup    func() // Optional cleanup after test
	}{
		{
			name:      "DebugEnabled",
			envVar:    "DASHLIGHTS_DISABLE_DEBUG_ENABLED",
			newSignal: func() Signal { return NewDebugEnabledSignal() },
			setupCheck: func() {
				os.Setenv("DEBUG", "1")
			},
			cleanup: func() {
				os.Unsetenv("DEBUG")
			},
		},
		{
			name:      "ProxyActive",
			envVar:    "DASHLIGHTS_DISABLE_PROXY_ACTIVE",
			newSignal: func() Signal { return NewProxyActiveSignal() },
			setupCheck: func() {
				os.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
			},
			cleanup: func() {
				os.Unsetenv("HTTP_PROXY")
			},
		},
		{
			name:      "HistoryDisabled",
			envVar:    "DASHLIGHTS_DISABLE_HISTORY_DISABLED",
			newSignal: func() Signal { return NewHistoryDisabledSignal() },
			setupCheck: func() {
				os.Setenv("HISTFILE", "/dev/null")
			},
			cleanup: func() {
				os.Unsetenv("HISTFILE")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup: ensure signal would normally trigger
			if tc.setupCheck != nil {
				tc.setupCheck()
			}
			if tc.cleanup != nil {
				defer tc.cleanup()
			}

			ctx := context.Background()

			// First, verify signal triggers without disable env var
			signal := tc.newSignal()
			resultWithoutDisable := signal.Check(ctx)
			if !resultWithoutDisable {
				t.Skipf("Signal %s did not trigger in test environment, skipping disable test", tc.name)
			}

			// Now set the disable env var
			os.Setenv(tc.envVar, "1")
			defer os.Unsetenv(tc.envVar)

			// Create new signal instance and check again
			signal = tc.newSignal()
			resultWithDisable := signal.Check(ctx)

			if resultWithDisable {
				t.Errorf("Signal %s should return false when %s is set", tc.name, tc.envVar)
			}
		})
	}
}

// TestSignalDisableEnvVarAnyValue tests that any non-empty value disables the signal
func TestSignalDisableEnvVarAnyValue(t *testing.T) {
	testValues := []string{"1", "true", "yes", "anything", "0", "false"}

	for _, value := range testValues {
		t.Run("value_"+value, func(t *testing.T) {
			// Use DebugEnabled as a representative signal
			os.Setenv("DEBUG", "1") // Ensure signal would trigger
			defer os.Unsetenv("DEBUG")

			os.Setenv("DASHLIGHTS_DISABLE_DEBUG_ENABLED", value)
			defer os.Unsetenv("DASHLIGHTS_DISABLE_DEBUG_ENABLED")

			signal := NewDebugEnabledSignal()
			ctx := context.Background()

			if signal.Check(ctx) {
				t.Errorf("Signal should be disabled when DASHLIGHTS_DISABLE_DEBUG_ENABLED=%s", value)
			}
		})
	}
}

// TestSignalNotDisabledWhenEnvVarEmpty tests that empty env var does not disable signal
func TestSignalNotDisabledWhenEnvVarEmpty(t *testing.T) {
	// Ensure the disable env var is not set
	os.Unsetenv("DASHLIGHTS_DISABLE_DEBUG_ENABLED")

	// Set DEBUG to trigger the signal
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	signal := NewDebugEnabledSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Signal should trigger when disable env var is not set")
	}
}
