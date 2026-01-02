package install

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExitCode represents the exit code of an installation operation.
type ExitCode int

const (
	ExitSuccess ExitCode = 0
	ExitError   ExitCode = 1
)

// InstallOptions contains options for installation.
type InstallOptions struct {
	InstallAll         bool // Unified install (binary, prompt, agents)
	InstallPrompt      bool
	InstallAgent       string // Agent name (claude, cursor)
	ConfigPathOverride string // If set, overrides auto-detected config path
	NonInteractive     bool
	DryRun             bool
}

// InstallResult contains the result of an installation operation.
type InstallResult struct {
	ExitCode    ExitCode
	Message     string
	BackupPath  string
	ConfigPath  string
	WhatChanged string
}

// Installer is the main orchestrator for installation operations.
type Installer struct {
	fs            Filesystem
	shellInstall  *ShellInstaller
	agentInstall  *AgentInstaller
	binaryInstall *BinaryInstaller
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
}

// NewInstaller creates a new Installer with OS filesystem.
func NewInstaller() *Installer {
	fs := &OSFilesystem{}
	return &Installer{
		fs:            fs,
		shellInstall:  NewShellInstaller(fs),
		agentInstall:  NewAgentInstaller(fs),
		binaryInstall: NewBinaryInstaller(fs),
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}
}

// NewInstallerWithFS creates a new Installer with a custom filesystem.
func NewInstallerWithFS(fs Filesystem) *Installer {
	return &Installer{
		fs:            fs,
		shellInstall:  NewShellInstaller(fs),
		agentInstall:  NewAgentInstaller(fs),
		binaryInstall: NewBinaryInstaller(fs),
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}
}

// SetIO sets custom input/output streams for testing.
func (i *Installer) SetIO(stdin io.Reader, stdout, stderr io.Writer) {
	i.stdin = stdin
	i.stdout = stdout
	i.stderr = stderr
}

// Run executes the installation based on the provided options.
func (i *Installer) Run(opts InstallOptions) ExitCode {
	// Validate options
	if opts.ConfigPathOverride != "" && opts.InstallAgent != "" {
		fmt.Fprintln(i.stderr, "Error: --configpath cannot be used with --installagent")
		return ExitError
	}

	// Unified install handles everything
	if opts.InstallAll {
		return i.runUnifiedInstall(opts)
	}

	if opts.InstallPrompt {
		return i.runPromptInstall(opts)
	}

	if opts.InstallAgent != "" {
		return i.runAgentInstall(opts)
	}

	// Should not reach here if called correctly from main
	fmt.Fprintln(i.stderr, "Error: no installation action specified")
	return ExitError
}

// runPromptInstall handles shell prompt installation.
func (i *Installer) runPromptInstall(opts InstallOptions) ExitCode {
	// Get shell configuration
	config, err := i.shellInstall.GetShellConfig(opts.ConfigPathOverride)
	if err != nil {
		fmt.Fprintf(i.stderr, "Error: %v\n", err)
		return ExitError
	}

	// Check for config path being a directory
	if info, err := i.fs.Stat(opts.ConfigPathOverride); err == nil && info.IsDir() {
		fmt.Fprintln(i.stderr, "Error: --configpath must be a file, not a directory")
		return ExitError
	}

	// Interactive mode: show preview and confirm
	if !opts.NonInteractive && !opts.DryRun {
		if !i.confirmPromptInstall(config) {
			fmt.Fprintln(i.stdout, "Installation cancelled.")
			return ExitError
		}
	}

	// Perform installation
	result, err := i.shellInstall.Install(config, opts.DryRun)
	if err != nil {
		fmt.Fprintf(i.stderr, "Error: %v\n", err)
		return ExitError
	}

	fmt.Fprintln(i.stdout, result.Message)

	// Show next steps for successful installations
	if result.ExitCode == ExitSuccess && !opts.DryRun && result.WhatChanged != "" {
		fmt.Fprintln(i.stdout, "")
		fmt.Fprintln(i.stdout, "Next steps:")
		fmt.Fprintf(i.stdout, "  Restart your shell or run: source %s\n", config.ConfigPath)
	}

	return result.ExitCode
}

// runAgentInstall handles AI agent configuration installation.
func (i *Installer) runAgentInstall(opts InstallOptions) ExitCode {
	// Parse agent type
	agentType, err := ParseAgentType(opts.InstallAgent)
	if err != nil {
		fmt.Fprintf(i.stderr, "Error: %v\n", err)
		return ExitError
	}

	// Get agent configuration
	config, err := i.agentInstall.GetAgentConfig(agentType)
	if err != nil {
		fmt.Fprintf(i.stderr, "Error: %v\n", err)
		return ExitError
	}

	// Check for Cursor hook conflict in interactive mode
	if agentType == AgentCursor && !opts.NonInteractive && !opts.DryRun {
		existingCmd, hasConflict, err := i.agentInstall.CheckCursorConflict(config)
		if err != nil {
			fmt.Fprintf(i.stderr, "Error: %v\n", err)
			return ExitError
		}
		if hasConflict {
			fmt.Fprintf(i.stdout, "Warning: Cursor only supports one beforeShellExecution hook.\n")
			fmt.Fprintf(i.stdout, "Existing hook will be replaced. Current: %q\n", existingCmd)
			if !i.confirm("Proceed?") {
				fmt.Fprintln(i.stdout, "Installation cancelled. Existing hook preserved.")
				return ExitError
			}
		}
	}

	// Interactive mode: show preview and confirm
	if !opts.NonInteractive && !opts.DryRun {
		if !i.confirmAgentInstall(config) {
			fmt.Fprintln(i.stdout, "Installation cancelled.")
			return ExitError
		}
	}

	// Perform installation
	result, err := i.agentInstall.Install(config, opts.DryRun, opts.NonInteractive)
	if err != nil {
		fmt.Fprintf(i.stderr, "Error: %v\n", err)
		return ExitError
	}

	fmt.Fprintln(i.stdout, result.Message)

	// Show next steps for successful installations
	if result.ExitCode == ExitSuccess && !opts.DryRun && result.WhatChanged != "" {
		fmt.Fprintln(i.stdout, "")
		fmt.Fprintln(i.stdout, "Next steps:")
		fmt.Fprintf(i.stdout, "  Restart %s to apply changes\n", config.Name)
	}

	return result.ExitCode
}

// confirmPromptInstall shows an interactive confirmation for shell prompt installation.
func (i *Installer) confirmPromptInstall(config *ShellConfig) bool {
	fmt.Fprintln(i.stdout, "Dashlights Shell Installation")
	fmt.Fprintln(i.stdout, "=============================")
	fmt.Fprintln(i.stdout, "")
	fmt.Fprintf(i.stdout, "Detected shell: %s\n", config.Shell)
	fmt.Fprintf(i.stdout, "Using template: %s\n", config.Name)
	fmt.Fprintf(i.stdout, "Config file: %s\n", config.ConfigPath)
	fmt.Fprintln(i.stdout, "")
	fmt.Fprintln(i.stdout, "The following changes will be made:")
	fmt.Fprintf(i.stdout, "  - Backup: %s.dashlights-backup\n", config.ConfigPath)
	fmt.Fprintln(i.stdout, "  - Add dashlights prompt function")
	fmt.Fprintln(i.stdout, "")

	return i.confirm("Proceed?")
}

// confirmAgentInstall shows an interactive confirmation for agent installation.
func (i *Installer) confirmAgentInstall(config *AgentConfig) bool {
	fmt.Fprintf(i.stdout, "Dashlights %s Installation\n", config.Name)
	fmt.Fprintln(i.stdout, strings.Repeat("=", 30+len(config.Name)))
	fmt.Fprintln(i.stdout, "")
	fmt.Fprintf(i.stdout, "Config file: %s\n", config.ConfigPath)
	fmt.Fprintln(i.stdout, "")
	fmt.Fprintln(i.stdout, "The following changes will be made:")

	if i.fs.Exists(config.ConfigPath) {
		fmt.Fprintf(i.stdout, "  - Backup: %s.dashlights-backup\n", config.ConfigPath)
	}
	fmt.Fprintf(i.stdout, "  - Add dashlights hook to %s\n", config.Name)
	fmt.Fprintln(i.stdout, "")

	return i.confirm("Proceed?")
}

// confirm prompts for y/N confirmation and returns true if user confirms.
func (i *Installer) confirm(prompt string) bool {
	fmt.Fprintf(i.stdout, "%s [y/N]: ", prompt)

	reader := bufio.NewReader(i.stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// installComponent represents a component to install during unified installation.
type installComponent struct {
	name       string
	action     string // Description of what will be done
	targetPath string
	execute    func() (*InstallResult, error)
}

// runUnifiedInstall handles the unified --install flag.
func (i *Installer) runUnifiedInstall(opts InstallOptions) ExitCode {
	var components []installComponent
	var warnings []string

	// Get shell configuration first (needed for binary PATH export)
	shellConfig, shellErr := i.shellInstall.GetShellConfig("")
	if shellErr != nil {
		warnings = append(warnings, fmt.Sprintf("Shell detection failed: %v", shellErr))
	}

	// 1. Binary installation (always first)
	binaryConfig, binaryErr := i.binaryInstall.GetBinaryConfig(shellConfig)
	if binaryErr != nil {
		warnings = append(warnings, fmt.Sprintf("Binary config failed: %v (skipping binary installation)", binaryErr))
	} else {
		action := "Install binary"
		state, stateErr := i.binaryInstall.CheckBinaryState(binaryConfig)
		if stateErr != nil {
			warnings = append(warnings, fmt.Sprintf("Binary state check failed: %v", stateErr))
		}
		switch state {
		case BinaryInstalled:
			action = "Binary already installed (up to date)"
		case BinaryOutdated:
			action = "Update binary"
		case BinaryIsSymlink:
			action = "Binary is symlink (will skip)"
		}

		pathAction := ""
		if binaryConfig.PathNeedsExport {
			pathAction = "; add ~/.local/bin to PATH"
		}

		components = append(components, installComponent{
			name:       "Binary",
			action:     action + pathAction,
			targetPath: binaryConfig.TargetPath,
			execute: func() (*InstallResult, error) {
				return i.binaryInstall.EnsureBinaryInstalled(shellConfig, opts.DryRun)
			},
		})
	}

	// 2. Shell prompt (always, if shell detected)
	if shellConfig != nil {
		promptState, promptErr := i.shellInstall.CheckInstallState(shellConfig)
		if promptErr != nil {
			warnings = append(warnings, fmt.Sprintf("Prompt state check failed: %v", promptErr))
		}
		action := "Add dashlights prompt function"
		if promptState == FullyInstalled {
			action = "Already installed"
		}

		components = append(components, installComponent{
			name:       "Shell Prompt",
			action:     action,
			targetPath: shellConfig.ConfigPath,
			execute: func() (*InstallResult, error) {
				return i.shellInstall.Install(shellConfig, opts.DryRun)
			},
		})
	}

	// 3. Detect and add supported agents
	homeDir, homeErr := i.fs.UserHomeDir()
	if homeErr != nil {
		warnings = append(warnings, fmt.Sprintf("Cannot determine home directory: %v", homeErr))
	} else {
		// Claude Code (if ~/.claude exists)
		claudeDir := filepath.Join(homeDir, ".claude")
		if i.fs.Exists(claudeDir) {
			config, err := i.agentInstall.GetAgentConfig(AgentClaude)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Claude config error: %v", err))
			} else {
				action := "Add PreToolUse hook"
				if installed, installErr := i.agentInstall.IsInstalled(config); installErr != nil {
					warnings = append(warnings, fmt.Sprintf("Claude install check failed: %v", installErr))
				} else if installed {
					action = "Already installed"
				}

				components = append(components, installComponent{
					name:       "Claude Code",
					action:     action,
					targetPath: config.ConfigPath,
					execute: func() (*InstallResult, error) {
						return i.agentInstall.Install(config, opts.DryRun, opts.NonInteractive)
					},
				})
			}
		}

		// Cursor (if ~/.cursor exists)
		cursorDir := filepath.Join(homeDir, ".cursor")
		if i.fs.Exists(cursorDir) {
			config, err := i.agentInstall.GetAgentConfig(AgentCursor)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Cursor config error: %v", err))
			} else {
				action := "Add beforeShellExecution hook"
				if installed, installErr := i.agentInstall.IsInstalled(config); installErr != nil {
					warnings = append(warnings, fmt.Sprintf("Cursor install check failed: %v", installErr))
				} else if installed {
					action = "Already installed"
				} else {
					// Check for conflict
					existingCmd, hasConflict, conflictErr := i.agentInstall.CheckCursorConflict(config)
					if conflictErr != nil {
						warnings = append(warnings, fmt.Sprintf("Cursor conflict check failed: %v", conflictErr))
					} else if hasConflict {
						action = fmt.Sprintf("Replace existing hook (%s)", existingCmd)
					}
				}

				components = append(components, installComponent{
					name:       "Cursor",
					action:     action,
					targetPath: config.ConfigPath,
					execute: func() (*InstallResult, error) {
						return i.agentInstall.Install(config, opts.DryRun, opts.NonInteractive)
					},
				})
			}
		}
	}

	// Nothing to install?
	if len(components) == 0 {
		fmt.Fprintln(i.stderr, "Error: No components to install")
		for _, w := range warnings {
			fmt.Fprintf(i.stderr, "  - %s\n", w)
		}
		return ExitError
	}

	// Interactive confirmation
	if !opts.NonInteractive && !opts.DryRun {
		if !i.confirmUnifiedInstall(components, warnings) {
			fmt.Fprintln(i.stdout, "Installation cancelled.")
			return ExitError
		}
	}

	// Execute installations
	return i.executeUnifiedInstall(components, warnings, opts.DryRun, shellConfig)
}

// confirmUnifiedInstall shows the unified install preview and asks for confirmation.
func (i *Installer) confirmUnifiedInstall(components []installComponent, warnings []string) bool {
	fmt.Fprintln(i.stdout, "Dashlights Unified Installation")
	fmt.Fprintln(i.stdout, "================================")
	fmt.Fprintln(i.stdout, "")
	fmt.Fprintln(i.stdout, "The following components will be installed:")
	fmt.Fprintln(i.stdout, "")

	for _, c := range components {
		fmt.Fprintf(i.stdout, "  %s:\n", c.name)
		fmt.Fprintf(i.stdout, "    Action: %s\n", c.action)
		if c.targetPath != "" {
			fmt.Fprintf(i.stdout, "    Target: %s\n", c.targetPath)
		}
		fmt.Fprintln(i.stdout, "")
	}

	if len(warnings) > 0 {
		fmt.Fprintln(i.stdout, "Warnings:")
		for _, w := range warnings {
			fmt.Fprintf(i.stdout, "  - %s\n", w)
		}
		fmt.Fprintln(i.stdout, "")
	}

	return i.confirm("Proceed?")
}

// executeUnifiedInstall runs all installation components and reports results.
func (i *Installer) executeUnifiedInstall(components []installComponent, warnings []string, dryRun bool, shellConfig *ShellConfig) ExitCode {
	fmt.Fprintln(i.stdout, "Installing dashlights...")
	fmt.Fprintln(i.stdout, "")

	var results []string
	var errors []string
	hasChanges := false

	for idx, c := range components {
		result, err := c.execute()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", c.name, err))
			results = append(results, fmt.Sprintf("  [%d/%d] %s: Failed - %v", idx+1, len(components), c.name, err))
		} else {
			results = append(results, fmt.Sprintf("  [%d/%d] %s: %s", idx+1, len(components), c.name, result.Message))
			if result.WhatChanged != "" {
				hasChanges = true
			}
		}
	}

	// Print results
	for _, r := range results {
		fmt.Fprintln(i.stdout, r)
	}
	fmt.Fprintln(i.stdout, "")

	// Print warnings
	if len(warnings) > 0 || len(errors) > 0 {
		if len(errors) > 0 {
			fmt.Fprintln(i.stdout, "Errors:")
			for _, e := range errors {
				fmt.Fprintf(i.stdout, "  - %s\n", e)
			}
		}
		if len(warnings) > 0 {
			fmt.Fprintln(i.stdout, "Warnings:")
			for _, w := range warnings {
				fmt.Fprintf(i.stdout, "  - %s\n", w)
			}
		}
		fmt.Fprintln(i.stdout, "")
	}

	// Show next steps
	if hasChanges && !dryRun {
		fmt.Fprintln(i.stdout, "Next steps:")
		if shellConfig != nil {
			fmt.Fprintf(i.stdout, "  Restart your shell or run: source %s\n", shellConfig.ConfigPath)
		}
		fmt.Fprintln(i.stdout, "  Restart any AI coding assistants to apply hook changes")
	}

	if len(errors) > 0 {
		fmt.Fprintln(i.stdout, "Installation completed with errors.")
		return ExitError
	}

	fmt.Fprintln(i.stdout, "Installation complete!")
	return ExitSuccess
}
