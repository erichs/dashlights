package agentic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CursorShellInput represents the input format for Cursor beforeShellExecution hook.
type CursorShellInput struct {
	ConversationID string   `json:"conversation_id"`
	GenerationID   string   `json:"generation_id"`
	Model          string   `json:"model"`
	Command        string   `json:"command"`
	Cwd            string   `json:"cwd"`
	HookEventName  string   `json:"hook_event_name"`
	CursorVersion  string   `json:"cursor_version"`
	WorkspaceRoots []string `json:"workspace_roots"`
	UserEmail      *string  `json:"user_email"`
}

// CursorOutput represents the output format expected by Cursor hooks.
type CursorOutput struct {
	Permission   string `json:"permission"`              // "allow", "deny", "ask"
	UserMessage  string `json:"user_message,omitempty"`  // Shown in client
	AgentMessage string `json:"agent_message,omitempty"` // Sent to agent
}

// ParseCursorInput parses Cursor hook input and normalizes it to HookInput.
func ParseCursorInput(raw []byte) (*HookInput, error) {
	var input CursorShellInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid Cursor input: %w", err)
	}

	// Normalize to canonical HookInput (same format as Claude Code)
	return &HookInput{
		SessionID:     input.ConversationID,
		Cwd:           input.Cwd,
		HookEventName: input.HookEventName,
		ToolName:      "Bash", // Cursor shell commands map to Bash tool
		ToolInput: map[string]interface{}{
			"command": input.Command,
		},
	}, nil
}

// GenerateCursorOutput converts analysis results to Cursor output format.
// Returns (jsonOutput, exitCode, stderrMessage).
func GenerateCursorOutput(result *AnalysisResult) ([]byte, int, string) {
	count := result.CapabilityCount()
	mode := GetAgenticMode()

	switch {
	case count >= 3:
		return generateCursorViolationOutput(result, mode)
	case count == 2:
		return generateCursorWarningOutput(result)
	default:
		return generateCursorAllowOutput()
	}
}

// generateCursorViolationOutput handles Rule of Two violations for Cursor.
func generateCursorViolationOutput(result *AnalysisResult, mode AgenticMode) ([]byte, int, string) {
	reasons := result.AllReasons()
	reasonStr := strings.Join(reasons, "; ")

	if mode == ModeBlock {
		// Hard block with exit code 2
		output := CursorOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Rule of Two Violation: %s combines all three capabilities (A+B+C). Reasons: %s", result.ToolName, reasonStr),
		}
		stderrMsg := fmt.Sprintf("Rule of Two Violation: %s combines A+B+C. %s", result.ToolName, reasonStr)
		return marshalCursorOutput(output), 2, stderrMsg
	}

	// Ask mode - prompt user instead of blocking
	output := CursorOutput{
		Permission:   "ask",
		UserMessage:  fmt.Sprintf("Rule of Two: %s combines all three capabilities. Confirm?", result.ToolName),
		AgentMessage: fmt.Sprintf("Security check triggered. Reasons: %s", reasonStr),
	}
	return marshalCursorOutput(output), 0, ""
}

// generateCursorWarningOutput creates output for two-capability warnings.
func generateCursorWarningOutput(result *AnalysisResult) ([]byte, int, string) {
	caps := result.CapabilityString()
	reasons := result.AllReasons()
	reasonStr := strings.Join(reasons, "; ")

	output := CursorOutput{
		Permission:   "allow",
		AgentMessage: fmt.Sprintf("Rule of Two: %s combines %s capabilities. Reasons: %s", result.ToolName, caps, reasonStr),
	}
	return marshalCursorOutput(output), 0, ""
}

// generateCursorAllowOutput creates output for safe operations.
func generateCursorAllowOutput() ([]byte, int, string) {
	output := CursorOutput{
		Permission: "allow",
	}
	return marshalCursorOutput(output), 0, ""
}

// GenerateCursorThreatOutput converts critical threat to Cursor output format.
// Returns (jsonOutput, exitCode, stderrMessage).
func GenerateCursorThreatOutput(threat *CriticalThreat) ([]byte, int, string) {
	mode := GetAgenticMode()

	switch threat.Type {
	case "agent_config_write":
		// Always block, never ask
		output := CursorOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Blocked: %s", threat.Details),
		}
		return marshalCursorOutput(output), 2, fmt.Sprintf("Blocked: Attempted write to agent configuration. %s", threat.Details)

	case "invisible_unicode":
		if mode == ModeAsk && threat.AllowAskMode {
			// Ask mode - prompt user
			output := CursorOutput{
				Permission:   "ask",
				UserMessage:  fmt.Sprintf("Invisible Unicode detected: %s", threat.Details),
				AgentMessage: "Security check: invisible characters detected in input",
			}
			return marshalCursorOutput(output), 0, ""
		}

		// Block mode (default)
		output := CursorOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Blocked: Invisible Unicode detected. %s", threat.Details),
		}
		return marshalCursorOutput(output), 2, fmt.Sprintf("Blocked: Invisible Unicode detected. %s", threat.Details)

	default:
		// Unknown threat type - block to be safe
		output := CursorOutput{
			Permission:  "deny",
			UserMessage: fmt.Sprintf("Blocked: %s", threat.Details),
		}
		return marshalCursorOutput(output), 2, fmt.Sprintf("Blocked: Unknown critical threat: %s", threat.Type)
	}
}

// GenerateCursorDisabledOutput creates output when agentic checks are disabled.
func GenerateCursorDisabledOutput() ([]byte, int, string) {
	output := CursorOutput{
		Permission: "allow",
	}
	return marshalCursorOutput(output), 0, ""
}

// marshalCursorOutput marshals CursorOutput to JSON.
// This struct has fixed fields that cannot fail to marshal, so we return empty
// JSON on error rather than propagating it (which would complicate all callers).
func marshalCursorOutput(output CursorOutput) []byte {
	jsonOut, err := json.Marshal(output)
	if err != nil {
		// This should never happen with a simple struct like CursorOutput,
		// but return a valid allow response as fallback
		return []byte(`{"permission":"allow"}`)
	}
	return jsonOut
}
