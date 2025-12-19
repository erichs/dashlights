package agentic

import (
	"encoding/json"
	"os"
)

// AgentType represents the type of AI coding assistant.
type AgentType string

const (
	// AgentUnknown indicates the agent type could not be determined.
	AgentUnknown AgentType = "unknown"
	// AgentClaudeCode indicates Claude Code (Anthropic's CLI).
	AgentClaudeCode AgentType = "claude_code"
	// AgentCursor indicates Cursor IDE.
	AgentCursor AgentType = "cursor"
)

// DetectAgent returns the agent type based on environment variables.
// Priority: CURSOR_AGENT=1 > CLAUDECODE=1 > unknown
func DetectAgent() AgentType {
	if os.Getenv("CURSOR_AGENT") == "1" {
		return AgentCursor
	}
	if os.Getenv("CLAUDECODE") == "1" {
		return AgentClaudeCode
	}
	return AgentUnknown
}

// DetectAgentFromInput attempts to determine the agent type from the JSON input structure.
// This is used as a fallback when environment variable detection fails.
func DetectAgentFromInput(raw []byte) AgentType {
	if len(raw) == 0 {
		return AgentUnknown
	}

	// Probe the JSON structure to identify the agent
	var probe struct {
		// Claude Code fields
		ToolName      string `json:"tool_name"`
		HookEventName string `json:"hook_event_name"`

		// Cursor fields
		Command       string `json:"command"`
		CursorVersion string `json:"cursor_version"`
	}

	if err := json.Unmarshal(raw, &probe); err != nil {
		return AgentUnknown
	}

	// Cursor: has cursor_version or hook_event_name is "beforeShellExecution"
	if probe.CursorVersion != "" {
		return AgentCursor
	}
	if probe.HookEventName == "beforeShellExecution" || probe.HookEventName == "beforeMCPExecution" {
		return AgentCursor
	}

	// Claude Code: has tool_name field and hook_event_name is "PreToolUse"
	if probe.ToolName != "" && (probe.HookEventName == "PreToolUse" || probe.HookEventName == "") {
		return AgentClaudeCode
	}

	// If there's a command but no tool_name, it's likely Cursor shell input
	if probe.Command != "" && probe.ToolName == "" {
		return AgentCursor
	}

	return AgentUnknown
}
