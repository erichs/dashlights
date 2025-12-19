package agentic

import (
	"fmt"
	"os"
	"strings"
)

// AgenticMode controls behavior when Rule of Two is violated.
type AgenticMode string

const (
	// ModeBlock blocks the action with exit code 2.
	ModeBlock AgenticMode = "block"
	// ModeAsk prompts user for confirmation instead of blocking.
	ModeAsk AgenticMode = "ask"
)

// HookOutput represents the JSON output for Claude Code PreToolUse hooks.
type HookOutput struct {
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
}

// HookSpecificOutput contains PreToolUse-specific response fields.
type HookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

// GetAgenticMode returns the configured agentic mode from environment.
func GetAgenticMode() AgenticMode {
	mode := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	if strings.ToLower(mode) == "ask" {
		return ModeAsk
	}
	return ModeBlock // default
}

// IsDisabled returns true if agentic mode is disabled via environment.
func IsDisabled() bool {
	return os.Getenv("DASHLIGHTS_DISABLE_AGENTIC") != ""
}

// GenerateOutput creates the appropriate hook output based on analysis results.
// Returns (output, exitCode, stderrMessage).
// - exitCode 0: allow (with optional systemMessage warning)
// - exitCode 2: block (stderrMessage contains error)
func GenerateOutput(result *AnalysisResult) (*HookOutput, int, string) {
	count := result.CapabilityCount()
	mode := GetAgenticMode()

	switch {
	case count >= 3:
		// Rule of Two violation - all three capabilities detected
		return generateViolationOutput(result, mode)

	case count == 2:
		// Two capabilities - warn but allow
		return generateWarningOutput(result), 0, ""

	default:
		// Zero or one capability - allow silently
		return generateAllowOutput(), 0, ""
	}
}

// generateViolationOutput handles the case where all three capabilities are detected.
func generateViolationOutput(result *AnalysisResult, mode AgenticMode) (*HookOutput, int, string) {
	reasons := result.AllReasons()
	reasonStr := strings.Join(reasons, "; ")

	if mode == ModeBlock {
		// Hard block with exit code 2
		stderrMsg := fmt.Sprintf(
			"🚫 Rule of Two Violation: %s combines all three capabilities "+
				"(A: untrustworthy input, B: sensitive access, C: state change). "+
				"Reasons: %s",
			result.ToolName, reasonStr)
		return nil, 2, stderrMsg
	}

	// Ask mode - prompt user instead of blocking
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "ask",
			PermissionDecisionReason: fmt.Sprintf(
				"Rule of Two: %s combines A+B+C capabilities. Reasons: %s",
				result.ToolName, reasonStr),
		},
		SystemMessage: fmt.Sprintf(
			"⚠️ Rule of Two Violation: %s combines all three capabilities (A+B+C). "+
				"This action processes untrustworthy input, accesses sensitive data, "+
				"AND changes state. Reasons: %s",
			result.ToolName, reasonStr),
	}, 0, ""
}

// generateWarningOutput creates output for two-capability warnings.
func generateWarningOutput(result *AnalysisResult) *HookOutput {
	caps := result.CapabilityString()
	reasons := result.AllReasons()
	reasonStr := strings.Join(reasons, "; ")

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			PermissionDecisionReason: fmt.Sprintf(
				"Rule of Two: %s combines %s capabilities (2 of 3)",
				result.ToolName, caps),
		},
		SystemMessage: fmt.Sprintf(
			"⚠️ Rule of Two: %s combines %s capabilities. Reasons: %s",
			result.ToolName, caps, reasonStr),
	}
}

// generateAllowOutput creates output for safe operations.
func generateAllowOutput() *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "allow",
			PermissionDecisionReason: "Rule of Two: OK",
		},
	}
}

// FormatBlockMessage creates a formatted error message for blocked operations.
func FormatBlockMessage(result *AnalysisResult) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Tool: %s", result.ToolName))
	parts = append(parts, fmt.Sprintf("Capabilities: %s", result.CapabilityString()))

	if result.CapabilityA.Detected {
		parts = append(parts, fmt.Sprintf("  [A] Untrustworthy input: %s",
			strings.Join(result.CapabilityA.Reasons, ", ")))
	}
	if result.CapabilityB.Detected {
		parts = append(parts, fmt.Sprintf("  [B] Sensitive access: %s",
			strings.Join(result.CapabilityB.Reasons, ", ")))
	}
	if result.CapabilityC.Detected {
		parts = append(parts, fmt.Sprintf("  [C] State change: %s",
			strings.Join(result.CapabilityC.Reasons, ", ")))
	}

	if len(result.SignalHits) > 0 {
		parts = append(parts, fmt.Sprintf("  Signals: %s",
			strings.Join(result.SignalHits, ", ")))
	}

	return strings.Join(parts, "\n")
}
