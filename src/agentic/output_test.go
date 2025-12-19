package agentic

import (
	"os"
	"testing"
)

func TestGetAgenticMode(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	tests := []struct {
		envValue string
		want     AgenticMode
	}{
		{"", ModeBlock},
		{"block", ModeBlock},
		{"BLOCK", ModeBlock},
		{"ask", ModeAsk},
		{"ASK", ModeAsk},
		{"Ask", ModeAsk},
		{"invalid", ModeBlock}, // defaults to block
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("DASHLIGHTS_AGENTIC_MODE", tt.envValue)
			if got := GetAgenticMode(); got != tt.want {
				t.Errorf("GetAgenticMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDisabled(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_DISABLE_AGENTIC")
	defer os.Setenv("DASHLIGHTS_DISABLE_AGENTIC", original)

	tests := []struct {
		envValue string
		want     bool
	}{
		{"", false},
		{"1", true},
		{"true", true},
		{"yes", true},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("DASHLIGHTS_DISABLE_AGENTIC", tt.envValue)
			if got := IsDisabled(); got != tt.want {
				t.Errorf("IsDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateOutput_AllowSafe(t *testing.T) {
	result := &AnalysisResult{
		ToolName: "Read",
	}

	output, exitCode, stderrMsg := GenerateOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr, got '%s'", stderrMsg)
	}
	if output == nil {
		t.Fatal("Expected non-nil output")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("Expected 'allow', got '%s'", output.HookSpecificOutput.PermissionDecision)
	}
}

func TestGenerateOutput_WarnTwoCapabilities(t *testing.T) {
	result := &AnalysisResult{
		ToolName:    "Write",
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{"sensitive file"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"state change"}},
	}

	output, exitCode, stderrMsg := GenerateOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr, got '%s'", stderrMsg)
	}
	if output == nil {
		t.Fatal("Expected non-nil output")
	}
	if output.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("Expected 'allow' for warning, got '%s'", output.HookSpecificOutput.PermissionDecision)
	}
	if output.SystemMessage == "" {
		t.Error("Expected non-empty SystemMessage for warning")
	}
}

func TestGenerateOutput_BlockThreeCapabilities(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")

	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"external data"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{"sensitive access"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"state change"}},
	}

	output, exitCode, stderrMsg := GenerateOutput(result)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr message for block")
	}
	if output != nil {
		t.Error("Expected nil output for block")
	}
}

func TestGenerateOutput_AskModeThreeCapabilities(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")

	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"external data"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{"sensitive access"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"state change"}},
	}

	output, exitCode, stderrMsg := GenerateOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for ask mode, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr for ask mode, got '%s'", stderrMsg)
	}
	if output == nil {
		t.Fatal("Expected non-nil output for ask mode")
	}
	if output.HookSpecificOutput.PermissionDecision != "ask" {
		t.Errorf("Expected 'ask', got '%s'", output.HookSpecificOutput.PermissionDecision)
	}
	if output.SystemMessage == "" {
		t.Error("Expected non-empty SystemMessage for ask mode")
	}
}

func TestFormatBlockMessage(t *testing.T) {
	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"curl detected"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{"aws credentials"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"file write"}},
		SignalHits:  []string{"Naked Credential"},
	}

	msg := FormatBlockMessage(result)

	if msg == "" {
		t.Error("Expected non-empty message")
	}
	// Check that message contains key information
	if !contains(msg, "Bash") {
		t.Error("Message should contain tool name")
	}
	if !contains(msg, "A+B+C") {
		t.Error("Message should contain capability string")
	}
	if !contains(msg, "curl detected") {
		t.Error("Message should contain A reason")
	}
	if !contains(msg, "aws credentials") {
		t.Error("Message should contain B reason")
	}
	if !contains(msg, "file write") {
		t.Error("Message should contain C reason")
	}
	if !contains(msg, "Naked Credential") {
		t.Error("Message should contain signal hits")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
