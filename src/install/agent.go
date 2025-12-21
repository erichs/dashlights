package install

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// AgentType represents the type of AI agent.
type AgentType string

const (
	AgentClaude AgentType = "claude"
	AgentCursor AgentType = "cursor"
)

// AgentConfig contains agent configuration information.
type AgentConfig struct {
	Type       AgentType
	ConfigPath string // Full path to config file
	Name       string // Human-readable, e.g., "Claude Code"
}

// AgentInstaller handles AI agent configuration installation.
type AgentInstaller struct {
	fs     Filesystem
	backup *BackupManager
}

// NewAgentInstaller creates a new AgentInstaller.
func NewAgentInstaller(fs Filesystem) *AgentInstaller {
	return &AgentInstaller{
		fs:     fs,
		backup: NewBackupManager(fs),
	}
}

// ParseAgentType parses an agent type string.
func ParseAgentType(s string) (AgentType, error) {
	switch strings.ToLower(s) {
	case "claude":
		return AgentClaude, nil
	case "cursor":
		return AgentCursor, nil
	default:
		return "", fmt.Errorf("unsupported agent '%s'. Supported: claude, cursor", s)
	}
}

// GetAgentConfig returns the configuration for an agent type.
func (i *AgentInstaller) GetAgentConfig(agentType AgentType) (*AgentConfig, error) {
	homeDir, err := i.fs.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	switch agentType {
	case AgentClaude:
		return &AgentConfig{
			Type:       AgentClaude,
			ConfigPath: filepath.Join(homeDir, ".claude", "settings.json"),
			Name:       "Claude Code",
		}, nil
	case AgentCursor:
		return &AgentConfig{
			Type:       AgentCursor,
			ConfigPath: filepath.Join(homeDir, ".cursor", "hooks.json"),
			Name:       "Cursor",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// IsInstalled checks if dashlights is already installed for the agent.
func (i *AgentInstaller) IsInstalled(config *AgentConfig) (bool, error) {
	content, err := i.fs.ReadFile(config.ConfigPath)
	if err != nil {
		// File doesn't exist = not installed
		return false, nil
	}

	return strings.Contains(string(content), DashlightsCommand), nil
}

// Install installs the dashlights hook for the specified agent.
func (i *AgentInstaller) Install(config *AgentConfig, dryRun bool, nonInteractive bool) (*InstallResult, error) {
	// Check if already installed
	installed, err := i.IsInstalled(config)
	if err != nil {
		return nil, err
	}
	if installed {
		return &InstallResult{
			ExitCode:   ExitSuccess,
			Message:    fmt.Sprintf("Dashlights is already installed in %s\nNo changes needed.", config.ConfigPath),
			ConfigPath: config.ConfigPath,
		}, nil
	}

	// Read existing content
	existing, err := i.fs.ReadFile(config.ConfigPath)
	if err != nil {
		// File doesn't exist, start with empty
		existing = nil
	}

	// Merge configuration based on agent type
	var newContent []byte
	var conflictWarning string

	switch config.Type {
	case AgentClaude:
		newContent, err = i.mergeClaudeConfig(existing)
	case AgentCursor:
		newContent, conflictWarning, err = i.mergeCursorConfig(existing, nonInteractive)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", config.Type)
	}

	if err != nil {
		return &InstallResult{
			ExitCode: ExitError,
			Message:  err.Error(),
		}, nil
	}

	if dryRun {
		return i.generateAgentDryRunResult(config, newContent, existing != nil), nil
	}

	// Create backup if file exists
	var backupResult *BackupResult
	if existing != nil {
		backupResult, err = i.backup.CreateBackup(config.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(config.ConfigPath)
	if err := i.fs.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the new content
	if err := i.fs.WriteFile(config.ConfigPath, newContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	result := &InstallResult{
		ExitCode:    ExitSuccess,
		ConfigPath:  config.ConfigPath,
		WhatChanged: fmt.Sprintf("Added dashlights hook to %s", config.Name),
	}

	if backupResult != nil && backupResult.Created {
		result.BackupPath = backupResult.BackupPath
		result.Message = fmt.Sprintf("Installed dashlights into %s\nBackup: %s%s",
			config.ConfigPath, backupResult.BackupPath, conflictWarning)
	} else {
		result.Message = fmt.Sprintf("Installed dashlights into %s (new file created)%s",
			config.ConfigPath, conflictWarning)
	}

	return result, nil
}

// mergeClaudeConfig merges dashlights hook into Claude's settings.json.
func (i *AgentInstaller) mergeClaudeConfig(existing []byte) ([]byte, error) {
	var config map[string]interface{}

	if len(existing) == 0 {
		config = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(existing, &config); err != nil {
			return nil, fmt.Errorf("invalid JSON in existing config: %w", err)
		}
	}

	// Initialize hooks structure if needed
	hooks, _ := config["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	// Get existing PreToolUse array or create empty
	preToolUse, _ := hooks["PreToolUse"].([]interface{})
	if preToolUse == nil {
		preToolUse = []interface{}{}
	}

	// Check if dashlights hook already exists (idempotency)
	for _, entry := range preToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if innerHooks, ok := m["hooks"].([]interface{}); ok {
				for _, h := range innerHooks {
					if hm, ok := h.(map[string]interface{}); ok {
						if cmd, _ := hm["command"].(string); strings.Contains(cmd, "dashlights") {
							return existing, nil // Already installed
						}
					}
				}
			}
		}
	}

	// Append our hook entry
	dashHook := map[string]interface{}{
		"matcher": "Bash|Write|Edit",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": DashlightsCommand,
			},
		},
	}
	preToolUse = append(preToolUse, dashHook)
	hooks["PreToolUse"] = preToolUse
	config["hooks"] = hooks

	return json.MarshalIndent(config, "", "  ")
}

// mergeCursorConfig merges dashlights hook into Cursor's hooks.json.
// Returns (content, warning, error).
func (i *AgentInstaller) mergeCursorConfig(existing []byte, nonInteractive bool) ([]byte, string, error) {
	var config map[string]interface{}

	if len(existing) == 0 {
		config = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(existing, &config); err != nil {
			return nil, "", fmt.Errorf("invalid JSON in existing config: %w", err)
		}
	}

	// Check if dashlights already configured (idempotency)
	if bse, ok := config["beforeShellExecution"].(map[string]interface{}); ok {
		if cmd, _ := bse["command"].(string); strings.Contains(cmd, "dashlights") {
			return existing, "", nil // Already installed
		}

		// There's an existing non-dashlights hook
		existingCmd, _ := bse["command"].(string)
		if existingCmd != "" {
			if nonInteractive {
				// In non-interactive mode, refuse to overwrite
				return nil, "", fmt.Errorf("Error: Cursor already has a beforeShellExecution hook: %q\n"+
					"Dashlights cannot be installed without replacing it.\n"+
					"To force replacement, manually remove the existing hook first, then retry.", existingCmd)
			}
			// Return warning for interactive mode (caller handles confirmation)
			// For now, we'll proceed but warn
		}
	}

	// Set our hook
	config["beforeShellExecution"] = map[string]interface{}{
		"command": DashlightsCommand,
	}

	content, err := json.MarshalIndent(config, "", "  ")
	return content, "", err
}

// generateAgentDryRunResult generates a dry-run preview for agent installation.
func (i *AgentInstaller) generateAgentDryRunResult(config *AgentConfig, newContent []byte, hasExisting bool) *InstallResult {
	var preview strings.Builder
	preview.WriteString("[DRY-RUN] Would make the following changes:\n\n")

	if hasExisting {
		preview.WriteString(fmt.Sprintf("Backup: %s -> %s.dashlights-backup\n\n",
			config.ConfigPath, config.ConfigPath))
	}

	preview.WriteString(fmt.Sprintf("Write to %s:\n", config.ConfigPath))
	preview.WriteString(strings.Repeat("-", 48) + "\n")

	// Pretty print the JSON
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(newContent, &prettyJSON); err == nil {
		if pretty, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
			for _, line := range strings.Split(string(pretty), "\n") {
				preview.WriteString("| " + line + "\n")
			}
		}
	} else {
		preview.WriteString("| " + string(newContent) + "\n")
	}

	preview.WriteString(strings.Repeat("-", 48) + "\n")
	preview.WriteString("\nNo changes made.")

	return &InstallResult{
		ExitCode:   ExitSuccess,
		Message:    preview.String(),
		ConfigPath: config.ConfigPath,
	}
}

// CheckCursorConflict checks if there's an existing Cursor hook that would be overwritten.
// Returns (existingCommand, hasConflict).
func (i *AgentInstaller) CheckCursorConflict(config *AgentConfig) (string, bool, error) {
	content, err := i.fs.ReadFile(config.ConfigPath)
	if err != nil {
		return "", false, nil // No file = no conflict
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return "", false, fmt.Errorf("invalid JSON: %w", err)
	}

	if bse, ok := cfg["beforeShellExecution"].(map[string]interface{}); ok {
		if cmd, _ := bse["command"].(string); cmd != "" && !strings.Contains(cmd, "dashlights") {
			return cmd, true, nil
		}
	}

	return "", false, nil
}
