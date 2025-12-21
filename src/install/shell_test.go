package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShellInstaller_DetectShell_Bash tests bash detection from various paths.
func TestShellInstaller_DetectShell_Bash(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
		wantShell ShellType
		wantErr   bool
	}{
		{
			name:      "bash from /bin",
			shellPath: "/bin/bash",
			wantShell: ShellBash,
			wantErr:   false,
		},
		{
			name:      "bash from /usr/bin",
			shellPath: "/usr/bin/bash",
			wantShell: ShellBash,
			wantErr:   false,
		},
		{
			name:      "bash from /usr/local/bin",
			shellPath: "/usr/local/bin/bash",
			wantShell: ShellBash,
			wantErr:   false,
		},
		{
			name:      "bash from homebrew path",
			shellPath: "/opt/homebrew/bin/bash",
			wantShell: ShellBash,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = tt.shellPath
			installer := NewShellInstaller(fs)

			got, err := installer.DetectShell()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantShell {
				t.Errorf("DetectShell() = %v, want %v", got, tt.wantShell)
			}
		})
	}
}

// TestShellInstaller_DetectShell_Zsh tests zsh detection.
func TestShellInstaller_DetectShell_Zsh(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
		wantShell ShellType
		wantErr   bool
	}{
		{
			name:      "zsh from /bin",
			shellPath: "/bin/zsh",
			wantShell: ShellZsh,
			wantErr:   false,
		},
		{
			name:      "zsh from /usr/bin",
			shellPath: "/usr/bin/zsh",
			wantShell: ShellZsh,
			wantErr:   false,
		},
		{
			name:      "zsh from /usr/local/bin",
			shellPath: "/usr/local/bin/zsh",
			wantShell: ShellZsh,
			wantErr:   false,
		},
		{
			name:      "zsh from homebrew",
			shellPath: "/opt/homebrew/bin/zsh",
			wantShell: ShellZsh,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = tt.shellPath
			installer := NewShellInstaller(fs)

			got, err := installer.DetectShell()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantShell {
				t.Errorf("DetectShell() = %v, want %v", got, tt.wantShell)
			}
		})
	}
}

// TestShellInstaller_DetectShell_Fish tests fish detection.
func TestShellInstaller_DetectShell_Fish(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
		wantShell ShellType
		wantErr   bool
	}{
		{
			name:      "fish from /usr/bin",
			shellPath: "/usr/bin/fish",
			wantShell: ShellFish,
			wantErr:   false,
		},
		{
			name:      "fish from /usr/local/bin",
			shellPath: "/usr/local/bin/fish",
			wantShell: ShellFish,
			wantErr:   false,
		},
		{
			name:      "fish from homebrew",
			shellPath: "/opt/homebrew/bin/fish",
			wantShell: ShellFish,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = tt.shellPath
			installer := NewShellInstaller(fs)

			got, err := installer.DetectShell()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantShell {
				t.Errorf("DetectShell() = %v, want %v", got, tt.wantShell)
			}
		})
	}
}

// TestShellInstaller_DetectShell_Unsupported tests error for unsupported shells.
func TestShellInstaller_DetectShell_Unsupported(t *testing.T) {
	tests := []struct {
		name      string
		shellPath string
	}{
		{
			name:      "tcsh",
			shellPath: "/bin/tcsh",
		},
		{
			name:      "csh",
			shellPath: "/bin/csh",
		},
		{
			name:      "ksh",
			shellPath: "/bin/ksh",
		},
		{
			name:      "sh",
			shellPath: "/bin/sh",
		},
		{
			name:      "dash",
			shellPath: "/bin/dash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = tt.shellPath
			installer := NewShellInstaller(fs)

			_, err := installer.DetectShell()
			if err == nil {
				t.Error("DetectShell() expected error for unsupported shell, got nil")
			}
			if !strings.Contains(err.Error(), "unsupported shell") {
				t.Errorf("DetectShell() error = %v, want 'unsupported shell' error", err)
			}
		})
	}
}

// TestShellInstaller_DetectShell_EmptyEnv tests error when $SHELL is empty.
func TestShellInstaller_DetectShell_EmptyEnv(t *testing.T) {
	fs := NewMockFilesystem()
	// Don't set SHELL env var
	installer := NewShellInstaller(fs)

	_, err := installer.DetectShell()
	if err == nil {
		t.Error("DetectShell() expected error for empty $SHELL, got nil")
	}
	if !strings.Contains(err.Error(), "could not detect shell") {
		t.Errorf("DetectShell() error = %v, want 'could not detect shell' error", err)
	}
}

// TestInferTemplateFromPath tests template inference from various paths.
func TestInferTemplateFromPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantTemplate TemplateType
		wantOk       bool
	}{
		// P10k patterns
		{
			name:         "p10k.zsh in home",
			path:         "/home/user/.p10k.zsh",
			wantTemplate: TemplateP10k,
			wantOk:       true,
		},
		{
			name:         "p10k.zsh uppercase",
			path:         "/home/user/.P10K.ZSH",
			wantTemplate: TemplateP10k,
			wantOk:       true,
		},
		{
			name:         "path contains p10k",
			path:         "/home/user/configs/p10k-theme.zsh",
			wantTemplate: TemplateP10k,
			wantOk:       true,
		},
		// Bash patterns
		{
			name:         "bashrc in home",
			path:         "/home/user/.bashrc",
			wantTemplate: TemplateBash,
			wantOk:       true,
		},
		{
			name:         "bashrc uppercase",
			path:         "/home/user/.BASHRC",
			wantTemplate: TemplateBash,
			wantOk:       true,
		},
		{
			name:         "bash_profile",
			path:         "/home/user/.bash_profile",
			wantTemplate: TemplateBash,
			wantOk:       true,
		},
		{
			name:         "path contains bash",
			path:         "/home/user/.config/bash/rc",
			wantTemplate: TemplateBash,
			wantOk:       true,
		},
		// Zsh patterns
		{
			name:         "zshrc in home",
			path:         "/home/user/.zshrc",
			wantTemplate: TemplateZsh,
			wantOk:       true,
		},
		{
			name:         "zshrc uppercase",
			path:         "/home/user/.ZSHRC",
			wantTemplate: TemplateZsh,
			wantOk:       true,
		},
		{
			name:         "path contains zsh not p10k",
			path:         "/home/user/.config/zsh/rc",
			wantTemplate: TemplateZsh,
			wantOk:       true,
		},
		// Fish patterns
		{
			name:         "config.fish",
			path:         "/home/user/.config/fish/config.fish",
			wantTemplate: TemplateFish,
			wantOk:       true,
		},
		{
			name:         "config.fish uppercase",
			path:         "/home/user/.config/fish/CONFIG.FISH",
			wantTemplate: TemplateFish,
			wantOk:       true,
		},
		{
			name:         "path contains fish",
			path:         "/home/user/fish/myconfig",
			wantTemplate: TemplateFish,
			wantOk:       true,
		},
		// Unknown patterns
		{
			name:         "unknown config",
			path:         "/home/user/.config/unknown",
			wantTemplate: "",
			wantOk:       false,
		},
		{
			name:         "empty path",
			path:         "",
			wantTemplate: "",
			wantOk:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTemplate, gotOk := InferTemplateFromPath(tt.path)
			if gotOk != tt.wantOk {
				t.Errorf("InferTemplateFromPath() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotTemplate != tt.wantTemplate {
				t.Errorf("InferTemplateFromPath() template = %v, want %v", gotTemplate, tt.wantTemplate)
			}
		})
	}
}

// TestShellInstaller_GetShellConfig_Bash tests bash configuration.
func TestShellInstaller_GetShellConfig_Bash(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"
	installer := NewShellInstaller(fs)

	config, err := installer.GetShellConfig("")
	if err != nil {
		t.Fatalf("GetShellConfig() error = %v", err)
	}

	if config.Shell != ShellBash {
		t.Errorf("config.Shell = %v, want %v", config.Shell, ShellBash)
	}
	if config.Template != TemplateBash {
		t.Errorf("config.Template = %v, want %v", config.Template, TemplateBash)
	}
	expectedPath := filepath.Join(fs.HomeDir, ".bashrc")
	if config.ConfigPath != expectedPath {
		t.Errorf("config.ConfigPath = %v, want %v", config.ConfigPath, expectedPath)
	}
	if config.Name != "Bash" {
		t.Errorf("config.Name = %v, want %v", config.Name, "Bash")
	}
}

// TestShellInstaller_GetShellConfig_Zsh tests zsh configuration.
func TestShellInstaller_GetShellConfig_Zsh(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"
	installer := NewShellInstaller(fs)

	config, err := installer.GetShellConfig("")
	if err != nil {
		t.Fatalf("GetShellConfig() error = %v", err)
	}

	if config.Shell != ShellZsh {
		t.Errorf("config.Shell = %v, want %v", config.Shell, ShellZsh)
	}
	if config.Template != TemplateZsh {
		t.Errorf("config.Template = %v, want %v", config.Template, TemplateZsh)
	}
	expectedPath := filepath.Join(fs.HomeDir, ".zshrc")
	if config.ConfigPath != expectedPath {
		t.Errorf("config.ConfigPath = %v, want %v", config.ConfigPath, expectedPath)
	}
	if config.Name != "Zsh" {
		t.Errorf("config.Name = %v, want %v", config.Name, "Zsh")
	}
}

// TestShellInstaller_GetShellConfig_ZshWithP10k tests zsh with Powerlevel10k.
func TestShellInstaller_GetShellConfig_ZshWithP10k(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"
	// Create .p10k.zsh file
	p10kPath := filepath.Join(fs.HomeDir, ".p10k.zsh")
	fs.Files[p10kPath] = []byte("# P10k config")

	installer := NewShellInstaller(fs)

	config, err := installer.GetShellConfig("")
	if err != nil {
		t.Fatalf("GetShellConfig() error = %v", err)
	}

	if config.Shell != ShellZsh {
		t.Errorf("config.Shell = %v, want %v", config.Shell, ShellZsh)
	}
	if config.Template != TemplateP10k {
		t.Errorf("config.Template = %v, want %v", config.Template, TemplateP10k)
	}
	if config.ConfigPath != p10kPath {
		t.Errorf("config.ConfigPath = %v, want %v", config.ConfigPath, p10kPath)
	}
	if config.Name != "Zsh (Powerlevel10k)" {
		t.Errorf("config.Name = %v, want %v", config.Name, "Zsh (Powerlevel10k)")
	}
}

// TestShellInstaller_GetShellConfig_Fish tests fish configuration.
func TestShellInstaller_GetShellConfig_Fish(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/usr/bin/fish"
	fs.HomeDir = "/home/testuser"
	installer := NewShellInstaller(fs)

	config, err := installer.GetShellConfig("")
	if err != nil {
		t.Fatalf("GetShellConfig() error = %v", err)
	}

	if config.Shell != ShellFish {
		t.Errorf("config.Shell = %v, want %v", config.Shell, ShellFish)
	}
	if config.Template != TemplateFish {
		t.Errorf("config.Template = %v, want %v", config.Template, TemplateFish)
	}
	expectedPath := filepath.Join(fs.HomeDir, ".config", "fish", "config.fish")
	if config.ConfigPath != expectedPath {
		t.Errorf("config.ConfigPath = %v, want %v", config.ConfigPath, expectedPath)
	}
	if config.Name != "Fish" {
		t.Errorf("config.Name = %v, want %v", config.Name, "Fish")
	}
}

// TestShellInstaller_GetShellConfig_WithOverride tests config path override.
func TestShellInstaller_GetShellConfig_WithOverride(t *testing.T) {
	tests := []struct {
		name           string
		shell          string
		overridePath   string
		wantTemplate   TemplateType
		wantName       string
		wantConfigPath string
	}{
		{
			name:           "override to p10k",
			shell:          "/bin/zsh",
			overridePath:   "/custom/.p10k.zsh",
			wantTemplate:   TemplateP10k,
			wantName:       "Zsh (Powerlevel10k)",
			wantConfigPath: "/custom/.p10k.zsh",
		},
		{
			name:           "override to bashrc",
			shell:          "/bin/bash",
			overridePath:   "/custom/.bashrc",
			wantTemplate:   TemplateBash,
			wantName:       "Bash",
			wantConfigPath: "/custom/.bashrc",
		},
		{
			name:           "override to zshrc",
			shell:          "/bin/zsh",
			overridePath:   "/custom/.zshrc",
			wantTemplate:   TemplateZsh,
			wantName:       "Zsh",
			wantConfigPath: "/custom/.zshrc",
		},
		{
			name:           "override to fish config",
			shell:          "/usr/bin/fish",
			overridePath:   "/custom/config.fish",
			wantTemplate:   TemplateFish,
			wantName:       "Fish",
			wantConfigPath: "/custom/config.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = tt.shell
			fs.HomeDir = "/home/testuser"
			installer := NewShellInstaller(fs)

			config, err := installer.GetShellConfig(tt.overridePath)
			if err != nil {
				t.Fatalf("GetShellConfig() error = %v", err)
			}

			if config.Template != tt.wantTemplate {
				t.Errorf("config.Template = %v, want %v", config.Template, tt.wantTemplate)
			}
			if config.Name != tt.wantName {
				t.Errorf("config.Name = %v, want %v", config.Name, tt.wantName)
			}
			if config.ConfigPath != tt.wantConfigPath {
				t.Errorf("config.ConfigPath = %v, want %v", config.ConfigPath, tt.wantConfigPath)
			}
		})
	}
}

// TestShellInstaller_CheckInstallState_NotInstalled tests not installed state.
func TestShellInstaller_CheckInstallState_NotInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	fs.Files[configPath] = []byte("# My bashrc\nexport PATH=$PATH:/usr/local/bin\n")

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	state, err := installer.CheckInstallState(config)
	if err != nil {
		t.Fatalf("CheckInstallState() error = %v", err)
	}
	if state != NotInstalled {
		t.Errorf("CheckInstallState() = %v, want %v", state, NotInstalled)
	}
}

// TestShellInstaller_CheckInstallState_FullyInstalled tests fully installed state.
func TestShellInstaller_CheckInstallState_FullyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	fs.Files[configPath] = []byte(fmt.Sprintf("# My bashrc\n%s\nsome code\n%s\n", SentinelBegin, SentinelEnd))

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	state, err := installer.CheckInstallState(config)
	if err != nil {
		t.Fatalf("CheckInstallState() error = %v", err)
	}
	if state != FullyInstalled {
		t.Errorf("CheckInstallState() = %v, want %v", state, FullyInstalled)
	}
}

// TestShellInstaller_CheckInstallState_PartialInstall tests partial install state.
func TestShellInstaller_CheckInstallState_PartialInstall(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "only begin marker",
			content: fmt.Sprintf("# My bashrc\n%s\nsome code\n", SentinelBegin),
		},
		{
			name:    "only end marker",
			content: fmt.Sprintf("# My bashrc\nsome code\n%s\n", SentinelEnd),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			fs.EnvVars["SHELL"] = "/bin/bash"
			fs.HomeDir = "/home/testuser"

			configPath := filepath.Join(fs.HomeDir, ".bashrc")
			fs.Files[configPath] = []byte(tt.content)

			installer := NewShellInstaller(fs)
			config := &ShellConfig{
				Shell:      ShellBash,
				Template:   TemplateBash,
				ConfigPath: configPath,
				Name:       "Bash",
			}

			state, err := installer.CheckInstallState(config)
			if err != nil {
				t.Fatalf("CheckInstallState() error = %v", err)
			}
			if state != PartialInstall {
				t.Errorf("CheckInstallState() = %v, want %v", state, PartialInstall)
			}
		})
	}
}

// TestShellInstaller_CheckInstallState_FileNotExist tests file not exist state.
func TestShellInstaller_CheckInstallState_FileNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	// Don't create the file

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	state, err := installer.CheckInstallState(config)
	if err != nil {
		t.Fatalf("CheckInstallState() error = %v", err)
	}
	if state != NotInstalled {
		t.Errorf("CheckInstallState() = %v, want %v", state, NotInstalled)
	}
}

// TestShellInstaller_Install_Bash tests bash installation.
func TestShellInstaller_Install_Bash(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	existingContent := "# My bashrc\nexport PATH=$PATH:/usr/local/bin\n"
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}
	if result.ConfigPath != configPath {
		t.Errorf("result.ConfigPath = %v, want %v", result.ConfigPath, configPath)
	}

	// Check that template was appended
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, SentinelBegin) {
		t.Error("new content should contain SentinelBegin")
	}
	if !strings.Contains(newContent, SentinelEnd) {
		t.Error("new content should contain SentinelEnd")
	}
	if !strings.Contains(newContent, existingContent) {
		t.Error("new content should contain existing content")
	}

	// Check backup was created
	backupPath := configPath + ".dashlights-backup"
	if _, exists := fs.Files[backupPath]; !exists {
		t.Error("backup file should be created")
	}
}

// TestShellInstaller_Install_Zsh tests zsh installation.
func TestShellInstaller_Install_Zsh(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".zshrc")
	existingContent := "# My zshrc\nexport PATH=$PATH:/usr/local/bin\n"
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellZsh,
		Template:   TemplateZsh,
		ConfigPath: configPath,
		Name:       "Zsh",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Check that zsh template was appended
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, "setopt prompt_subst") {
		t.Error("new content should contain zsh template")
	}
}

// TestShellInstaller_Install_Fish tests fish installation.
func TestShellInstaller_Install_Fish(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/usr/bin/fish"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".config", "fish", "config.fish")
	existingContent := "# My fish config\nset PATH $PATH /usr/local/bin\n"
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellFish,
		Template:   TemplateFish,
		ConfigPath: configPath,
		Name:       "Fish",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Check that fish template was appended
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, "function __dashlights_prompt --on-event fish_prompt") {
		t.Error("new content should contain fish template")
	}
}

// TestShellInstaller_Install_P10k tests p10k installation.
func TestShellInstaller_Install_P10k(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".p10k.zsh")
	existingContent := "# P10k config\n"
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellZsh,
		Template:   TemplateP10k,
		ConfigPath: configPath,
		Name:       "Zsh (Powerlevel10k)",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Check that p10k template was appended
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, "function prompt_dashlights()") {
		t.Error("new content should contain p10k template")
	}
	if !strings.Contains(newContent, "p10k segment") {
		t.Error("new content should contain p10k segment command")
	}
}

// TestShellInstaller_Install_P10k_AutoModifyLeft tests p10k auto-modification of LEFT_PROMPT_ELEMENTS.
func TestShellInstaller_Install_P10k_AutoModifyLeft(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".p10k.zsh")
	existingContent := `# P10k config
typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
    os_icon
    dir
    vcs
)
`
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellZsh,
		Template:   TemplateP10k,
		ConfigPath: configPath,
		Name:       "Zsh (Powerlevel10k)",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Check that dashlights was added to LEFT_PROMPT_ELEMENTS
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, "dashlights") {
		t.Error("new content should contain dashlights in prompt elements")
	}
	if !strings.Contains(newContent, "POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(") {
		t.Error("LEFT_PROMPT_ELEMENTS should still be present")
	}
}

// TestShellInstaller_Install_P10k_AutoModifyRight tests p10k auto-modification of RIGHT_PROMPT_ELEMENTS.
func TestShellInstaller_Install_P10k_AutoModifyRight(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".p10k.zsh")
	existingContent := `# P10k config
typeset -g POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(
    status
    background_jobs
    time
)
`
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellZsh,
		Template:   TemplateP10k,
		ConfigPath: configPath,
		Name:       "Zsh (Powerlevel10k)",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Check that dashlights was added to RIGHT_PROMPT_ELEMENTS
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, "dashlights") {
		t.Error("new content should contain dashlights in prompt elements")
	}
	if !strings.Contains(newContent, "POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(") {
		t.Error("RIGHT_PROMPT_ELEMENTS should still be present")
	}
}

// TestShellInstaller_Install_AlreadyInstalled tests already installed case.
func TestShellInstaller_Install_AlreadyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	fs.Files[configPath] = []byte(BashTemplate)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}
	if !strings.Contains(result.Message, "already installed") {
		t.Errorf("result.Message should indicate already installed, got: %v", result.Message)
	}

	// Content should remain unchanged
	newContent := string(fs.Files[configPath])
	if newContent != BashTemplate {
		t.Error("content should remain unchanged when already installed")
	}
}

// TestShellInstaller_Install_PartialInstall tests partial install error case.
func TestShellInstaller_Install_PartialInstall(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	// Only BEGIN marker, no END marker (corrupted state)
	fs.Files[configPath] = []byte(fmt.Sprintf("# My bashrc\n%s\nsome code\n", SentinelBegin))

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitError {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitError)
	}
	if !strings.Contains(result.Message, "Corrupted") {
		t.Errorf("result.Message should indicate corrupted installation, got: %v", result.Message)
	}
}

// TestShellInstaller_Install_DryRun tests dry-run mode.
func TestShellInstaller_Install_DryRun(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	existingContent := "# My bashrc\nexport PATH=$PATH:/usr/local/bin\n"
	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	result, err := installer.Install(config, true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}
	if !strings.Contains(result.Message, "[DRY-RUN]") {
		t.Errorf("result.Message should indicate dry-run, got: %v", result.Message)
	}
	if !strings.Contains(result.Message, "No changes made") {
		t.Errorf("result.Message should say no changes made, got: %v", result.Message)
	}

	// Content should remain unchanged
	newContent := string(fs.Files[configPath])
	if newContent != existingContent {
		t.Error("content should remain unchanged in dry-run mode")
	}

	// No backup should be created
	backupPath := configPath + ".dashlights-backup"
	if _, exists := fs.Files[backupPath]; exists {
		t.Error("backup should not be created in dry-run mode")
	}
}

// TestShellInstaller_Install_CreatesNewFile tests creating a new config file.
func TestShellInstaller_Install_CreatesNewFile(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/bash"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".bashrc")
	// Don't create the file initially

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: configPath,
		Name:       "Bash",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}
	if !strings.Contains(result.Message, "new file created") {
		t.Errorf("result.Message should indicate new file created, got: %v", result.Message)
	}

	// Check that file was created with template
	newContent := string(fs.Files[configPath])
	if !strings.Contains(newContent, SentinelBegin) {
		t.Error("new file should contain SentinelBegin")
	}
	if !strings.Contains(newContent, SentinelEnd) {
		t.Error("new file should contain SentinelEnd")
	}

	// No backup should be created for new file
	backupPath := configPath + ".dashlights-backup"
	if _, exists := fs.Files[backupPath]; exists {
		t.Error("backup should not be created for new file")
	}
}

// TestShellInstaller_Install_P10k_AlreadyInPromptElements tests idempotency when dashlights already in prompt elements.
func TestShellInstaller_Install_P10k_AlreadyInPromptElements(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/bin/zsh"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".p10k.zsh")
	existingContent := `# P10k config
typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
    dashlights
    os_icon
    dir
    vcs
)
` + BashTemplate // Already has the template too

	fs.Files[configPath] = []byte(existingContent)

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellZsh,
		Template:   TemplateP10k,
		ConfigPath: configPath,
		Name:       "Zsh (Powerlevel10k)",
	}

	result, err := installer.Install(config, false)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.ExitCode != ExitSuccess {
		t.Errorf("result.ExitCode = %v, want %v", result.ExitCode, ExitSuccess)
	}

	// Should indicate already installed
	if !strings.Contains(result.Message, "already installed") {
		t.Errorf("result.Message should indicate already installed, got: %v", result.Message)
	}
}

// TestShellInstaller_modifyP10kPromptElements tests the P10k prompt elements modification.
func TestShellInstaller_modifyP10kPromptElements(t *testing.T) {
	tests := []struct {
		name               string
		content            string
		wantModified       bool
		wantAlreadyPresent bool
		wantErr            bool
		checkContent       func(t *testing.T, content string)
	}{
		{
			name: "add to LEFT_PROMPT_ELEMENTS",
			content: `typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
    os_icon
    dir
)`,
			wantModified:       true,
			wantAlreadyPresent: false,
			wantErr:            false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "dashlights") {
					t.Error("content should contain dashlights")
				}
				// Should be added after the opening paren
				if !strings.Contains(content, "POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(") {
					t.Error("LEFT_PROMPT_ELEMENTS should be preserved")
				}
			},
		},
		{
			name: "add to RIGHT_PROMPT_ELEMENTS",
			content: `typeset -g POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(
    status
    time
)`,
			wantModified:       true,
			wantAlreadyPresent: false,
			wantErr:            false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "dashlights") {
					t.Error("content should contain dashlights")
				}
			},
		},
		{
			name: "already present in LEFT_PROMPT_ELEMENTS",
			content: `typeset -g POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
    dashlights
    os_icon
    dir
)`,
			wantModified:       false,
			wantAlreadyPresent: true,
			wantErr:            false,
			checkContent: func(t *testing.T, content string) {
				// Content should not be duplicated
				count := strings.Count(content, "dashlights")
				if count != 1 {
					t.Errorf("dashlights should appear exactly once, found %d times", count)
				}
			},
		},
		{
			name: "already present in RIGHT_PROMPT_ELEMENTS",
			content: `typeset -g POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(
    dashlights
    status
)`,
			wantModified:       false,
			wantAlreadyPresent: true,
			wantErr:            false,
			checkContent: func(t *testing.T, content string) {
				count := strings.Count(content, "dashlights")
				if count != 1 {
					t.Errorf("dashlights should appear exactly once, found %d times", count)
				}
			},
		},
		{
			name:               "no prompt elements found",
			content:            `# Some other config without prompt elements`,
			wantModified:       false,
			wantAlreadyPresent: false,
			wantErr:            false,
			checkContent: func(t *testing.T, content string) {
				if strings.Contains(content, "dashlights") {
					t.Error("content should not be modified when no prompt elements found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			installer := NewShellInstaller(fs)

			modifiedContent, modified, alreadyPresent, err := installer.modifyP10kPromptElements([]byte(tt.content))

			if (err != nil) != tt.wantErr {
				t.Errorf("modifyP10kPromptElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if modified != tt.wantModified {
				t.Errorf("modifyP10kPromptElements() modified = %v, want %v", modified, tt.wantModified)
			}
			if alreadyPresent != tt.wantAlreadyPresent {
				t.Errorf("modifyP10kPromptElements() alreadyPresent = %v, want %v", alreadyPresent, tt.wantAlreadyPresent)
			}

			if tt.checkContent != nil {
				tt.checkContent(t, string(modifiedContent))
			}
		})
	}
}

// TestShellInstaller_GetShellConfig_ErrorCases tests error handling in GetShellConfig.
func TestShellInstaller_GetShellConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFS     func(*MockFilesystem)
		wantErr     bool
		errContains string
	}{
		{
			name: "shell detection error",
			setupFS: func(fs *MockFilesystem) {
				// Don't set SHELL env var
			},
			wantErr:     true,
			errContains: "could not detect shell",
		},
		{
			name: "home directory error",
			setupFS: func(fs *MockFilesystem) {
				fs.EnvVars["SHELL"] = "/bin/bash"
				fs.HomeDirErr = os.ErrPermission
			},
			wantErr:     true,
			errContains: "could not determine home directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMockFilesystem()
			tt.setupFS(fs)

			installer := NewShellInstaller(fs)
			_, err := installer.GetShellConfig("")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetShellConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("GetShellConfig() error = %v, should contain %q", err, tt.errContains)
			}
		})
	}
}

// TestShellInstaller_CheckInstallState_ErrorCases tests error handling in CheckInstallState.
func TestShellInstaller_CheckInstallState_ErrorCases(t *testing.T) {
	fs := NewMockFilesystem()
	fs.ReadFileErr = os.ErrPermission

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellBash,
		Template:   TemplateBash,
		ConfigPath: "/home/testuser/.bashrc",
		Name:       "Bash",
	}

	_, err := installer.CheckInstallState(config)
	if err == nil {
		t.Error("CheckInstallState() expected error for read permission denied")
	}
}

// TestShellInstaller_Install_MkdirError tests handling of directory creation errors.
func TestShellInstaller_Install_MkdirError(t *testing.T) {
	fs := NewMockFilesystem()
	fs.EnvVars["SHELL"] = "/usr/bin/fish"
	fs.HomeDir = "/home/testuser"

	configPath := filepath.Join(fs.HomeDir, ".config", "fish", "config.fish")
	// Don't create the file initially (new file scenario)

	// Set up mkdir error
	fs.MkdirAllErr = os.ErrPermission

	installer := NewShellInstaller(fs)
	config := &ShellConfig{
		Shell:      ShellFish,
		Template:   TemplateFish,
		ConfigPath: configPath,
		Name:       "Fish",
	}

	_, err := installer.Install(config, false)
	if err == nil {
		t.Error("Install() expected error for mkdir permission denied")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("Install() error = %v, should contain 'failed to create directory'", err)
	}
}
