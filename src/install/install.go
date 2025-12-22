package install

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	fs           Filesystem
	shellInstall *ShellInstaller
	agentInstall *AgentInstaller
	stdin        io.Reader
	stdout       io.Writer
	stderr       io.Writer
}

// NewInstaller creates a new Installer with OS filesystem.
func NewInstaller() *Installer {
	fs := &OSFilesystem{}
	return &Installer{
		fs:           fs,
		shellInstall: NewShellInstaller(fs),
		agentInstall: NewAgentInstaller(fs),
		stdin:        os.Stdin,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
	}
}

// NewInstallerWithFS creates a new Installer with a custom filesystem.
func NewInstallerWithFS(fs Filesystem) *Installer {
	return &Installer{
		fs:           fs,
		shellInstall: NewShellInstaller(fs),
		agentInstall: NewAgentInstaller(fs),
		stdin:        os.Stdin,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
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
