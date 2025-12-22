package install

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
)

// TestInstaller_Run_ConfigPathWithAgent tests that using --configpath with --installagent returns error.
func TestInstaller_Run_ConfigPathWithAgent(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	var stderr bytes.Buffer
	installer.SetIO(nil, nil, &stderr)

	opts := InstallOptions{
		InstallAgent:       "claude",
		ConfigPathOverride: "/custom/path",
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v", exitCode, ExitError)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "--configpath cannot be used with --installagent") {
		t.Errorf("stderr = %q, want error about --configpath with --installagent", stderrOutput)
	}
}

// TestInstaller_Run_NoAction tests that no action specified returns error.
func TestInstaller_Run_NoAction(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	var stderr bytes.Buffer
	installer.SetIO(nil, nil, &stderr)

	opts := InstallOptions{
		InstallPrompt: false,
		InstallAgent:  "",
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v", exitCode, ExitError)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "no installation action specified") {
		t.Errorf("stderr = %q, want error about no action", stderrOutput)
	}
}

// TestInstaller_Run_InstallPrompt_Success tests successful prompt installation.
func TestInstaller_Run_InstallPrompt_Success(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt:  true,
		NonInteractive: true, // Skip confirmation
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, ".bashrc") {
		t.Errorf("stdout = %q, want .bashrc in output", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Next steps:") {
		t.Errorf("stdout = %q, want next steps", stdoutOutput)
	}

	// Verify file was written
	configPath := "/home/testuser/.bashrc"
	if !mockFS.Exists(configPath) {
		t.Errorf("config file not created at %s", configPath)
	}

	content, _ := mockFS.ReadFile(configPath)
	if !strings.Contains(string(content), SentinelBegin) {
		t.Errorf("config missing SentinelBegin")
	}
}

// TestInstaller_Run_InstallPrompt_AlreadyInstalled tests that already installed returns success.
func TestInstaller_Run_InstallPrompt_AlreadyInstalled(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	configPath := "/home/testuser/.bashrc"
	mockFS.Files[configPath] = []byte(BashTemplate)

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt:  true,
		NonInteractive: true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "already installed") {
		t.Errorf("stdout = %q, want 'already installed' message", stdoutOutput)
	}
	if strings.Contains(stdoutOutput, "Next steps:") {
		t.Errorf("stdout = %q, should not have next steps for already installed", stdoutOutput)
	}
}

// TestInstaller_Run_InstallPrompt_DryRun tests dry-run mode for prompt installation.
func TestInstaller_Run_InstallPrompt_DryRun(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt: true,
		DryRun:        true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "[DRY-RUN]") {
		t.Errorf("stdout = %q, want [DRY-RUN] marker", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "No changes made") {
		t.Errorf("stdout = %q, want 'No changes made'", stdoutOutput)
	}
	if strings.Contains(stdoutOutput, "Next steps:") {
		t.Errorf("stdout = %q, should not have next steps in dry-run", stdoutOutput)
	}

	// Verify file was NOT written
	configPath := "/home/testuser/.bashrc"
	if mockFS.Exists(configPath) {
		t.Errorf("config file should not exist in dry-run mode")
	}
}

// TestInstaller_Run_InstallPrompt_NonInteractive tests non-interactive mode.
func TestInstaller_Run_InstallPrompt_NonInteractive(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/zsh"
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt:  true,
		NonInteractive: true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	// Should not contain interactive prompts
	if strings.Contains(stdoutOutput, "[y/N]") {
		t.Errorf("stdout = %q, should not contain interactive prompt in non-interactive mode", stdoutOutput)
	}
}

// TestInstaller_Run_InstallPrompt_ConfigPathIsDirectory tests error when config path is a directory.
func TestInstaller_Run_InstallPrompt_ConfigPathIsDirectory(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	// Create a directory entry in the mock filesystem
	dirPath := "/custom/dir"
	mockFS.Files[dirPath] = []byte{} // Empty content
	mockFS.Modes[dirPath] = fs.ModeDir | 0755

	installer := NewInstallerWithFS(mockFS)

	var stderr bytes.Buffer
	installer.SetIO(nil, nil, &stderr)

	opts := InstallOptions{
		InstallPrompt:      true,
		ConfigPathOverride: dirPath,
		NonInteractive:     true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v", exitCode, ExitError)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "must be a file, not a directory") {
		t.Errorf("stderr = %q, want error about directory", stderrOutput)
	}
}

// TestInstaller_Run_InstallAgent_Claude_Success tests successful Claude agent installation.
func TestInstaller_Run_InstallAgent_Claude_Success(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "claude",
		NonInteractive: true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "settings.json") {
		t.Errorf("stdout = %q, want settings.json in output", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Next steps:") {
		t.Errorf("stdout = %q, want next steps", stdoutOutput)
	}

	// Verify file was written
	configPath := "/home/testuser/.claude/settings.json"
	if !mockFS.Exists(configPath) {
		t.Errorf("config file not created at %s", configPath)
	}

	content, _ := mockFS.ReadFile(configPath)
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config missing dashlights command")
	}
}

// TestInstaller_Run_InstallAgent_Cursor_Success tests successful Cursor agent installation.
func TestInstaller_Run_InstallAgent_Cursor_Success(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "cursor",
		NonInteractive: true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "hooks.json") {
		t.Errorf("stdout = %q, want hooks.json in output", stdoutOutput)
	}

	// Verify file was written
	configPath := "/home/testuser/.cursor/hooks.json"
	if !mockFS.Exists(configPath) {
		t.Errorf("config file not created at %s", configPath)
	}

	content, _ := mockFS.ReadFile(configPath)
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config missing dashlights command")
	}
}

// TestInstaller_Run_InstallAgent_InvalidAgent tests error for invalid agent name.
func TestInstaller_Run_InstallAgent_InvalidAgent(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	var stderr bytes.Buffer
	installer.SetIO(nil, nil, &stderr)

	opts := InstallOptions{
		InstallAgent: "invalid-agent",
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v", exitCode, ExitError)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "unsupported agent") {
		t.Errorf("stderr = %q, want error about unsupported agent", stderrOutput)
	}
}

// TestInstaller_Run_InstallAgent_DryRun tests dry-run mode for agent installation.
func TestInstaller_Run_InstallAgent_DryRun(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallAgent: "claude",
		DryRun:       true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "[DRY-RUN]") {
		t.Errorf("stdout = %q, want [DRY-RUN] marker", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "No changes made") {
		t.Errorf("stdout = %q, want 'No changes made'", stdoutOutput)
	}
	if strings.Contains(stdoutOutput, "Next steps:") {
		t.Errorf("stdout = %q, should not have next steps in dry-run", stdoutOutput)
	}

	// Verify file was NOT written
	configPath := "/home/testuser/.claude/settings.json"
	if mockFS.Exists(configPath) {
		t.Errorf("config file should not exist in dry-run mode")
	}
}

// TestInstaller_Run_InstallAgent_NonInteractive tests non-interactive mode for agent installation.
func TestInstaller_Run_InstallAgent_NonInteractive(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "claude",
		NonInteractive: true,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	// Should not contain interactive prompts
	if strings.Contains(stdoutOutput, "[y/N]") {
		t.Errorf("stdout = %q, should not contain interactive prompt in non-interactive mode", stdoutOutput)
	}
}

// TestInstaller_Run_InstallAgent_Cursor_WithConflict tests Cursor installation with existing hook conflict.
func TestInstaller_Run_InstallAgent_Cursor_WithConflict(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	// Create existing Cursor config with a different hook
	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "beforeShellExecution": {
    "command": "some-other-command"
  }
}`
	mockFS.Files[configPath] = []byte(existingConfig)

	installer := NewInstallerWithFS(mockFS)

	var stdout bytes.Buffer
	installer.SetIO(nil, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "cursor",
		NonInteractive: true, // In non-interactive mode, should error
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v", exitCode, ExitError)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "already has a beforeShellExecution hook") {
		t.Errorf("stdout = %q, want error about existing hook", stdoutOutput)
	}
}

// TestInstaller_Run_InstallAgent_Cursor_InteractiveWithConflict tests Cursor installation with interactive conflict resolution.
func TestInstaller_Run_InstallAgent_Cursor_InteractiveWithConflict(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	// Create existing Cursor config with a different hook
	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "beforeShellExecution": {
    "command": "some-other-command"
  }
}`
	mockFS.Files[configPath] = []byte(existingConfig)

	installer := NewInstallerWithFS(mockFS)

	// Simulate user declining the replacement
	stdin := strings.NewReader("n\n")
	var stdout, stderr bytes.Buffer
	installer.SetIO(stdin, &stdout, &stderr)

	opts := InstallOptions{
		InstallAgent:   "cursor",
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v (user declined)", exitCode, ExitError)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Warning: Cursor only supports one beforeShellExecution hook") {
		t.Errorf("stdout = %q, want warning about Cursor limitation", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Installation cancelled. Existing hook preserved.") {
		t.Errorf("stdout = %q, want cancellation message", stdoutOutput)
	}
}

// TestInstaller_Run_InstallAgent_Cursor_InteractiveAcceptConflict tests Cursor installation with user accepting replacement.
func TestInstaller_Run_InstallAgent_Cursor_InteractiveAcceptConflict(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	// Create existing Cursor config with a different hook
	configPath := "/home/testuser/.cursor/hooks.json"
	existingConfig := `{
  "beforeShellExecution": {
    "command": "some-other-command"
  }
}`
	mockFS.Files[configPath] = []byte(existingConfig)

	installer := NewInstallerWithFS(mockFS)

	// Simulate user accepting both confirmations (conflict + install)
	// Need to provide enough "y\n" for both the conflict warning AND the install confirmation
	// Use unbuffered reader to prevent bufio from buffering all input on first read
	stdin := newUnbufferedReader("y\ny\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "cursor",
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		stdoutOutput := stdout.String()
		t.Errorf("Run() = %v, want %v. Output: %s", exitCode, ExitSuccess, stdoutOutput)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}

	// Verify the hook was replaced
	content, _ := mockFS.ReadFile(configPath)
	if !strings.Contains(string(content), DashlightsCommand) {
		t.Errorf("config should contain dashlights command")
	}
	if strings.Contains(string(content), "some-other-command") {
		t.Errorf("config should not contain old command")
	}
}

// TestInstaller_confirm_Yes tests interactive confirmation with yes response.
func TestInstaller_confirm_Yes(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase y", "y\n", true},
		{"uppercase Y", "Y\n", true},
		{"full yes", "yes\n", true},
		{"uppercase YES", "YES\n", true},
		{"yes with spaces", "  yes  \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			var stdout bytes.Buffer
			installer.SetIO(stdin, &stdout, nil)

			got := installer.confirm("Proceed?")
			if got != tt.want {
				t.Errorf("confirm() = %v, want %v for input %q", got, tt.want, tt.input)
			}

			if !strings.Contains(stdout.String(), "Proceed? [y/N]:") {
				t.Errorf("stdout should contain prompt")
			}
		})
	}
}

// TestInstaller_confirm_No tests interactive confirmation with no response.
func TestInstaller_confirm_No(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase n", "n\n", false},
		{"uppercase N", "N\n", false},
		{"full no", "no\n", false},
		{"empty line", "\n", false},
		{"random text", "maybe\n", false},
		{"just spaces", "   \n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			var stdout bytes.Buffer
			installer.SetIO(stdin, &stdout, nil)

			got := installer.confirm("Proceed?")
			if got != tt.want {
				t.Errorf("confirm() = %v, want %v for input %q", got, tt.want, tt.input)
			}
		})
	}
}

// TestInstaller_confirm_ReadError tests confirmation with read error.
func TestInstaller_confirm_ReadError(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	// Use a reader that returns an error
	stdin := &errorReader{err: os.ErrClosed}
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	got := installer.confirm("Proceed?")
	if got != false {
		t.Errorf("confirm() with read error = %v, want false", got)
	}
}

// TestInstaller_Run_InstallPrompt_InteractiveDeclined tests interactive prompt installation when user declines.
func TestInstaller_Run_InstallPrompt_InteractiveDeclined(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	stdin := strings.NewReader("n\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt:  true,
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v (user declined)", exitCode, ExitError)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installation cancelled.") {
		t.Errorf("stdout = %q, want cancellation message", stdoutOutput)
	}

	// Verify file was NOT written
	configPath := "/home/testuser/.bashrc"
	if mockFS.Exists(configPath) {
		t.Errorf("config file should not exist when user declines")
	}
}

// TestInstaller_Run_InstallPrompt_InteractiveAccepted tests interactive prompt installation when user accepts.
func TestInstaller_Run_InstallPrompt_InteractiveAccepted(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.EnvVars["SHELL"] = "/bin/bash"
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	stdin := strings.NewReader("y\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	opts := InstallOptions{
		InstallPrompt:  true,
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Dashlights Shell Installation") {
		t.Errorf("stdout = %q, want installation header", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Proceed? [y/N]:") {
		t.Errorf("stdout = %q, want confirmation prompt", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}

	// Verify file was written
	configPath := "/home/testuser/.bashrc"
	if !mockFS.Exists(configPath) {
		t.Errorf("config file not created at %s", configPath)
	}
}

// TestInstaller_Run_InstallAgent_InteractiveDeclined tests interactive agent installation when user declines.
func TestInstaller_Run_InstallAgent_InteractiveDeclined(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	stdin := strings.NewReader("n\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "claude",
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitError {
		t.Errorf("Run() = %v, want %v (user declined)", exitCode, ExitError)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Installation cancelled.") {
		t.Errorf("stdout = %q, want cancellation message", stdoutOutput)
	}

	// Verify file was NOT written
	configPath := "/home/testuser/.claude/settings.json"
	if mockFS.Exists(configPath) {
		t.Errorf("config file should not exist when user declines")
	}
}

// TestInstaller_Run_InstallAgent_InteractiveAccepted tests interactive agent installation when user accepts.
func TestInstaller_Run_InstallAgent_InteractiveAccepted(t *testing.T) {
	mockFS := NewMockFilesystem()
	mockFS.HomeDir = "/home/testuser"

	installer := NewInstallerWithFS(mockFS)

	stdin := strings.NewReader("y\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	opts := InstallOptions{
		InstallAgent:   "claude",
		NonInteractive: false,
	}

	exitCode := installer.Run(opts)

	if exitCode != ExitSuccess {
		t.Errorf("Run() = %v, want %v", exitCode, ExitSuccess)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Dashlights Claude Code Installation") {
		t.Errorf("stdout = %q, want installation header", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Proceed? [y/N]:") {
		t.Errorf("stdout = %q, want confirmation prompt", stdoutOutput)
	}
	if !strings.Contains(stdoutOutput, "Installed dashlights into") {
		t.Errorf("stdout = %q, want success message", stdoutOutput)
	}

	// Verify file was written
	configPath := "/home/testuser/.claude/settings.json"
	if !mockFS.Exists(configPath) {
		t.Errorf("config file not created at %s", configPath)
	}
}

// TestInstaller_confirmPromptInstall tests the prompt installation confirmation UI.
func TestInstaller_confirmPromptInstall(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: "/home/user/.bashrc",
		Name:       "Bash",
	}

	stdin := strings.NewReader("y\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	result := installer.confirmPromptInstall(config)

	if !result {
		t.Errorf("confirmPromptInstall() = false, want true")
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Dashlights Shell Installation") {
		t.Errorf("stdout missing header")
	}
	if !strings.Contains(stdoutOutput, "Detected shell: bash") {
		t.Errorf("stdout missing shell detection")
	}
	if !strings.Contains(stdoutOutput, "Using template: Bash") {
		t.Errorf("stdout missing template info")
	}
	if !strings.Contains(stdoutOutput, "Config file: /home/user/.bashrc") {
		t.Errorf("stdout missing config path")
	}
	if !strings.Contains(stdoutOutput, "Backup: /home/user/.bashrc.dashlights-backup") {
		t.Errorf("stdout missing backup info")
	}
	if !strings.Contains(stdoutOutput, "Add dashlights prompt function") {
		t.Errorf("stdout missing changes description")
	}
}

// TestInstaller_confirmAgentInstall tests the agent installation confirmation UI.
func TestInstaller_confirmAgentInstall(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	// Test with existing config (should show backup)
	config := &AgentConfig{
		Type:       AgentClaude,
		ConfigPath: "/home/user/.claude/settings.json",
		Name:       "Claude Code",
	}
	mockFS.Files[config.ConfigPath] = []byte("{}")

	stdin := strings.NewReader("y\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	result := installer.confirmAgentInstall(config)

	if !result {
		t.Errorf("confirmAgentInstall() = false, want true")
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Dashlights Claude Code Installation") {
		t.Errorf("stdout missing header")
	}
	if !strings.Contains(stdoutOutput, "Config file: /home/user/.claude/settings.json") {
		t.Errorf("stdout missing config path")
	}
	if !strings.Contains(stdoutOutput, "Backup:") {
		t.Errorf("stdout missing backup info for existing file")
	}
	if !strings.Contains(stdoutOutput, "Add dashlights hook to Claude Code") {
		t.Errorf("stdout missing changes description")
	}
}

// TestInstaller_confirmAgentInstall_NewFile tests agent confirmation for new file (no backup).
func TestInstaller_confirmAgentInstall_NewFile(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	config := &AgentConfig{
		Type:       AgentCursor,
		ConfigPath: "/home/user/.cursor/hooks.json",
		Name:       "Cursor",
	}
	// Don't create the file, so it doesn't exist

	stdin := strings.NewReader("y\n")
	var stdout bytes.Buffer
	installer.SetIO(stdin, &stdout, nil)

	result := installer.confirmAgentInstall(config)

	if !result {
		t.Errorf("confirmAgentInstall() = false, want true")
	}

	stdoutOutput := stdout.String()
	if strings.Contains(stdoutOutput, "Backup:") {
		t.Errorf("stdout should not mention backup for new file")
	}
}

// TestInstaller_SetIO tests the SetIO method.
func TestInstaller_SetIO(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	stdin := strings.NewReader("test\n")
	var stdout, stderr bytes.Buffer

	installer.SetIO(stdin, &stdout, &stderr)

	// Verify IO streams were set by using them
	if installer.stdin != stdin {
		t.Errorf("stdin not set correctly")
	}
	if installer.stdout != &stdout {
		t.Errorf("stdout not set correctly")
	}
	if installer.stderr != &stderr {
		t.Errorf("stderr not set correctly")
	}
}

// TestNewInstaller tests that NewInstaller creates a valid installer.
func TestNewInstaller(t *testing.T) {
	installer := NewInstaller()

	if installer == nil {
		t.Fatal("NewInstaller() returned nil")
	}
	if installer.fs == nil {
		t.Error("installer.fs is nil")
	}
	if installer.shellInstall == nil {
		t.Error("installer.shellInstall is nil")
	}
	if installer.agentInstall == nil {
		t.Error("installer.agentInstall is nil")
	}
	if installer.stdin != os.Stdin {
		t.Error("installer.stdin not set to os.Stdin")
	}
	if installer.stdout != os.Stdout {
		t.Error("installer.stdout not set to os.Stdout")
	}
	if installer.stderr != os.Stderr {
		t.Error("installer.stderr not set to os.Stderr")
	}
}

// TestNewInstallerWithFS tests that NewInstallerWithFS creates a valid installer.
func TestNewInstallerWithFS(t *testing.T) {
	mockFS := NewMockFilesystem()
	installer := NewInstallerWithFS(mockFS)

	if installer == nil {
		t.Fatal("NewInstallerWithFS() returned nil")
	}
	if installer.fs != mockFS {
		t.Error("installer.fs not set to provided filesystem")
	}
	if installer.shellInstall == nil {
		t.Error("installer.shellInstall is nil")
	}
	if installer.agentInstall == nil {
		t.Error("installer.agentInstall is nil")
	}
}

// errorReader is a helper type that always returns an error when reading.
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// unbufferedReader wraps a string and provides one byte at a time to avoid buffering issues.
type unbufferedReader struct {
	data []byte
	pos  int
}

func newUnbufferedReader(s string) *unbufferedReader {
	return &unbufferedReader{data: []byte(s)}
}

func (r *unbufferedReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	// Only read one byte at a time to prevent bufio from buffering ahead
	n = 1
	if len(p) > 0 {
		p[0] = r.data[r.pos]
		r.pos++
	}
	return n, nil
}
