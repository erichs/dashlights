// Package agentic provides security analysis for AI coding assistants.
// It detects critical threats (config writes, invisible unicode) and performs
// Rule of Two analysis to detect potential security violations where an action
// combines more than two of: [A] untrustworthy inputs, [B] sensitive access,
// [C] state changes or external communication.
package agentic

// HookInput represents the JSON input from Claude Code PreToolUse hook.
// This structure matches the JSON schema provided by Claude Code's hook system.
type HookInput struct {
	SessionID      string                 `json:"session_id"`
	TranscriptPath string                 `json:"transcript_path,omitempty"`
	Cwd            string                 `json:"cwd"`
	HookEventName  string                 `json:"hook_event_name"`
	ToolName       string                 `json:"tool_name"`
	ToolInput      map[string]interface{} `json:"tool_input"`
	ToolUseID      string                 `json:"tool_use_id,omitempty"`
}

// WriteInput represents the tool_input for Write tool calls.
type WriteInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// EditInput represents the tool_input for Edit tool calls.
type EditInput struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// BashInput represents the tool_input for Bash tool calls.
type BashInput struct {
	Command     string `json:"command"`
	Description string `json:"description,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
}

// ReadInput represents the tool_input for Read tool calls.
type ReadInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// WebFetchInput represents the tool_input for WebFetch tool calls.
type WebFetchInput struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// WebSearchInput represents the tool_input for WebSearch tool calls.
type WebSearchInput struct {
	Query string `json:"query"`
}

// GrepInput represents the tool_input for Grep tool calls.
type GrepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Glob    string `json:"glob,omitempty"`
}

// GlobInput represents the tool_input for Glob tool calls.
type GlobInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// ParseWriteInput extracts WriteInput from generic tool_input map.
func ParseWriteInput(input map[string]interface{}) WriteInput {
	return WriteInput{
		FilePath: getStringField(input, "file_path"),
		Content:  getStringField(input, "content"),
	}
}

// ParseEditInput extracts EditInput from generic tool_input map.
func ParseEditInput(input map[string]interface{}) EditInput {
	return EditInput{
		FilePath:  getStringField(input, "file_path"),
		OldString: getStringField(input, "old_string"),
		NewString: getStringField(input, "new_string"),
	}
}

// ParseBashInput extracts BashInput from generic tool_input map.
func ParseBashInput(input map[string]interface{}) BashInput {
	return BashInput{
		Command:     getStringField(input, "command"),
		Description: getStringField(input, "description"),
		Timeout:     getIntField(input, "timeout"),
	}
}

// ParseReadInput extracts ReadInput from generic tool_input map.
func ParseReadInput(input map[string]interface{}) ReadInput {
	return ReadInput{
		FilePath: getStringField(input, "file_path"),
		Offset:   getIntField(input, "offset"),
		Limit:    getIntField(input, "limit"),
	}
}

// ParseWebFetchInput extracts WebFetchInput from generic tool_input map.
func ParseWebFetchInput(input map[string]interface{}) WebFetchInput {
	return WebFetchInput{
		URL:    getStringField(input, "url"),
		Prompt: getStringField(input, "prompt"),
	}
}

// ParseWebSearchInput extracts WebSearchInput from generic tool_input map.
func ParseWebSearchInput(input map[string]interface{}) WebSearchInput {
	return WebSearchInput{
		Query: getStringField(input, "query"),
	}
}

// ParseGrepInput extracts GrepInput from generic tool_input map.
func ParseGrepInput(input map[string]interface{}) GrepInput {
	return GrepInput{
		Pattern: getStringField(input, "pattern"),
		Path:    getStringField(input, "path"),
		Glob:    getStringField(input, "glob"),
	}
}

// ParseGlobInput extracts GlobInput from generic tool_input map.
func ParseGlobInput(input map[string]interface{}) GlobInput {
	return GlobInput{
		Pattern: getStringField(input, "pattern"),
		Path:    getStringField(input, "path"),
	}
}

// getStringField safely extracts a string field from a map.
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getIntField safely extracts an int field from a map.
func getIntField(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return 0
}
