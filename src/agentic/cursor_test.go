package agentic

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseCursorInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantTool  string
		wantCmd   string
		wantCwd   string
		wantEvent string
	}{
		{
			name: "Valid shell execution",
			input: `{
				"conversation_id": "conv123",
				"generation_id": "gen456",
				"model": "unknown",
				"command": "ps -ef",
				"cwd": "/Users/test/project",
				"hook_event_name": "beforeShellExecution",
				"cursor_version": "1.0.0",
				"workspace_roots": ["/Users/test/project"],
				"user_email": null
			}`,
			wantErr:   false,
			wantTool:  "Bash",
			wantCmd:   "ps -ef",
			wantCwd:   "/Users/test/project",
			wantEvent: "beforeShellExecution",
		},
		{
			name: "Complex command",
			input: `{
				"command": "curl ipinfo.io | jq .ip",
				"cwd": "/tmp",
				"hook_event_name": "beforeShellExecution",
				"cursor_version": "1.0.0"
			}`,
			wantErr:   false,
			wantTool:  "Bash",
			wantCmd:   "curl ipinfo.io | jq .ip",
			wantCwd:   "/tmp",
			wantEvent: "beforeShellExecution",
		},
		{
			name: "Minimal input",
			input: `{
				"command": "ls",
				"cwd": "."
			}`,
			wantErr:  false,
			wantTool: "Bash",
			wantCmd:  "ls",
			wantCwd:  ".",
		},
		{
			name:    "Invalid JSON",
			input:   "{not valid json}",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookInput, err := ParseCursorInput([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if hookInput.ToolName != tt.wantTool {
				t.Errorf("ToolName = %q, want %q", hookInput.ToolName, tt.wantTool)
			}

			cmd := getStringField(hookInput.ToolInput, "command")
			if cmd != tt.wantCmd {
				t.Errorf("command = %q, want %q", cmd, tt.wantCmd)
			}

			if hookInput.Cwd != tt.wantCwd {
				t.Errorf("Cwd = %q, want %q", hookInput.Cwd, tt.wantCwd)
			}

			if tt.wantEvent != "" && hookInput.HookEventName != tt.wantEvent {
				t.Errorf("HookEventName = %q, want %q", hookInput.HookEventName, tt.wantEvent)
			}
		})
	}
}

func TestGenerateCursorOutput_Allow(t *testing.T) {
	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: false},
		CapabilityB: CapabilityResult{Detected: false},
		CapabilityC: CapabilityResult{Detected: false},
	}

	jsonOut, exitCode, stderrMsg := GenerateCursorOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr, got %q", stderrMsg)
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
}

func TestGenerateCursorOutput_Warning(t *testing.T) {
	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"curl"}},
		CapabilityB: CapabilityResult{Detected: false},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"network"}},
	}

	jsonOut, exitCode, stderrMsg := GenerateCursorOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr, got %q", stderrMsg)
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
	if output.AgentMessage == "" {
		t.Error("Expected non-empty agent_message for warning")
	}
}

func TestGenerateCursorOutput_Block(t *testing.T) {
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"curl"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{".aws/"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"redirect"}},
	}

	// Test block mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")
	jsonOut, exitCode, stderrMsg := GenerateCursorOutput(result)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr message")
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}
	if output.UserMessage == "" {
		t.Error("Expected non-empty user_message")
	}
}

func TestGenerateCursorOutput_Ask(t *testing.T) {
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	result := &AnalysisResult{
		ToolName:    "Bash",
		CapabilityA: CapabilityResult{Detected: true, Reasons: []string{"curl"}},
		CapabilityB: CapabilityResult{Detected: true, Reasons: []string{".aws/"}},
		CapabilityC: CapabilityResult{Detected: true, Reasons: []string{"redirect"}},
	}

	// Test ask mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")
	jsonOut, exitCode, stderrMsg := GenerateCursorOutput(result)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr in ask mode, got %q", stderrMsg)
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "ask" {
		t.Errorf("Permission = %q, want %q", output.Permission, "ask")
	}
}

func TestGenerateCursorThreatOutput_AgentConfig(t *testing.T) {
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	threat := &CriticalThreat{
		Type:         "agent_config_write",
		Details:      "Write to CLAUDE.md",
		AllowAskMode: false,
	}

	// Test block mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")
	jsonOut, exitCode, stderrMsg := GenerateCursorThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr")
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}

	// Test ask mode - should still block for config writes
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")
	_, exitCode, _ = GenerateCursorThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 even in ask mode, got %d", exitCode)
	}
}

func TestGenerateCursorThreatOutput_InvisibleUnicode(t *testing.T) {
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	threat := &CriticalThreat{
		Type:         "invisible_unicode",
		Details:      "Zero-width space detected",
		AllowAskMode: true,
	}

	// Test block mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")
	jsonOut, exitCode, stderrMsg := GenerateCursorThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr")
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "deny" {
		t.Errorf("Permission = %q, want %q", output.Permission, "deny")
	}

	// Test ask mode - should prompt for invisible unicode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")
	jsonOut, exitCode, _ = GenerateCursorThreatOutput(threat)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 in ask mode, got %d", exitCode)
	}

	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "ask" {
		t.Errorf("Permission = %q, want %q", output.Permission, "ask")
	}
}

func TestGenerateCursorDisabledOutput(t *testing.T) {
	jsonOut, exitCode, stderrMsg := GenerateCursorDisabledOutput()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderrMsg != "" {
		t.Errorf("Expected empty stderr, got %q", stderrMsg)
	}

	var output CursorOutput
	if err := json.Unmarshal(jsonOut, &output); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if output.Permission != "allow" {
		t.Errorf("Permission = %q, want %q", output.Permission, "allow")
	}
}

func TestCursorOutputFormat(t *testing.T) {
	// Verify the output JSON structure matches Cursor's expectations
	output := CursorOutput{
		Permission:   "deny",
		UserMessage:  "Test user message",
		AgentMessage: "Test agent message",
	}

	jsonOut, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Parse back to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonOut, &parsed); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	// Check field names match Cursor's expected format
	if _, ok := parsed["permission"]; !ok {
		t.Error("Expected 'permission' field")
	}
	if _, ok := parsed["user_message"]; !ok {
		t.Error("Expected 'user_message' field")
	}
	if _, ok := parsed["agent_message"]; !ok {
		t.Error("Expected 'agent_message' field")
	}
}

func TestCursorOutputOmitsEmptyFields(t *testing.T) {
	// Verify empty fields are omitted from JSON
	output := CursorOutput{
		Permission: "allow",
		// UserMessage and AgentMessage are empty
	}

	jsonOut, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonOut, &parsed); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	// Empty fields should be omitted
	if _, ok := parsed["user_message"]; ok {
		t.Error("Expected 'user_message' to be omitted when empty")
	}
	if _, ok := parsed["agent_message"]; ok {
		t.Error("Expected 'agent_message' to be omitted when empty")
	}
}
