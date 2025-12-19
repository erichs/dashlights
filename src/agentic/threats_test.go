package agentic

import (
	"os"
	"testing"
)

func TestDetectAgentConfigWrite(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		toolInput  map[string]interface{}
		wantThreat bool
	}{
		// Claude Code config paths
		{
			name:     "Write to .claude/settings.json",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": ".claude/settings.json",
				"content":   "{}",
			},
			wantThreat: true,
		},
		{
			name:     "Write to CLAUDE.md",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "CLAUDE.md",
				"content":   "# Malicious instructions",
			},
			wantThreat: true,
		},
		{
			name:     "Write to absolute path CLAUDE.md",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "/Users/test/project/CLAUDE.md",
				"content":   "# Malicious",
			},
			wantThreat: true,
		},
		{
			name:     "Edit to .claude/commands/custom.md",
			toolName: "Edit",
			toolInput: map[string]interface{}{
				"file_path":  ".claude/commands/custom.md",
				"old_string": "old",
				"new_string": "new",
			},
			wantThreat: true,
		},
		// Cursor config paths
		{
			name:     "Write to .cursor/hooks.json",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": ".cursor/hooks.json",
				"content":   "{}",
			},
			wantThreat: true,
		},
		{
			name:     "Write to .cursor/rules",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": ".cursor/rules",
				"content":   "malicious rules",
			},
			wantThreat: true,
		},
		{
			name:     "Edit .cursor/hooks.json in project",
			toolName: "Edit",
			toolInput: map[string]interface{}{
				"file_path":  "/Users/test/project/.cursor/hooks.json",
				"old_string": "old",
				"new_string": "new",
			},
			wantThreat: true,
		},
		// Safe subdirectories (should NOT trigger)
		{
			name:     "Write to .claude/plans/ - safe subdir",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": ".claude/plans/my-plan.md",
				"content":   "# Plan",
			},
			wantThreat: false,
		},
		{
			name:     "Write to absolute .claude/plans/ - safe subdir",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "/Users/test/.claude/plans/plan.md",
				"content":   "# Plan",
			},
			wantThreat: false,
		},
		{
			name:     "Write to .claude/todos/ - safe subdir",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": ".claude/todos/todo.json",
				"content":   "{}",
			},
			wantThreat: false,
		},
		// Normal files (should NOT trigger)
		{
			name:     "Write to normal file",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "src/main.go",
				"content":   "package main",
			},
			wantThreat: false,
		},
		{
			name:     "Write to file containing claude in name",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "docs/using-claude.md",
				"content":   "# How to use Claude",
			},
			wantThreat: false,
		},
		{
			name:     "Bash command - not a write",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "cat CLAUDE.md",
			},
			wantThreat: false,
		},
		{
			name:     "Bash redirect write to .claude",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo test > .claude/settings.json",
			},
			wantThreat: true,
		},
		{
			name:     "Bash tee write to CLAUDE.md",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "printf 'x' | tee CLAUDE.md",
			},
			wantThreat: true,
		},
		{
			name:     "Bash redirect write with quotes",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo test > \".claude/settings.json\"",
			},
			wantThreat: true,
		},
		{
			name:     "Bash redirect to non-config path",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo test > ./tmp/output.txt",
			},
			wantThreat: false,
		},
		{
			name:     "Bash sed -i write to cursor hooks",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "sed -i '' 's/old/new/' ~/.cursor/hooks.json",
			},
			wantThreat: true,
		},
		{
			name:     "Bash perl -pi write to cursor hooks",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "perl -pi -e 's/old/new/' ~/.cursor/hooks.json",
			},
			wantThreat: true,
		},
		{
			name:     "Bash redirect to .claude/plans/ - safe",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo test > .claude/plans/output.md",
			},
			wantThreat: false,
		},
		{
			name:     "Read - not a write",
			toolName: "Read",
			toolInput: map[string]interface{}{
				"file_path": ".claude/settings.json",
			},
			wantThreat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &HookInput{
				ToolName:  tt.toolName,
				ToolInput: tt.toolInput,
			}

			threat := detectAgentConfigWrite(input)

			if tt.wantThreat && threat == nil {
				t.Error("Expected threat to be detected, got nil")
			}
			if !tt.wantThreat && threat != nil {
				t.Errorf("Expected no threat, got: %+v", threat)
			}
			if threat != nil && threat.Type != "agent_config_write" {
				t.Errorf("Expected type 'agent_config_write', got '%s'", threat.Type)
			}
			if threat != nil && threat.AllowAskMode {
				t.Error("Agent config writes should never allow ask mode")
			}
		})
	}
}

func TestDetectInvisibleUnicode(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		toolInput  map[string]interface{}
		wantCount  int
		wantThreat bool
	}{
		{
			name:     "Zero-width space in content",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test.txt",
				"content":   "Hello\u200BWorld", // Zero-width space between words
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "Multiple invisible chars",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test.txt",
				"content":   "\u200B\u200C\u200D", // ZWS, ZWNJ, ZWJ
			},
			wantCount:  3,
			wantThreat: true,
		},
		{
			name:     "Right-to-left override in bash",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "cat file\u202E.txt", // RLO can spoof filenames
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "Invisible char in file path",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test\u200B.txt", // ZWS in filename
				"content":   "normal content",
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "BOM in content",
			toolName: "Edit",
			toolInput: map[string]interface{}{
				"file_path":  "test.txt",
				"old_string": "old",
				"new_string": "\uFEFFnew", // BOM prefix
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "Normal content - no invisible chars",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test.txt",
				"content":   "Hello, World! This is normal text.",
			},
			wantCount:  0,
			wantThreat: false,
		},
		{
			name:     "Emoji - not invisible",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test.txt",
				"content":   "Hello 🎉 World",
			},
			wantCount:  0,
			wantThreat: false,
		},
		{
			name:     "Newlines and tabs - allowed control chars",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "test.txt",
				"content":   "Line1\nLine2\tTabbed",
			},
			wantCount:  0,
			wantThreat: false,
		},
		{
			name:     "Invisible in Grep pattern",
			toolName: "Grep",
			toolInput: map[string]interface{}{
				"pattern": "search\u200Bterm",
				"path":    ".",
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "Invisible in Glob pattern",
			toolName: "Glob",
			toolInput: map[string]interface{}{
				"pattern": "*.txt\u200B",
			},
			wantCount:  1,
			wantThreat: true,
		},
		{
			name:     "Tag character - used for invisible encoding",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo \U000E0041hello", // Tag Latin Capital Letter A
			},
			wantCount:  1,
			wantThreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &HookInput{
				ToolName:  tt.toolName,
				ToolInput: tt.toolInput,
			}

			findings := detectInvisibleUnicode(input)

			if len(findings) != tt.wantCount {
				t.Errorf("Expected %d invisible chars, found %d", tt.wantCount, len(findings))
				for _, f := range findings {
					t.Logf("  Found: %s (U+%04X) at pos %d", f.Name, f.Rune, f.Position)
				}
			}

			threat := detectInvisibleUnicodeThreat(input)
			if tt.wantThreat && threat == nil {
				t.Error("Expected threat to be detected, got nil")
			}
			if !tt.wantThreat && threat != nil {
				t.Errorf("Expected no threat, got: %+v", threat)
			}
			if threat != nil && threat.Type != "invisible_unicode" {
				t.Errorf("Expected type 'invisible_unicode', got '%s'", threat.Type)
			}
			if threat != nil && !threat.AllowAskMode {
				t.Error("Invisible unicode threats should allow ask mode")
			}
		})
	}
}

func TestDetectCriticalThreat(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		toolInput   map[string]interface{}
		wantThreat  bool
		wantType    string
		wantAskMode bool
	}{
		{
			name:     "Agent config takes priority",
			toolName: "Write",
			toolInput: map[string]interface{}{
				"file_path": "CLAUDE.md",
				"content":   "content\u200B", // Has invisible char too
			},
			wantThreat:  true,
			wantType:    "agent_config_write",
			wantAskMode: false,
		},
		{
			name:     "Invisible unicode when no config write",
			toolName: "Bash",
			toolInput: map[string]interface{}{
				"command": "echo \u200B",
			},
			wantThreat:  true,
			wantType:    "invisible_unicode",
			wantAskMode: true,
		},
		{
			name:     "Safe input - no threat",
			toolName: "Read",
			toolInput: map[string]interface{}{
				"file_path": "/tmp/test.txt",
			},
			wantThreat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &HookInput{
				ToolName:  tt.toolName,
				ToolInput: tt.toolInput,
			}

			threat := DetectCriticalThreat(input)

			if tt.wantThreat && threat == nil {
				t.Error("Expected threat to be detected, got nil")
			}
			if !tt.wantThreat && threat != nil {
				t.Errorf("Expected no threat, got: %+v", threat)
			}
			if threat != nil {
				if threat.Type != tt.wantType {
					t.Errorf("Expected type '%s', got '%s'", tt.wantType, threat.Type)
				}
				if threat.AllowAskMode != tt.wantAskMode {
					t.Errorf("Expected AllowAskMode=%v, got %v", tt.wantAskMode, threat.AllowAskMode)
				}
			}
		})
	}
}

func TestGenerateThreatOutput_AgentConfig(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	threat := &CriticalThreat{
		Type:         "agent_config_write",
		Details:      "Write to CLAUDE.md",
		AllowAskMode: false,
	}

	// Test block mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")
	output, exitCode, stderrMsg := GenerateThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if output != nil {
		t.Error("Expected nil output for blocked threat")
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr message")
	}

	// Test ask mode - should STILL block for agent config
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")
	output, exitCode, stderrMsg = GenerateThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 even in ask mode, got %d", exitCode)
	}
	if output != nil {
		t.Error("Expected nil output - agent config should always block")
	}
}

func TestGenerateThreatOutput_InvisibleUnicode(t *testing.T) {
	// Save original value
	original := os.Getenv("DASHLIGHTS_AGENTIC_MODE")
	defer os.Setenv("DASHLIGHTS_AGENTIC_MODE", original)

	threat := &CriticalThreat{
		Type:         "invisible_unicode",
		Details:      "Zero-width space detected",
		AllowAskMode: true,
	}

	// Test block mode
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "block")
	output, exitCode, stderrMsg := GenerateThreatOutput(threat)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 in block mode, got %d", exitCode)
	}
	if output != nil {
		t.Error("Expected nil output for blocked threat")
	}
	if stderrMsg == "" {
		t.Error("Expected non-empty stderr message")
	}

	// Test ask mode - should prompt user
	os.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")
	output, exitCode, stderrMsg = GenerateThreatOutput(threat)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 in ask mode, got %d", exitCode)
	}
	if output == nil {
		t.Fatal("Expected non-nil output in ask mode")
	}
	if output.HookSpecificOutput.PermissionDecision != "ask" {
		t.Errorf("Expected 'ask' decision, got '%s'", output.HookSpecificOutput.PermissionDecision)
	}
	if output.SystemMessage == "" {
		t.Error("Expected non-empty system message")
	}
}

func TestMatchesAgentConfigPath(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		// .claude/settings.json file pattern
		{".claude/settings.json", ".claude/settings.json", true},
		{"path/to/.claude/settings.json", ".claude/settings.json", true},
		{"/Users/test/project/.claude/settings.json", ".claude/settings.json", true},

		// .claude/commands/ directory pattern
		{".claude/commands/foo.md", ".claude/commands/", true},
		{"path/to/.claude/commands/custom.md", ".claude/commands/", true},

		// CLAUDE.md file pattern
		{"CLAUDE.md", "CLAUDE.md", true},
		{"/Users/test/project/CLAUDE.md", "CLAUDE.md", true},
		{"CLAUDE.md", "CLAUDE.md", true},

		// Cursor config patterns
		{".cursor/hooks.json", ".cursor/hooks.json", true},
		{"/Users/test/project/.cursor/hooks.json", ".cursor/hooks.json", true},
		{".cursor/rules", ".cursor/rules", true},

		// Should NOT match
		{"claude.md", "CLAUDE.md", false},                             // case sensitive
		{"src/claudeutils.go", ".claude/settings.json", false},        // not settings.json
		{"docs/using-claude.md", "CLAUDE.md", false},                  // not CLAUDE.md
		{".claude/plans/plan.md", ".claude/settings.json", false},     // different file
		{".claude/settings.json.bak", ".claude/settings.json", false}, // different file
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.pattern, func(t *testing.T) {
			normalized := normalizePath(tt.path)
			got := matchesAgentConfigPath(normalized, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesAgentConfigPath(%q, %q) = %v, want %v",
					normalized, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestIsInSafeSubdir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Safe subdirectories
		{".claude/plans/my-plan.md", true},
		{".claude/plans/subdir/plan.md", true},
		{".claude/todos/todo.json", true},
		{"/Users/test/.claude/plans/plan.md", true},
		{"path/to/.claude/plans/file.md", true},

		// Not safe
		{".claude/settings.json", false},
		{".claude/commands/cmd.md", false},
		{"CLAUDE.md", false},
		{".cursor/hooks.json", false},
		{"src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isInSafeSubdir(tt.path)
			if got != tt.want {
				t.Errorf("isInSafeSubdir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchesHomeConfigPath(t *testing.T) {
	// This test is environment-dependent, so we test the logic
	// with known paths

	// Non-absolute paths should always return false
	if matchesHomeConfigPath(".cursor/hooks.json") {
		t.Error("Expected false for relative path")
	}
	if matchesHomeConfigPath("cursor/cli-config.json") {
		t.Error("Expected false for relative path")
	}

	// Random absolute paths should not match
	if matchesHomeConfigPath("/tmp/hooks.json") {
		t.Error("Expected false for /tmp path")
	}
	if matchesHomeConfigPath("/var/log/test.json") {
		t.Error("Expected false for /var/log path")
	}
}

func TestGetInvisibleRuneName(t *testing.T) {
	tests := []struct {
		r       rune
		wantNon string // just check if non-empty when expected
	}{
		{0x200B, "Zero-width space"},
		{0x200C, "Zero-width non-joiner"},
		{0x200D, "Zero-width joiner"},
		{0x202E, "Right-to-left override"},
		{0xFEFF, "Zero-width no-break space (BOM)"},
		{0x00AD, "Soft hyphen"},
		{0xE0041, "Tag characters"},
		{'A', ""},  // Normal ASCII
		{'\n', ""}, // Allowed control char
		{'\t', ""}, // Allowed control char
		{'🎉', ""},  // Emoji - not invisible
	}

	for _, tt := range tests {
		name := getInvisibleRuneName(tt.r)
		if tt.wantNon != "" && name == "" {
			t.Errorf("Expected non-empty name for U+%04X, got empty", tt.r)
		}
		if tt.wantNon == "" && name != "" {
			t.Errorf("Expected empty name for U+%04X, got '%s'", tt.r, name)
		}
	}
}

func TestFormatInvisibleChars(t *testing.T) {
	// Empty findings
	result := formatInvisibleChars(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil findings, got '%s'", result)
	}

	// Single finding
	single := []InvisibleCharInfo{
		{Rune: 0x200B, Name: "Zero-width space", Position: 5, Context: "Hello[HERE]World"},
	}
	result = formatInvisibleChars(single)
	if result == "" {
		t.Error("Expected non-empty result for single finding")
	}

	// Multiple findings
	multiple := []InvisibleCharInfo{
		{Rune: 0x200B, Name: "Zero-width space", Position: 0},
		{Rune: 0x200B, Name: "Zero-width space", Position: 5},
		{Rune: 0x200C, Name: "Zero-width non-joiner", Position: 10},
	}
	result = formatInvisibleChars(multiple)
	if result == "" {
		t.Error("Expected non-empty result for multiple findings")
	}
}

func TestGetContext(t *testing.T) {
	runes := []rune("Hello World")
	ctx := getContext(runes, 5) // Space between Hello and World

	if ctx == "" {
		t.Error("Expected non-empty context")
	}
	if len(ctx) > 20 { // contextLen is 5, so max ~11 chars + marker
		t.Errorf("Context too long: %q", ctx)
	}
}

func TestExtractBashWriteTargets(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "Simple redirect",
			command: "echo hi > .claude/settings.json",
			want:    []string{".claude/settings.json"},
		},
		{
			name:    "Append redirect with fd",
			command: "echo hi 1>>CLAUDE.md",
			want:    []string{"CLAUDE.md"},
		},
		{
			name:    "Redirect with combined fd",
			command: "echo hi &>>.claude/settings.json",
			want:    []string{".claude/settings.json"},
		},
		{
			name:    "Redirect with attached path",
			command: "echo hi >/tmp/out.txt",
			want:    []string{"/tmp/out.txt"},
		},
		{
			name:    "Tee command",
			command: "echo hi | tee ./CLAUDE.md",
			want:    []string{"./CLAUDE.md"},
		},
		{
			name:    "Tee with absolute path",
			command: "echo hi | /usr/bin/tee .claude/settings.json",
			want:    []string{".claude/settings.json"},
		},
		{
			name:    "Tee with options",
			command: "echo hi | tee -a .claude/settings.json",
			want:    []string{".claude/settings.json"},
		},
		{
			name:    "No write targets",
			command: "cat CLAUDE.md | wc -l",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBashWriteTargets(tt.command)
			if len(got) != len(tt.want) {
				t.Fatalf("Expected %d targets, got %d: %v", len(tt.want), len(got), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Target %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTokenizeBashCommand(t *testing.T) {
	command := "echo 'hello world' | tee \"file name.txt\" > out.txt"
	tokens := tokenizeBashCommand(command)
	if len(tokens) == 0 {
		t.Fatal("Expected tokens, got none")
	}
	if tokens[2] != "|" {
		t.Errorf("Expected pipe token at index 2, got %q", tokens[2])
	}

	command = "echo hi\\;there; echo done"
	tokens = tokenizeBashCommand(command)
	if len(tokens) < 3 {
		t.Fatalf("Expected more tokens, got %v", tokens)
	}
	if tokens[2] != ";" {
		t.Errorf("Expected semicolon token at index 2, got %q", tokens[2])
	}
}

func TestCleanBashPathToken(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"\"file.txt\"", "file.txt"},
		{"'file.txt'", "file.txt"},
		{"file.txt;", "file.txt"},
		{"file.txt|", "file.txt"},
		{" file.txt ", "file.txt"},
	}

	for _, tt := range tests {
		got := cleanBashPathToken(tt.token)
		if got != tt.want {
			t.Errorf("cleanBashPathToken(%q) = %q, want %q", tt.token, got, tt.want)
		}
	}
}

func TestExtractRedirectionTarget(t *testing.T) {
	tokens := []string{"echo", "hi", ">", "out.txt"}
	target := extractRedirectionTarget(tokens[2], tokens, 2)
	if target != "out.txt" {
		t.Errorf("Expected out.txt, got %q", target)
	}

	tokens = []string{"echo", "hi", "2>/tmp/err.txt"}
	target = extractRedirectionTarget(tokens[2], tokens, 2)
	if target != "/tmp/err.txt" {
		t.Errorf("Expected /tmp/err.txt, got %q", target)
	}

	tokens = []string{"echo", "hi", ">", "&1"}
	target = extractRedirectionTarget(tokens[2], tokens, 2)
	if target != "" {
		t.Errorf("Expected empty target for fd redirect, got %q", target)
	}

	target = extractRedirectionTarget("", tokens, 0)
	if target != "" {
		t.Errorf("Expected empty target for empty token, got %q", target)
	}
}

func TestTeeHelpers(t *testing.T) {
	if !isTeeCommand("tee") {
		t.Error("Expected tee to be recognized")
	}
	if !isTeeCommand("/usr/bin/tee") {
		t.Error("Expected /usr/bin/tee to be recognized")
	}
	if isTeeCommand("nottee") {
		t.Error("Expected nottee to be ignored")
	}

	tokens := []string{"-a", "out.txt", "|", "wc"}
	targets := extractTeeTargets(tokens)
	if len(targets) != 1 || targets[0] != "out.txt" {
		t.Errorf("Unexpected tee targets: %v", targets)
	}

	if !isRedirectionOperator(">>") {
		t.Error("Expected >> to be recognized as redirection")
	}
	if isRedirectionOperator("<") {
		t.Error("Expected < to be ignored as redirection")
	}
}
