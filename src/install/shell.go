package install

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ShellType represents the type of shell.
type ShellType string

const (
	ShellBash ShellType = "bash"
	ShellZsh  ShellType = "zsh"
	ShellFish ShellType = "fish"
)

// TemplateType determines which prompt template to use.
type TemplateType string

const (
	TemplateBash TemplateType = "bash"
	TemplateZsh  TemplateType = "zsh"
	TemplateFish TemplateType = "fish"
	TemplateP10k TemplateType = "p10k"
)

// ShellConfig contains shell detection and configuration information.
type ShellConfig struct {
	Shell      ShellType    // The actual shell (bash, zsh, fish)
	Template   TemplateType // Which template to use (may differ, e.g., p10k for zsh)
	ConfigPath string       // Full path, e.g., /home/user/.bashrc
	Name       string       // Human-readable, e.g., "Zsh (Powerlevel10k)"
}

// InstallState represents the current installation state.
type InstallState int

const (
	NotInstalled   InstallState = iota // Neither marker present
	FullyInstalled                     // Both markers present
	PartialInstall                     // Only one marker present (corrupted)
)

// ShellInstaller handles shell prompt installation.
type ShellInstaller struct {
	fs     Filesystem
	backup *BackupManager
}

// NewShellInstaller creates a new ShellInstaller.
func NewShellInstaller(fs Filesystem) *ShellInstaller {
	return &ShellInstaller{
		fs:     fs,
		backup: NewBackupManager(fs),
	}
}

// DetectShell detects the shell type from the $SHELL environment variable.
func (i *ShellInstaller) DetectShell() (ShellType, error) {
	shellPath := i.fs.Getenv("SHELL")
	if shellPath == "" {
		return "", fmt.Errorf("could not detect shell from $SHELL environment variable")
	}

	base := filepath.Base(shellPath)
	switch base {
	case "bash":
		return ShellBash, nil
	case "zsh":
		return ShellZsh, nil
	case "fish":
		return ShellFish, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", base)
	}
}

// GetShellConfig returns the shell configuration including paths and templates.
func (i *ShellInstaller) GetShellConfig(configPathOverride string) (*ShellConfig, error) {
	shell, err := i.DetectShell()
	if err != nil {
		return nil, err
	}

	homeDir, err := i.fs.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	config := &ShellConfig{Shell: shell}

	// Determine default config path and template
	switch shell {
	case ShellBash:
		config.ConfigPath = filepath.Join(homeDir, ".bashrc")
		config.Template = TemplateBash
		config.Name = "Bash"
	case ShellZsh:
		// Check for Powerlevel10k
		p10kPath := filepath.Join(homeDir, ".p10k.zsh")
		if i.fs.Exists(p10kPath) {
			config.ConfigPath = p10kPath
			config.Template = TemplateP10k
			config.Name = "Zsh (Powerlevel10k)"
		} else {
			config.ConfigPath = filepath.Join(homeDir, ".zshrc")
			config.Template = TemplateZsh
			config.Name = "Zsh"
		}
	case ShellFish:
		config.ConfigPath = filepath.Join(homeDir, ".config", "fish", "config.fish")
		config.Template = TemplateFish
		config.Name = "Fish"
	}

	// Override path if specified
	if configPathOverride != "" {
		config.ConfigPath = configPathOverride

		// Infer template from path
		if inferredTemplate, ok := InferTemplateFromPath(configPathOverride); ok {
			config.Template = inferredTemplate
			// Update name based on inferred template
			switch inferredTemplate {
			case TemplateP10k:
				config.Name = "Zsh (Powerlevel10k)"
			case TemplateBash:
				config.Name = "Bash"
			case TemplateZsh:
				config.Name = "Zsh"
			case TemplateFish:
				config.Name = "Fish"
			}
		}
	}

	return config, nil
}

// InferTemplateFromPath determines the template type from a config file path.
func InferTemplateFromPath(configPath string) (TemplateType, bool) {
	path := strings.ToLower(configPath)
	base := filepath.Base(path)

	switch {
	case strings.Contains(path, "p10k") || strings.HasSuffix(path, ".p10k.zsh"):
		return TemplateP10k, true
	case strings.Contains(base, "bashrc") || strings.Contains(path, "bash"):
		return TemplateBash, true
	case strings.Contains(base, "zshrc") || strings.Contains(path, "zsh"):
		return TemplateZsh, true
	case base == "config.fish" || strings.Contains(path, "fish"):
		return TemplateFish, true
	default:
		return "", false
	}
}

// CheckInstallState checks if dashlights is already installed in the config file.
func (i *ShellInstaller) CheckInstallState(config *ShellConfig) (InstallState, error) {
	content, err := i.fs.ReadFile(config.ConfigPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "not exist") {
			return NotInstalled, nil
		}
		return NotInstalled, err
	}

	hasBegin := strings.Contains(string(content), SentinelBegin)
	hasEnd := strings.Contains(string(content), SentinelEnd)

	switch {
	case hasBegin && hasEnd:
		return FullyInstalled, nil
	case hasBegin || hasEnd:
		return PartialInstall, nil
	default:
		return NotInstalled, nil
	}
}

// Install installs dashlights into the shell config file.
func (i *ShellInstaller) Install(config *ShellConfig, dryRun bool) (*InstallResult, error) {
	// Check current state
	state, err := i.CheckInstallState(config)
	if err != nil {
		return nil, err
	}

	switch state {
	case FullyInstalled:
		return &InstallResult{
			ExitCode:   ExitSuccess,
			Message:    fmt.Sprintf("Dashlights is already installed in %s\nNo changes needed.", config.ConfigPath),
			ConfigPath: config.ConfigPath,
		}, nil
	case PartialInstall:
		return &InstallResult{
			ExitCode: ExitError,
			Message: fmt.Sprintf("Error: Corrupted dashlights installation detected in %s\n"+
				"Found \"%s\" without matching \"%s\" (or vice versa).\n"+
				"Please manually remove any partial dashlights blocks and retry.",
				config.ConfigPath, SentinelBegin, SentinelEnd),
		}, nil
	}

	// Get the template
	template := GetShellTemplate(config.Template)
	if template == "" {
		return nil, fmt.Errorf("no template for type: %s", config.Template)
	}

	// Read existing content or start fresh
	existingContent, err := i.fs.ReadFile(config.ConfigPath)
	if err != nil {
		// File doesn't exist, we'll create it
		existingContent = []byte{}
	}

	// Build new content
	var newContent []byte
	if len(existingContent) > 0 {
		// Append template with newline separator
		if !strings.HasSuffix(string(existingContent), "\n") {
			existingContent = append(existingContent, '\n')
		}
		newContent = append(existingContent, '\n')
		newContent = append(newContent, []byte(template)...)
	} else {
		newContent = []byte(template)
	}

	// For P10k, also try to modify the prompt elements array
	var p10kNote string
	if config.Template == TemplateP10k {
		modifiedContent, modified, alreadyPresent, modErr := i.modifyP10kPromptElements(newContent)
		if modErr == nil {
			if alreadyPresent {
				// dashlights already in prompt elements, use unmodified content
			} else if modified {
				newContent = modifiedContent
			} else {
				// Could not find prompt elements array
				p10kNote = "\nNote: Could not locate POWERLEVEL9K_LEFT_PROMPT_ELEMENTS in " + config.ConfigPath + "\n" +
					"Please add 'dashlights' to your prompt elements array manually."
			}
		}
	}

	if dryRun {
		return i.generateDryRunResult(config, template, existingContent, p10kNote), nil
	}

	// Create backup
	backupResult, err := i.backup.CreateBackup(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
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
		WhatChanged: "Added dashlights prompt function",
	}

	if backupResult.Created {
		result.BackupPath = backupResult.BackupPath
		result.Message = fmt.Sprintf("Installed dashlights into %s\nBackup: %s%s",
			config.ConfigPath, backupResult.BackupPath, p10kNote)
	} else {
		result.Message = fmt.Sprintf("Installed dashlights into %s (new file created)%s",
			config.ConfigPath, p10kNote)
	}

	return result, nil
}

// generateDryRunResult generates a dry-run preview result.
func (i *ShellInstaller) generateDryRunResult(config *ShellConfig, template string, existingContent []byte, p10kNote string) *InstallResult {
	var preview strings.Builder
	preview.WriteString("[DRY-RUN] Would make the following changes:\n\n")

	if len(existingContent) > 0 {
		preview.WriteString(fmt.Sprintf("Backup: %s -> %s.dashlights-backup\n\n",
			config.ConfigPath, config.ConfigPath))
	}

	preview.WriteString(fmt.Sprintf("Append to %s:\n", config.ConfigPath))
	preview.WriteString(strings.Repeat("-", 48) + "\n")
	for _, line := range strings.Split(strings.TrimSuffix(template, "\n"), "\n") {
		preview.WriteString("| " + line + "\n")
	}
	preview.WriteString(strings.Repeat("-", 48) + "\n")
	preview.WriteString("\nNo changes made.")
	preview.WriteString(p10kNote)

	return &InstallResult{
		ExitCode:   ExitSuccess,
		Message:    preview.String(),
		ConfigPath: config.ConfigPath,
	}
}

// modifyP10kPromptElements adds 'dashlights' to the P10k prompt elements array.
func (i *ShellInstaller) modifyP10kPromptElements(content []byte) ([]byte, bool, bool, error) {
	// Check if dashlights is already in prompt elements (idempotency)
	elementsPattern := regexp.MustCompile(`POWERLEVEL9K_(LEFT|RIGHT)_PROMPT_ELEMENTS=\([^)]*\bdashlights\b`)
	if elementsPattern.Match(content) {
		return content, false, true, nil // Already present, no modification needed
	}

	// Regex to match: POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=( ... )
	leftPattern := regexp.MustCompile(`(?m)(POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=\()(\s*)`)

	if leftPattern.Match(content) {
		// Prepend 'dashlights' after the opening parenthesis
		modified := leftPattern.ReplaceAll(content, []byte("${1}${2}dashlights\n    "))
		return modified, true, false, nil
	}

	// Fall back to RIGHT_PROMPT_ELEMENTS
	rightPattern := regexp.MustCompile(`(?m)(POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=\()(\s*)`)
	if rightPattern.Match(content) {
		modified := rightPattern.ReplaceAll(content, []byte("${1}${2}dashlights\n    "))
		return modified, true, false, nil
	}

	// Could not find prompt elements array
	return content, false, false, nil
}
