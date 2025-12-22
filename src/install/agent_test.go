package install

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseAgentType_Claude tests parsing "claude" agent type.
func TestParseAgentType_Claude(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase", "claude"},
		{"uppercase", "CLAUDE"},
		{"mixed case", "Claude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgentType(tt.input)
			if err != nil {
				t.Fatalf("ParseAgentType(%q) error = %v, want nil", tt.input, err)
			}
			if got != AgentClaude {
				t.Errorf("ParseAgentType(%q) = %q, want %q", tt.input, got, AgentClaude)
			}
		})
	}
}

// TestParseAgentType_Cursor tests parsing "cursor" agent type.
func TestParseAgentType_Cursor(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase", "cursor"},
		{"uppercase", "CURSOR"},
		{"mixed case", "Cursor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgentType(tt.input)
			if err != nil {
				t.Fatalf("ParseAgentType(%q) error = %v, want nil", tt.input, err)
			}
			if got != AgentCursor {
				t.Errorf("ParseAgentType(%q) = %q, want %q", tt.input, got, AgentCursor)
			}
		})
	}
}

// TestParseAgentType_Invalid tests parsing invalid agent types.
func TestParseAgentType_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"unknown agent", "vscode"},
		{"invalid name", "invalid"},
		{"numeric", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgentType(tt.input)
			if err == nil {
				t.Fatalf("ParseAgentType(%q) = %q, want error", tt.input, got)
			}
			if got != "" {
				t.Errorf("ParseAgentType(%q) = %q, want empty string on error", tt.input, got)
			}
			if !strings.Contains(err.Error(), "unsupported agent") {
				t.Errorf("ParseAgentType(%q) error = %q, want error containing 'unsupported agent'", tt.input, err.Error())
			}
		})
	}
}

// TestAgentInstaller_GetAgentConfig_Claude tests Claude agent config.
func TestAgentInstaller_GetAgentConfig_Claude(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	config, err := installer.GetAgentConfig(AgentClaude)
	if err != nil {
		t.Fatalf("GetAgentConfig(AgentClaude) error = %v, want nil", err)
	}

	if config.Type != AgentClaude {
		t.Errorf("config.Type = %q, want %q", config.Type, AgentClaude)
	}

	expectedPath := filepath.Join("/home/testuser", ".claude", "settings.json")
	if config.ConfigPath != expectedPath {
		t.Errorf("config.ConfigPath = %q, want %q", config.ConfigPath, expectedPath)
	}

	if config.Name != "Claude Code" {
		t.Errorf("config.Name = %q, want %q", config.Name, "Claude Code")
	}
}

// TestAgentInstaller_GetAgentConfig_Cursor tests Cursor agent config.
func TestAgentInstaller_GetAgentConfig_Cursor(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	config, err := installer.GetAgentConfig(AgentCursor)
	if err != nil {
		t.Fatalf("GetAgentConfig(AgentCursor) error = %v, want nil", err)
	}

	if config.Type != AgentCursor {
		t.Errorf("config.Type = %q, want %q", config.Type, AgentCursor)
	}

	expectedPath := filepath.Join("/home/testuser", ".cursor", "hooks.json")
	if config.ConfigPath != expectedPath {
		t.Errorf("config.ConfigPath = %q, want %q", config.ConfigPath, expectedPath)
	}

	if config.Name != "Cursor" {
		t.Errorf("config.Name = %q, want %q", config.Name, "Cursor")
	}
}

// TestAgentInstaller_GetAgentConfig_Invalid tests invalid agent type.
func TestAgentInstaller_GetAgentConfig_Invalid(t *testing.T) {
	fs := NewMockFilesystem()
	installer := NewAgentInstaller(fs)

	config, err := installer.GetAgentConfig(AgentType("invalid"))
	if err == nil {
		t.Fatalf("GetAgentConfig(invalid) = %v, want error", config)
	}

	if config != nil {
		t.Errorf("GetAgentConfig(invalid) config = %v, want nil", config)
	}

	if !strings.Contains(err.Error(), "unsupported agent type") {
		t.Errorf("error = %q, want error containing 'unsupported agent type'", err.Error())
	}
}

// TestAgentInstaller_GetAgentConfig_HomeDirError tests error from UserHomeDir.
func TestAgentInstaller_GetAgentConfig_HomeDirError(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDirErr = fmt.Errorf("home dir error")
	installer := NewAgentInstaller(fs)

	config, err := installer.GetAgentConfig(AgentClaude)
	if err == nil {
		t.Fatalf("GetAgentConfig() = %v, want error", config)
	}

	if config != nil {
		t.Errorf("GetAgentConfig() config = %v, want nil", config)
	}

	if !strings.Contains(err.Error(), "could not determine home directory") {
		t.Errorf("error = %q, want error containing 'could not determine home directory'", err.Error())
	}
}

// TestAgentInstaller_IsInstalled_True tests when dashlights is installed.
func TestAgentInstaller_IsInstalled_True(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	fs.Files[configPath] = []byte(`{"hooks": {"PreToolUse": [{"hooks": [{"command": "dashlights --agentic"}]}]}}`)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	installed, err := installer.IsInstalled(config)
	if err != nil {
		t.Fatalf("IsInstalled() error = %v, want nil", err)
	}

	if !installed {
		t.Errorf("IsInstalled() = false, want true")
	}
}

// TestAgentInstaller_IsInstalled_False tests when dashlights is not installed.
func TestAgentInstaller_IsInstalled_False(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	fs.Files[configPath] = []byte(`{"hooks": {}}`)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	installed, err := installer.IsInstalled(config)
	if err != nil {
		t.Fatalf("IsInstalled() error = %v, want nil", err)
	}

	if installed {
		t.Errorf("IsInstalled() = true, want false")
	}
}

// TestAgentInstaller_IsInstalled_FileNotExist tests when config file doesn't exist.
func TestAgentInstaller_IsInstalled_FileNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: "/home/testuser/.claude/settings.json",
		Name:       "Claude Code",
	}

	installed, err := installer.IsInstalled(config)
	if err != nil {
		t.Fatalf("IsInstalled() error = %v, want nil", err)
	}

	if installed {
		t.Errorf("IsInstalled() = true, want false when file doesn't exist")
	}
}

// TestAgentInstaller_Install_Claude_NewFile tests installing to new Claude config.
func TestAgentInstaller_Install_Claude_NewFile(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "new file created") {
		t.Errorf("result.Message = %q, want message containing 'new file created'", result.Message)
	}

	// Verify the file was written
	content, ok := fs.Files[configPath]
	if !ok {
		t.Fatalf("config file not written to %q", configPath)
	}

	// Verify it contains dashlights command
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config does not contain dashlights command")
	}

	// Verify it's valid JSON
	var config_data map[string]interface{}
	if err := json.Unmarshal(content, &config_data); err != nil {
		t.Errorf("written config is not valid JSON: %v", err)
	}
}

// TestAgentInstaller_Install_Claude_MergeExisting tests merging with existing Claude config.
func TestAgentInstaller_Install_Claude_MergeExisting(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	existingConfig := `{
  "existingSetting": "value",
  "hooks": {
    "SomeOtherHook": ["data"]
  }
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "Installed dashlights") {
		t.Errorf("result.Message = %q, want message containing 'Installed dashlights'", result.Message)
	}

	if result.BackupPath == "" {
		t.Errorf("result.BackupPath is empty, want backup path")
	}

	// Verify the file was updated
	content, ok := fs.Files[configPath]
	if !ok {
		t.Fatalf("config file not found at %q", configPath)
	}

	// Verify it contains dashlights command
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config does not contain dashlights command")
	}

	// Verify existing settings are preserved
	if !strings.Contains(string(content), "existingSetting") {
		t.Errorf("existing settings were not preserved")
	}

	// Verify it's valid JSON
	var config_data map[string]interface{}
	if err := json.Unmarshal(content, &config_data); err != nil {
		t.Errorf("written config is not valid JSON: %v", err)
	}
}

// TestAgentInstaller_Install_Claude_AlreadyInstalled tests idempotency.
func TestAgentInstaller_Install_Claude_AlreadyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	existingConfig := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "dashlights --agentic"
          }
        ]
      }
    ]
  }
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "already installed") {
		t.Errorf("result.Message = %q, want message containing 'already installed'", result.Message)
	}

	// Verify backup was not created (no changes needed)
	if result.BackupPath != "" {
		t.Errorf("result.BackupPath = %q, want empty (no backup needed)", result.BackupPath)
	}
}

// TestAgentInstaller_Install_Claude_InvalidJSON tests handling of invalid JSON.
func TestAgentInstaller_Install_Claude_InvalidJSON(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	fs.Files[configPath] = []byte(`{invalid json}`)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() returned error instead of InstallResult: %v", err)
	}

	if result.ExitCode != ExitError {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitError)
	}

	if !strings.Contains(result.Message, "invalid JSON") {
		t.Errorf("result.Message = %q, want message containing 'invalid JSON'", result.Message)
	}
}

// TestAgentInstaller_Install_Claude_DryRun tests dry-run mode for Claude.
func TestAgentInstaller_Install_Claude_DryRun(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, true, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "[DRY-RUN]") {
		t.Errorf("result.Message = %q, want message containing '[DRY-RUN]'", result.Message)
	}

	if !strings.Contains(result.Message, "No changes made") {
		t.Errorf("result.Message = %q, want message containing 'No changes made'", result.Message)
	}

	// Verify the file was NOT written
	if _, ok := fs.Files[configPath]; ok {
		t.Errorf("config file was written in dry-run mode")
	}
}

// TestAgentInstaller_Install_Cursor_NewFile tests installing to new Cursor config.
func TestAgentInstaller_Install_Cursor_NewFile(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "new file created") {
		t.Errorf("result.Message = %q, want message containing 'new file created'", result.Message)
	}

	// Verify the file was written
	content, ok := fs.Files[configPath]
	if !ok {
		t.Fatalf("config file not written to %q", configPath)
	}

	// Verify it contains dashlights command
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config does not contain dashlights command")
	}

	// Verify it's valid JSON with correct structure
	var config_data map[string]interface{}
	if err := json.Unmarshal(content, &config_data); err != nil {
		t.Errorf("written config is not valid JSON: %v", err)
	}

	bse, ok := config_data["beforeShellExecution"].(map[string]interface{})
	if !ok {
		t.Errorf("config missing beforeShellExecution")
	} else {
		cmd, _ := bse["command"].(string)
		if cmd != DashlightsCommand {
			t.Errorf("beforeShellExecution.command = %q, want %q", cmd, DashlightsCommand)
		}
	}
}

// TestAgentInstaller_Install_Cursor_MergeExisting tests merging with existing Cursor config.
func TestAgentInstaller_Install_Cursor_MergeExisting(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "someSetting": "value"
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if result.BackupPath == "" {
		t.Errorf("result.BackupPath is empty, want backup path")
	}

	// Verify the file was updated
	content, ok := fs.Files[configPath]
	if !ok {
		t.Fatalf("config file not found at %q", configPath)
	}

	// Verify it contains dashlights command
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config does not contain dashlights command")
	}

	// Verify existing settings are preserved
	if !strings.Contains(string(content), "someSetting") {
		t.Errorf("existing settings were not preserved")
	}
}

// TestAgentInstaller_Install_Cursor_AlreadyInstalled tests idempotency for Cursor.
func TestAgentInstaller_Install_Cursor_AlreadyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "beforeShellExecution": {
    "command": "dashlights --agentic"
  }
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "already installed") {
		t.Errorf("result.Message = %q, want message containing 'already installed'", result.Message)
	}
}

// TestAgentInstaller_Install_Cursor_ConflictNonInteractive tests conflict handling in non-interactive mode.
func TestAgentInstaller_Install_Cursor_ConflictNonInteractive(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "beforeShellExecution": {
    "command": "some-other-command"
  }
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, false, true)
	if err != nil {
		t.Fatalf("Install() returned error instead of InstallResult: %v", err)
	}

	if result.ExitCode != ExitError {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitError)
	}

	if !strings.Contains(result.Message, "already has a beforeShellExecution hook") {
		t.Errorf("result.Message = %q, want message about existing hook conflict", result.Message)
	}

	// Verify the file was NOT modified
	content, _ := fs.Files[configPath]
	if !strings.Contains(string(content), "some-other-command") {
		t.Errorf("original config was modified in non-interactive mode")
	}

	if strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("dashlights was installed despite conflict in non-interactive mode")
	}
}

// TestAgentInstaller_Install_Cursor_DryRun tests dry-run mode for Cursor.
func TestAgentInstaller_Install_Cursor_DryRun(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, true, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "[DRY-RUN]") {
		t.Errorf("result.Message = %q, want message containing '[DRY-RUN]'", result.Message)
	}

	if !strings.Contains(result.Message, "No changes made") {
		t.Errorf("result.Message = %q, want message containing 'No changes made'", result.Message)
	}

	// Verify the file was NOT written
	if _, ok := fs.Files[configPath]; ok {
		t.Errorf("config file was written in dry-run mode")
	}
}

// TestAgentInstaller_CheckCursorConflict_NoConflict tests no conflict case.
func TestAgentInstaller_CheckCursorConflict_NoConflict(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	fs.Files[configPath] = []byte(`{"someSetting": "value"}`)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	existingCmd, hasConflict, err := installer.CheckCursorConflict(config)
	if err != nil {
		t.Fatalf("CheckCursorConflict() error = %v, want nil", err)
	}

	if hasConflict {
		t.Errorf("CheckCursorConflict() hasConflict = true, want false")
	}

	if existingCmd != "" {
		t.Errorf("CheckCursorConflict() existingCmd = %q, want empty", existingCmd)
	}
}

// TestAgentInstaller_CheckCursorConflict_HasConflict tests conflict detection.
func TestAgentInstaller_CheckCursorConflict_HasConflict(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	conflictingCmd := "some-other-command --arg"
	fs.Files[configPath] = []byte(fmt.Sprintf(`{
  "beforeShellExecution": {
    "command": "%s"
  }
}`, conflictingCmd))

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	existingCmd, hasConflict, err := installer.CheckCursorConflict(config)
	if err != nil {
		t.Fatalf("CheckCursorConflict() error = %v, want nil", err)
	}

	if !hasConflict {
		t.Errorf("CheckCursorConflict() hasConflict = false, want true")
	}

	if existingCmd != conflictingCmd {
		t.Errorf("CheckCursorConflict() existingCmd = %q, want %q", existingCmd, conflictingCmd)
	}
}

// TestAgentInstaller_CheckCursorConflict_FileNotExist tests when file doesn't exist.
func TestAgentInstaller_CheckCursorConflict_FileNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: "/home/testuser/.cursor/hooks.json",
		Name:       "Cursor",
	}

	existingCmd, hasConflict, err := installer.CheckCursorConflict(config)
	if err != nil {
		t.Fatalf("CheckCursorConflict() error = %v, want nil", err)
	}

	if hasConflict {
		t.Errorf("CheckCursorConflict() hasConflict = true, want false when file doesn't exist")
	}

	if existingCmd != "" {
		t.Errorf("CheckCursorConflict() existingCmd = %q, want empty when file doesn't exist", existingCmd)
	}
}

// TestAgentInstaller_CheckCursorConflict_DashlightsInstalled tests no conflict when dashlights already installed.
func TestAgentInstaller_CheckCursorConflict_DashlightsInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	fs.Files[configPath] = []byte(`{
  "beforeShellExecution": {
    "command": "dashlights --agentic"
  }
}`)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	existingCmd, hasConflict, err := installer.CheckCursorConflict(config)
	if err != nil {
		t.Fatalf("CheckCursorConflict() error = %v, want nil", err)
	}

	if hasConflict {
		t.Errorf("CheckCursorConflict() hasConflict = true, want false when dashlights is the hook")
	}

	if existingCmd != "" {
		t.Errorf("CheckCursorConflict() existingCmd = %q, want empty when dashlights is the hook", existingCmd)
	}
}

// TestAgentInstaller_CheckCursorConflict_InvalidJSON tests error handling for invalid JSON.
func TestAgentInstaller_CheckCursorConflict_InvalidJSON(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	fs.Files[configPath] = []byte(`{invalid json}`)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	existingCmd, hasConflict, err := installer.CheckCursorConflict(config)
	if err == nil {
		t.Fatalf("CheckCursorConflict() = (%q, %v, nil), want error for invalid JSON", existingCmd, hasConflict)
	}

	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error = %q, want error containing 'invalid JSON'", err.Error())
	}
}

// TestAgentInstaller_Install_WriteFileError tests error handling when file write fails.
func TestAgentInstaller_Install_WriteFileError(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	fs.WriteFileErr = fmt.Errorf("permission denied")
	installer := NewAgentInstaller(fs)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: "/home/testuser/.claude/settings.json",
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err == nil {
		t.Fatalf("Install() = %v, want error", result)
	}

	if !strings.Contains(err.Error(), "failed to write config") {
		t.Errorf("error = %q, want error containing 'failed to write config'", err.Error())
	}
}

// TestAgentInstaller_Install_MkdirAllError tests error handling when directory creation fails.
func TestAgentInstaller_Install_MkdirAllError(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	fs.MkdirAllErr = fmt.Errorf("permission denied")
	installer := NewAgentInstaller(fs)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: "/home/testuser/.claude/settings.json",
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err == nil {
		t.Fatalf("Install() = %v, want error", result)
	}

	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("error = %q, want error containing 'failed to create directory'", err.Error())
	}
}

// TestNewAgentInstaller tests the constructor.
func TestNewAgentInstaller(t *testing.T) {
	fs := NewMockFilesystem()
	installer := NewAgentInstaller(fs)

	if installer == nil {
		t.Fatal("NewAgentInstaller() returned nil")
	}

	if installer.fs != fs {
		t.Errorf("installer.fs is not the provided filesystem")
	}

	if installer.backup == nil {
		t.Errorf("installer.backup is nil, want BackupManager")
	}
}

// TestAgentConfig_Fields tests the AgentConfig struct fields.
func TestAgentConfig_Fields(t *testing.T) {
	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: "/path/to/config",
		Name:       "Test Agent",
	}

	if config.Type != AgentClaude {
		t.Errorf("config.Type = %q, want %q", config.Type, AgentClaude)
	}

	if config.ConfigPath != "/path/to/config" {
		t.Errorf("config.ConfigPath = %q, want %q", config.ConfigPath, "/path/to/config")
	}

	if config.Name != "Test Agent" {
		t.Errorf("config.Name = %q, want %q", config.Name, "Test Agent")
	}
}

// TestAgentType_Constants tests the AgentType constants.
func TestAgentType_Constants(t *testing.T) {
	if AgentClaude != "claude" {
		t.Errorf("AgentClaude = %q, want %q", AgentClaude, "claude")
	}

	if AgentCursor != "cursor" {
		t.Errorf("AgentCursor = %q, want %q", AgentCursor, "cursor")
	}
}

// TestAgentInstaller_Install_Cursor_InteractiveMode tests Cursor installation in interactive mode (non-conflicting).
func TestAgentInstaller_Install_Cursor_InteractiveMode(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	// Interactive mode (nonInteractive=false) should work fine when there's no conflict
	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}
}

// TestAgentInstaller_Install_Claude_WithExistingHooksArray tests merging when hooks array already exists.
func TestAgentInstaller_Install_Claude_WithExistingHooksArray(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	// Config with existing PreToolUse hooks
	existingConfig := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "OtherTool",
        "hooks": [
          {
            "type": "command",
            "command": "other-tool"
          }
        ]
      }
    ]
  }
}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	// Verify both hooks exist
	content, _ := fs.Files[configPath]
	if !strings.Contains(string(content), "dashlights --agentic") {
		t.Errorf("config does not contain dashlights command")
	}
	if !strings.Contains(string(content), "other-tool") {
		t.Errorf("existing hook was not preserved")
	}
}

// TestAgentInstaller_Install_Claude_DryRunWithExisting tests dry-run with existing file.
func TestAgentInstaller_Install_Claude_DryRunWithExisting(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.claude/settings.json"
	existingConfig := `{"existingSetting": "value"}`
	fs.Files[configPath] = []byte(existingConfig)

	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: configPath,
		Name:       "Claude Code",
	}

	result, err := installer.Install(config, true, false)
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitSuccess)
	}

	if !strings.Contains(result.Message, "[DRY-RUN]") {
		t.Errorf("result.Message = %q, want message containing '[DRY-RUN]'", result.Message)
	}

	if !strings.Contains(result.Message, "Backup:") {
		t.Errorf("result.Message = %q, want message containing 'Backup:'", result.Message)
	}

	// Verify the original file was NOT modified
	content, _ := fs.Files[configPath]
	if string(content) != existingConfig {
		t.Errorf("original file was modified in dry-run mode")
	}
}

// TestAgentInstaller_Install_Cursor_InvalidJSONExisting tests handling invalid JSON in existing Cursor config.
func TestAgentInstaller_Install_Cursor_InvalidJSONExisting(t *testing.T) {
	fs := NewMockFilesystem()
	fs.HomeDir = "/home/testuser"
	installer := NewAgentInstaller(fs)

	configPath := "/home/testuser/.cursor/hooks.json"
	fs.Files[configPath] = []byte(`{invalid json`)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: configPath,
		Name:       "Cursor",
	}

	result, err := installer.Install(config, false, false)
	if err != nil {
		t.Fatalf("Install() returned error instead of InstallResult: %v", err)
	}

	if result.ExitCode != ExitError {
		t.Errorf("result.ExitCode = %d, want %d", result.ExitCode, ExitError)
	}

	if !strings.Contains(result.Message, "invalid JSON") {
		t.Errorf("result.Message = %q, want message containing 'invalid JSON'", result.Message)
	}
}
