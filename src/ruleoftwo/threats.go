package ruleoftwo

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// CriticalThreat represents a security threat that bypasses Rule of Two scoring.
// These are threats that warrant immediate blocking regardless of capability count.
type CriticalThreat struct {
	Type    string // "claude_config_write", "invisible_unicode"
	Details string
	// AllowAskMode indicates whether DASHLIGHTS_AGENTIC_MODE=ask should prompt
	// instead of blocking. Claude config writes always block (false).
	AllowAskMode bool
}

// InvisibleCharInfo describes a detected invisible Unicode character.
type InvisibleCharInfo struct {
	Rune     rune
	Name     string
	Position int
	Context  string // surrounding characters for display
}

// invisibleUnicodeRange defines a range of invisible Unicode characters.
type invisibleUnicodeRange struct {
	Name  string
	Start rune
	End   rune
}

// invisibleUnicodeRanges defines the suspicious Unicode character ranges.
// These characters are invisible or can be used for text spoofing attacks.
var invisibleUnicodeRanges = []invisibleUnicodeRange{
	{"Zero-width space", 0x200B, 0x200B},
	{"Zero-width non-joiner", 0x200C, 0x200C},
	{"Zero-width joiner", 0x200D, 0x200D},
	{"Word joiner", 0x2060, 0x2060},
	{"Zero-width no-break space (BOM)", 0xFEFF, 0xFEFF},
	{"Left-to-right mark", 0x200E, 0x200E},
	{"Right-to-left mark", 0x200F, 0x200F},
	{"Left-to-right embedding", 0x202A, 0x202A},
	{"Right-to-left embedding", 0x202B, 0x202B},
	{"Pop directional formatting", 0x202C, 0x202C},
	{"Left-to-right override", 0x202D, 0x202D},
	{"Right-to-left override", 0x202E, 0x202E},
	{"Soft hyphen", 0x00AD, 0x00AD},
	{"Invisible separator", 0x2063, 0x2063},
	{"Invisible times", 0x2062, 0x2062},
	{"Invisible plus", 0x2064, 0x2064},
	{"Function application", 0x2061, 0x2061},
	// Tag characters (used for invisible text encoding)
	{"Tag characters", 0xE0000, 0xE007F},
}

// claudeConfigPaths lists paths that should never be written to by an agent.
var claudeConfigPaths = []string{
	".claude/",
	"CLAUDE.md",
}

// DetectCriticalThreat checks for threats that bypass Rule of Two scoring.
// Returns nil if no critical threat is detected.
func DetectCriticalThreat(input *HookInput) *CriticalThreat {
	// Check Claude config writes first (always block, no ask mode)
	if threat := detectClaudeConfigWrite(input); threat != nil {
		return threat
	}

	// Check invisible Unicode (respects ask mode)
	if threat := detectInvisibleUnicodeThreat(input); threat != nil {
		return threat
	}

	return nil
}

// detectClaudeConfigWrite checks if the tool call attempts to write to Claude config.
func detectClaudeConfigWrite(input *HookInput) *CriticalThreat {
	var targetPath string

	switch input.ToolName {
	case "Write":
		parsed := ParseWriteInput(input.ToolInput)
		targetPath = parsed.FilePath
	case "Edit":
		parsed := ParseEditInput(input.ToolInput)
		targetPath = parsed.FilePath
	default:
		return nil
	}

	if targetPath == "" {
		return nil
	}

	// Normalize path for comparison
	normalizedPath := normalizePath(targetPath)

	for _, configPath := range claudeConfigPaths {
		if matchesClaudeConfigPath(normalizedPath, configPath) {
			return &CriticalThreat{
				Type:         "claude_config_write",
				Details:      fmt.Sprintf("Write to %s", targetPath),
				AllowAskMode: false, // Always block
			}
		}
	}

	return nil
}

// matchesClaudeConfigPath checks if a path matches a Claude config pattern.
func matchesClaudeConfigPath(path, pattern string) bool {
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		dir := strings.TrimSuffix(pattern, "/")
		// Check if path is in the directory
		return path == dir || strings.HasPrefix(path, pattern) ||
			strings.Contains(path, "/"+dir+"/") ||
			strings.HasSuffix(path, "/"+dir)
	}

	// Handle file patterns
	return path == pattern ||
		strings.HasSuffix(path, "/"+pattern) ||
		filepath.Base(path) == pattern
}

// normalizePath normalizes a file path for comparison.
func normalizePath(path string) string {
	// Clean the path
	path = filepath.Clean(path)
	// Remove leading ./ if present
	path = strings.TrimPrefix(path, "./")
	return path
}

// detectInvisibleUnicodeThreat checks for invisible Unicode in tool inputs.
func detectInvisibleUnicodeThreat(input *HookInput) *CriticalThreat {
	findings := detectInvisibleUnicode(input)
	if len(findings) == 0 {
		return nil
	}

	return &CriticalThreat{
		Type:         "invisible_unicode",
		Details:      formatInvisibleChars(findings),
		AllowAskMode: true, // Respect ask mode
	}
}

// detectInvisibleUnicode scans tool inputs for invisible Unicode characters.
func detectInvisibleUnicode(input *HookInput) []InvisibleCharInfo {
	var findings []InvisibleCharInfo

	switch input.ToolName {
	case "Write":
		parsed := ParseWriteInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.FilePath, "file_path")...)
		findings = append(findings, scanForInvisible(parsed.Content, "content")...)

	case "Edit":
		parsed := ParseEditInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.FilePath, "file_path")...)
		findings = append(findings, scanForInvisible(parsed.OldString, "old_string")...)
		findings = append(findings, scanForInvisible(parsed.NewString, "new_string")...)

	case "Bash":
		parsed := ParseBashInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.Command, "command")...)

	case "Read":
		parsed := ParseReadInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.FilePath, "file_path")...)

	case "Glob":
		parsed := ParseGlobInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.Pattern, "pattern")...)
		findings = append(findings, scanForInvisible(parsed.Path, "path")...)

	case "Grep":
		parsed := ParseGrepInput(input.ToolInput)
		findings = append(findings, scanForInvisible(parsed.Pattern, "pattern")...)
		findings = append(findings, scanForInvisible(parsed.Path, "path")...)
	}

	return findings
}

// scanForInvisible scans a string for invisible Unicode characters.
func scanForInvisible(s string, _ string) []InvisibleCharInfo {
	if s == "" {
		return nil
	}

	var found []InvisibleCharInfo
	runes := []rune(s)

	for i, r := range runes {
		if name := getInvisibleRuneName(r); name != "" {
			found = append(found, InvisibleCharInfo{
				Rune:     r,
				Name:     name,
				Position: i,
				Context:  getContext(runes, i),
			})
		}
	}

	return found
}

// getInvisibleRuneName returns the name of an invisible character, or empty string if not invisible.
func getInvisibleRuneName(r rune) string {
	for _, ir := range invisibleUnicodeRanges {
		if r >= ir.Start && r <= ir.End {
			return ir.Name
		}
	}

	// Also check for other control characters that shouldn't appear in code
	if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
		return fmt.Sprintf("Control character U+%04X", r)
	}

	return ""
}

// getContext returns surrounding characters for display.
func getContext(runes []rune, pos int) string {
	const contextLen = 5

	start := pos - contextLen
	if start < 0 {
		start = 0
	}
	end := pos + contextLen + 1
	if end > len(runes) {
		end = len(runes)
	}

	// Build context string, replacing the invisible char with a marker
	var result strings.Builder
	for i := start; i < end; i++ {
		if i == pos {
			result.WriteString("[HERE]")
		} else if name := getInvisibleRuneName(runes[i]); name != "" {
			result.WriteString("[?]")
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// formatInvisibleChars creates a human-readable description of invisible char findings.
func formatInvisibleChars(findings []InvisibleCharInfo) string {
	if len(findings) == 0 {
		return ""
	}

	if len(findings) == 1 {
		f := findings[0]
		return fmt.Sprintf("%s (U+%04X) at position %d: ...%s...",
			f.Name, f.Rune, f.Position, f.Context)
	}

	// Group by type
	typeCount := make(map[string]int)
	for _, f := range findings {
		typeCount[f.Name]++
	}

	var parts []string
	for name, count := range typeCount {
		parts = append(parts, fmt.Sprintf("%s (x%d)", name, count))
	}

	return fmt.Sprintf("%d invisible characters: %s", len(findings), strings.Join(parts, ", "))
}

// GenerateThreatOutput creates the appropriate hook output for a critical threat.
// Returns (output, exitCode, stderrMessage).
func GenerateThreatOutput(threat *CriticalThreat) (*HookOutput, int, string) {
	mode := GetAgenticMode()

	switch threat.Type {
	case "claude_config_write":
		// Always block, never ask
		stderrMsg := fmt.Sprintf(
			"Blocked: Attempted write to Claude agent configuration. %s",
			threat.Details)
		return nil, 2, stderrMsg

	case "invisible_unicode":
		if mode == ModeAsk && threat.AllowAskMode {
			// Ask mode - prompt user
			return &HookOutput{
				HookSpecificOutput: &HookSpecificOutput{
					HookEventName:      "PreToolUse",
					PermissionDecision: "ask",
					PermissionDecisionReason: fmt.Sprintf(
						"Invisible Unicode detected: %s", threat.Details),
				},
				SystemMessage: fmt.Sprintf(
					"Invisible Unicode characters detected in tool input. "+
						"These may indicate a prompt injection attack. Details: %s",
					threat.Details),
			}, 0, ""
		}

		// Block mode (default)
		stderrMsg := fmt.Sprintf(
			"Blocked: Invisible Unicode detected in tool input. %s",
			threat.Details)
		return nil, 2, stderrMsg

	default:
		// Unknown threat type - block to be safe
		stderrMsg := fmt.Sprintf("Blocked: Unknown critical threat: %s", threat.Type)
		return nil, 2, stderrMsg
	}
}
