package agentic

import (
	"os"
	"testing"
)

func TestDetectAgent(t *testing.T) {
	// Save original values
	originalCursor := os.Getenv("CURSOR_AGENT")
	originalClaude := os.Getenv("CLAUDECODE")
	defer func() {
		os.Setenv("CURSOR_AGENT", originalCursor)
		os.Setenv("CLAUDECODE", originalClaude)
	}()

	tests := []struct {
		name        string
		cursorAgent string
		claudeCode  string
		want        AgentType
	}{
		{
			name:        "Cursor agent detected",
			cursorAgent: "1",
			claudeCode:  "",
			want:        AgentCursor,
		},
		{
			name:        "Claude Code detected",
			cursorAgent: "",
			claudeCode:  "1",
			want:        AgentClaudeCode,
		},
		{
			name:        "Cursor takes priority over Claude",
			cursorAgent: "1",
			claudeCode:  "1",
			want:        AgentCursor,
		},
		{
			name:        "No agent detected",
			cursorAgent: "",
			claudeCode:  "",
			want:        AgentUnknown,
		},
		{
			name:        "Cursor with wrong value",
			cursorAgent: "true", // Not "1"
			claudeCode:  "",
			want:        AgentUnknown,
		},
		{
			name:        "Claude with wrong value",
			cursorAgent: "",
			claudeCode:  "true", // Not "1"
			want:        AgentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CURSOR_AGENT", tt.cursorAgent)
			os.Setenv("CLAUDECODE", tt.claudeCode)

			got := DetectAgent()
			if got != tt.want {
				t.Errorf("DetectAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectAgentFromInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  AgentType
	}{
		// Cursor inputs
		{
			name: "Cursor beforeShellExecution",
			input: `{
				"conversation_id": "",
				"command": "ps -ef",
				"cwd": "/tmp",
				"hook_event_name": "beforeShellExecution",
				"cursor_version": "1.0.0"
			}`,
			want: AgentCursor,
		},
		{
			name: "Cursor with cursor_version only",
			input: `{
				"command": "ls",
				"cursor_version": "1.0.0"
			}`,
			want: AgentCursor,
		},
		{
			name: "Cursor shell without hook_event_name but has command",
			input: `{
				"command": "echo hello",
				"cwd": "/tmp"
			}`,
			want: AgentCursor,
		},

		// Claude Code inputs
		{
			name: "Claude Code PreToolUse",
			input: `{
				"session_id": "abc123",
				"tool_name": "Bash",
				"tool_input": {"command": "ls"},
				"hook_event_name": "PreToolUse"
			}`,
			want: AgentClaudeCode,
		},
		{
			name: "Claude Code with tool_name only",
			input: `{
				"tool_name": "Write",
				"tool_input": {"file_path": "test.txt"}
			}`,
			want: AgentClaudeCode,
		},

		// Unknown/invalid inputs
		{
			name:  "Empty input",
			input: "",
			want:  AgentUnknown,
		},
		{
			name:  "Invalid JSON",
			input: "{not valid json}",
			want:  AgentUnknown,
		},
		{
			name:  "Empty JSON object",
			input: "{}",
			want:  AgentUnknown,
		},
		{
			name: "Ambiguous input",
			input: `{
				"some_field": "value"
			}`,
			want: AgentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectAgentFromInput([]byte(tt.input))
			if got != tt.want {
				t.Errorf("DetectAgentFromInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentTypeConstants(t *testing.T) {
	// Verify constant values are distinct and meaningful
	if AgentUnknown == AgentClaudeCode {
		t.Error("AgentUnknown should not equal AgentClaudeCode")
	}
	if AgentUnknown == AgentCursor {
		t.Error("AgentUnknown should not equal AgentCursor")
	}
	if AgentClaudeCode == AgentCursor {
		t.Error("AgentClaudeCode should not equal AgentCursor")
	}

	// Verify string representations
	if string(AgentUnknown) != "unknown" {
		t.Errorf("AgentUnknown = %q, want %q", AgentUnknown, "unknown")
	}
	if string(AgentClaudeCode) != "claude_code" {
		t.Errorf("AgentClaudeCode = %q, want %q", AgentClaudeCode, "claude_code")
	}
	if string(AgentCursor) != "cursor" {
		t.Errorf("AgentCursor = %q, want %q", AgentCursor, "cursor")
	}
}
