package agentic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CriticalThreat represents a security threat that bypasses Rule of Two scoring.
// These are threats that warrant immediate blocking regardless of capability count.
type CriticalThreat struct {
	Type    string // "agent_config_write", "invisible_unicode"
	Details string
	// AllowAskMode indicates whether DASHLIGHTS_AGENTIC_MODE=ask should prompt
	// instead of blocking. Agent config writes always block (false).
	AllowAskMode bool
}

// InvisibleCharInfo describes a detected invisible Unicode character.
type InvisibleCharInfo struct {
	Rune     rune
	Name     string
	Position int
	Context  string // surrounding characters for display
	Field    string // which input field contained this character
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

// agentConfigPaths lists paths that should never be written to by any agent.
// These are configuration files that could hijack agent behavior.
var agentConfigPaths = []string{
	// Claude Code config
	".claude/settings.json",
	".claude/settings.local.json",
	".claude/commands/", // Custom slash commands
	"CLAUDE.md",

	// Cursor config (project-level)
	".cursor/hooks.json",
	".cursor/rules",
}

// agentConfigHomePaths are config files relative to user home directory.
// These are matched against absolute paths after expanding ~.
var agentConfigHomePaths = []string{
	".cursor/cli-config.json",
	".cursor/hooks.json",
}

// agentConfigSafeSubdirs are subdirectories within .claude/ that are safe to write.
// These are working directories, not configuration files.
var agentConfigSafeSubdirs = []string{
	".claude/plans/",
	".claude/todos/",
}

// DetectCriticalThreat checks for threats that bypass Rule of Two scoring.
// Returns nil if no critical threat is detected.
func DetectCriticalThreat(input *HookInput) *CriticalThreat {
	// Check agent config writes first (always block, no ask mode)
	if threat := detectAgentConfigWrite(input); threat != nil {
		return threat
	}

	// Check invisible Unicode (respects ask mode)
	if threat := detectInvisibleUnicodeThreat(input); threat != nil {
		return threat
	}

	return nil
}

// detectAgentConfigWrite checks if the tool call attempts to write to agent config.
func detectAgentConfigWrite(input *HookInput) *CriticalThreat {
	var targetPaths []string

	switch input.ToolName {
	case "Write":
		parsed := ParseWriteInput(input.ToolInput)
		if parsed.FilePath != "" {
			targetPaths = append(targetPaths, parsed.FilePath)
		}
	case "Edit":
		parsed := ParseEditInput(input.ToolInput)
		if parsed.FilePath != "" {
			targetPaths = append(targetPaths, parsed.FilePath)
		}
	case "Bash":
		parsed := ParseBashInput(input.ToolInput)
		targetPaths = append(targetPaths, extractBashWriteTargets(parsed.Command)...)
	default:
		return nil
	}

	if len(targetPaths) == 0 {
		return nil
	}

	for _, targetPath := range targetPaths {
		if targetPath == "" {
			continue
		}
		// Normalize path for comparison
		normalizedPath := normalizePath(cleanBashPathToken(targetPath))

		// Check if path is in a safe subdirectory first
		if isInSafeSubdir(normalizedPath) {
			continue
		}

		// Check project-level config paths
		for _, configPath := range agentConfigPaths {
			if matchesAgentConfigPath(normalizedPath, configPath) {
				return &CriticalThreat{
					Type:         "agent_config_write",
					Details:      fmt.Sprintf("Write to %s", targetPath),
					AllowAskMode: false, // Always block
				}
			}
		}

		// Check home directory config paths
		if matchesHomeConfigPath(normalizedPath) {
			return &CriticalThreat{
				Type:         "agent_config_write",
				Details:      fmt.Sprintf("Write to %s", targetPath),
				AllowAskMode: false, // Always block
			}
		}
	}

	return nil
}

// isInSafeSubdir checks if a path is within a safe subdirectory.
func isInSafeSubdir(path string) bool {
	for _, safeDir := range agentConfigSafeSubdirs {
		dir := strings.TrimSuffix(safeDir, "/")
		// Check if path is in the safe directory
		if strings.HasPrefix(path, safeDir) ||
			strings.Contains(path, "/"+safeDir) ||
			strings.Contains(path, "/"+dir+"/") {
			return true
		}
	}
	return false
}

// matchesHomeConfigPath checks if an absolute path matches a home directory config.
func matchesHomeConfigPath(path string) bool {
	// Only check absolute paths
	if !filepath.IsAbs(path) {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	for _, configPath := range agentConfigHomePaths {
		fullPath := filepath.Join(homeDir, configPath)
		if path == fullPath || path == filepath.Clean(fullPath) {
			return true
		}
	}
	return false
}

// extractBashWriteTargets pulls likely file write targets from a Bash command.
// This is a heuristic that looks for redirects and tee targets.
func extractBashWriteTargets(command string) []string {
	if command == "" {
		return nil
	}

	tokens := tokenizeBashCommand(command)
	if len(tokens) == 0 {
		return nil
	}

	var targets []string

	if inPlaceTargets := extractInPlaceEditorTargets(tokens); len(inPlaceTargets) > 0 {
		targets = append(targets, inPlaceTargets...)
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if target := extractRedirectionTarget(tok, tokens, i); target != "" {
			targets = append(targets, target)
		}

		if isTeeCommand(tok) {
			teeTargets := extractTeeTargets(tokens[i+1:])
			if len(teeTargets) > 0 {
				targets = append(targets, teeTargets...)
			}
		}
	}

	return targets
}

func extractInPlaceEditorTargets(tokens []string) []string {
	if len(tokens) == 0 {
		return nil
	}

	cmd := filepath.Base(cleanBashPathToken(tokens[0]))
	switch cmd {
	case "sed", "gsed":
		return extractSedInPlaceTargets(tokens[1:])
	case "perl", "ruby":
		return extractPerlRubyInPlaceTargets(tokens[1:])
	default:
		return nil
	}
}

func extractSedInPlaceTargets(tokens []string) []string {
	var operands []string
	inPlace := false
	hasScriptOption := false

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "|" || tok == "||" || tok == "&&" || tok == ";" {
			break
		}
		if strings.HasPrefix(tok, "-") {
			if tok == "--" {
				operands = append(operands, tokens[i+1:]...)
				break
			}
			if strings.HasPrefix(tok, "-i") {
				inPlace = true
				if tok == "-i" && i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
					i++
				}
				continue
			}
			if tok == "-e" || tok == "-f" {
				hasScriptOption = true
				if i+1 < len(tokens) {
					i++
				}
				continue
			}
			continue
		}
		operands = append(operands, tok)
	}

	if !inPlace {
		return nil
	}

	if !hasScriptOption {
		if len(operands) <= 1 {
			return nil
		}
		operands = operands[1:]
	}

	return cleanTargets(operands)
}

func extractPerlRubyInPlaceTargets(tokens []string) []string {
	var operands []string
	inPlace := false
	hasScriptOption := false

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "|" || tok == "||" || tok == "&&" || tok == ";" {
			break
		}
		if strings.HasPrefix(tok, "-") {
			if tok == "--" {
				operands = append(operands, tokens[i+1:]...)
				break
			}
			if strings.Contains(tok, "i") {
				inPlace = true
			}
			if tok == "-e" {
				hasScriptOption = true
				if i+1 < len(tokens) {
					i++
				}
			}
			continue
		}
		operands = append(operands, tok)
	}

	if !inPlace {
		return nil
	}

	if !hasScriptOption {
		if len(operands) <= 1 {
			return nil
		}
		operands = operands[1:]
	}

	return cleanTargets(operands)
}

func cleanTargets(tokens []string) []string {
	var targets []string
	for _, tok := range tokens {
		target := cleanBashPathToken(tok)
		if target != "" {
			targets = append(targets, target)
		}
	}
	return targets
}

func extractRedirectionTarget(tok string, tokens []string, idx int) string {
	if tok == "" {
		return ""
	}

	// Exact operator tokens.
	if isRedirectionOperator(tok) {
		if idx+1 >= len(tokens) {
			return ""
		}
		next := cleanBashPathToken(tokens[idx+1])
		if strings.HasPrefix(next, "&") {
			return ""
		}
		return next
	}

	// Operator with attached path (e.g., >file, 2>/tmp/out).
	for _, prefix := range redirectionPrefixes() {
		if strings.HasPrefix(tok, prefix) && len(tok) > len(prefix) {
			return cleanBashPathToken(tok[len(prefix):])
		}
	}

	return ""
}

func redirectionPrefixes() []string {
	return []string{"&>>", "&>", "2>>", "2>", "1>>", "1>", ">>", ">"}
}

func isRedirectionOperator(tok string) bool {
	switch tok {
	case ">", ">>", "1>", "1>>", "2>", "2>>", "&>", "&>>":
		return true
	default:
		return false
	}
}

func isTeeCommand(tok string) bool {
	if tok == "" {
		return false
	}
	return filepath.Base(tok) == "tee"
}

func extractTeeTargets(tokens []string) []string {
	var targets []string

	for _, tok := range tokens {
		if tok == "|" || tok == "||" || tok == "&&" || tok == ";" {
			break
		}
		if strings.HasPrefix(tok, "-") {
			continue
		}
		target := cleanBashPathToken(tok)
		if target != "" {
			targets = append(targets, target)
		}
	}

	return targets
}

// tokenizeBashCommand is a lightweight tokenizer that respects quotes and pipes.
func tokenizeBashCommand(command string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		if r == '\'' && !inDouble {
			inSingle = !inSingle
			current.WriteRune(r)
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			current.WriteRune(r)
			continue
		}

		if !inSingle && !inDouble {
			if r == '|' || r == ';' {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				tokens = append(tokens, string(r))
				continue
			}
			if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// cleanBashPathToken trims quotes and common shell separators from a token.
func cleanBashPathToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}

	if len(token) >= 2 {
		if token[0] == '\'' && token[len(token)-1] == '\'' {
			token = token[1 : len(token)-1]
		} else if token[0] == '"' && token[len(token)-1] == '"' {
			token = token[1 : len(token)-1]
		}
	}

	token = strings.TrimRight(token, ";|&")
	return token
}

// matchesAgentConfigPath checks if a path matches an agent config pattern.
func matchesAgentConfigPath(path, pattern string) bool {
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
func scanForInvisible(s string, fieldName string) []InvisibleCharInfo {
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
				Field:    fieldName,
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
	case "agent_config_write":
		// Always block, never ask
		stderrMsg := fmt.Sprintf(
			"Blocked: Attempted write to agent configuration. %s",
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
