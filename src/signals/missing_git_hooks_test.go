package signals

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/erichs/dashlights/src/signals/internal/pathsec"
)

func TestMissingGitHooksSignal_NoGitDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No .git directory - should return false (not a repo, fail gracefully)
	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when .git directory doesn't exist")
	}
}

func TestMissingGitHooksSignal_HooksAlreadyInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks with a pre-commit hook
	os.MkdirAll(".git/hooks", 0755)
	os.WriteFile(".git/hooks/pre-commit", []byte("#!/bin/sh\nexit 0"), 0755)

	// Also create an intent marker
	os.WriteFile(".pre-commit-config.yaml", []byte("repos: []"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when hooks are already installed")
	}
}

func TestMissingGitHooksSignal_PreCommitConfigWithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty, just .sample files)
	os.MkdirAll(".git/hooks", 0755)
	os.WriteFile(".git/hooks/pre-commit.sample", []byte("#!/bin/sh\n# sample"), 0644)

	// Create pre-commit config
	os.WriteFile(".pre-commit-config.yaml", []byte("repos:\n  - repo: https://github.com/pre-commit/pre-commit-hooks"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .pre-commit-config.yaml exists but no hooks installed")
	}

	// Check diagnostic contains the intent marker
	diagnostic := signal.Diagnostic()
	if !strings.Contains(diagnostic, ".pre-commit-config.yaml") {
		t.Errorf("Expected diagnostic to contain '.pre-commit-config.yaml', got '%s'", diagnostic)
	}
}

func TestMissingGitHooksSignal_HuskyWithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	// Create .husky directory
	os.MkdirAll(".husky", 0755)
	os.WriteFile(".husky/pre-commit", []byte("#!/bin/sh\nnpm test"), 0755)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .husky exists but no hooks in .git/hooks")
	}

	// Check diagnostic mentions .husky
	diagnostic := signal.Diagnostic()
	if !strings.Contains(diagnostic, ".husky") {
		t.Errorf("Expected diagnostic to contain '.husky', got '%s'", diagnostic)
	}
}

func TestMissingGitHooksSignal_LefthookWithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	// Create .lefthook.yml
	os.WriteFile(".lefthook.yml", []byte("pre-commit:\n  commands:\n    lint:\n      run: npm run lint"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .lefthook.yml exists but no hooks installed")
	}
}

func TestMissingGitHooksSignal_GithooksWithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	// Create .githooks directory with a hook
	os.MkdirAll(".githooks", 0755)
	os.WriteFile(".githooks/pre-commit", []byte("#!/bin/sh\nmake lint"), 0755)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .githooks exists but no hooks in .git/hooks")
	}
}

func TestMissingGitHooksSignal_NoIntentNoHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no intent markers and no hooks")
	}
}

func TestMissingGitHooksSignal_CustomHooksPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git directory with custom hooksPath config
	os.MkdirAll(".git", 0755)
	gitConfig := `[core]
	repositoryformatversion = 0
	hooksPath = .husky
`
	os.WriteFile(".git/config", []byte(gitConfig), 0644)

	// Create .husky with an actual hook
	os.MkdirAll(".husky", 0755)
	os.WriteFile(".husky/pre-commit", []byte("#!/bin/sh\nnpm test"), 0755)

	// Also create a pre-commit config (intent marker)
	os.WriteFile(".pre-commit-config.yaml", []byte("repos: []"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when hooks exist in custom hooksPath (.husky)")
	}
}

func TestMissingGitHooksSignal_CustomHooksPathMissing(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git directory with custom hooksPath config pointing to non-existent dir
	os.MkdirAll(".git", 0755)
	gitConfig := `[core]
	hooksPath = .custom-hooks
`
	os.WriteFile(".git/config", []byte(gitConfig), 0644)

	// Create intent marker but custom hooks dir doesn't exist
	os.WriteFile(".pre-commit-config.yaml", []byte("repos: []"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when custom hooksPath dir doesn't exist but intent present")
	}
}

func TestMissingGitHooksSignal_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory
	os.MkdirAll(".git/hooks", 0755)

	// Create intent marker
	os.WriteFile(".pre-commit-config.yaml", []byte("repos: []"), 0644)

	signal := NewMissingGitHooksSignal()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when context is cancelled")
	}
}

func TestMissingGitHooksSignal_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory
	os.MkdirAll(".git/hooks", 0755)

	signal := NewMissingGitHooksSignal()

	// Create context that's already expired
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when context is expired")
	}
}

func TestMissingGitHooksSignal_DirectoryTraversalInHooksPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git directory with malicious hooksPath
	os.MkdirAll(".git", 0755)
	gitConfig := `[core]
	hooksPath = ../../../etc/evil
`
	os.WriteFile(".git/config", []byte(gitConfig), 0644)

	// Should fall back to default .git/hooks
	path := getHooksPath()
	if path != ".git/hooks" {
		t.Errorf("Expected '.git/hooks' for directory traversal path, got '%s'", path)
	}
}

func TestMissingGitHooksSignal_Methods(t *testing.T) {
	signal := NewMissingGitHooksSignal()

	if signal.Name() != "Missing Git Hooks" {
		t.Errorf("Unexpected Name: %s", signal.Name())
	}

	if signal.Emoji() != "⚓" {
		t.Errorf("Unexpected Emoji: %s", signal.Emoji())
	}

	remediation := signal.Remediation()
	if remediation == "" {
		t.Error("Remediation should not be empty")
	}
	if !strings.Contains(remediation, "install") {
		t.Error("Remediation should mention install")
	}
}

func TestMissingGitHooksSignal_LefthookAltLocation(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	// Create lefthook.yml (without dot prefix)
	os.WriteFile("lefthook.yml", []byte("pre-commit:\n  commands:\n    lint:\n      run: npm run lint"), 0644)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when lefthook.yml exists but no hooks installed")
	}
}

func TestMissingGitHooksSignal_GitHooksAltConvention(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/hooks directory (empty)
	os.MkdirAll(".git/hooks", 0755)

	// Create .git-hooks directory (alternative convention)
	os.MkdirAll(".git-hooks", 0755)
	os.WriteFile(".git-hooks/pre-commit", []byte("#!/bin/sh\nmake lint"), 0755)

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when .git-hooks exists but no hooks in .git/hooks")
	}
}

func TestGetHooksPath_DefaultWhenNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// No .git directory at all
	path := getHooksPath()
	if path != ".git/hooks" {
		t.Errorf("Expected '.git/hooks', got '%s'", path)
	}
}

func TestGetHooksPath_DefaultWhenNoHooksPathSet(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/config without hooksPath
	os.MkdirAll(".git", 0755)
	gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = git@github.com:example/repo.git
`
	os.WriteFile(".git/config", []byte(gitConfig), 0644)

	path := getHooksPath()
	if path != ".git/hooks" {
		t.Errorf("Expected '.git/hooks', got '%s'", path)
	}
}

func TestGetHooksPath_ParsesHooksPathCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .git/config with hooksPath
	os.MkdirAll(".git", 0755)
	gitConfig := `[core]
	repositoryformatversion = 0
	hooksPath = .husky
	filemode = true
[remote "origin"]
	url = git@github.com:example/repo.git
`
	os.WriteFile(".git/config", []byte(gitConfig), 0644)

	path := getHooksPath()
	if path != ".husky" {
		t.Errorf("Expected '.husky', got '%s'", path)
	}
}

func TestIsValidHooksPath(t *testing.T) {
	tests := []struct {
		path  string
		valid bool
	}{
		{".husky", true},
		{".git/hooks", true},
		{"scripts/hooks", true},
		{"", false},
		{"..", false},
		{"../etc", false},
		{"foo/../bar", false},
		{"/absolute/path", true}, // Absolute paths are valid
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := pathsec.IsValidPath(tc.path)
			if result != tc.valid {
				t.Errorf("pathsec.IsValidPath(%q) = %v, want %v", tc.path, result, tc.valid)
			}
		})
	}
}

func TestHasInstalledHooks_IgnoresSampleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create hooks directory with only .sample files
	hooksDir := filepath.Join(tmpDir, "hooks")
	os.MkdirAll(hooksDir, 0755)
	os.WriteFile(filepath.Join(hooksDir, "pre-commit.sample"), []byte("#!/bin/sh"), 0644)
	os.WriteFile(filepath.Join(hooksDir, "commit-msg.sample"), []byte("#!/bin/sh"), 0644)

	ctx := context.Background()
	result := hasInstalledHooks(ctx, hooksDir)
	if result {
		t.Error("Expected false when only .sample files exist")
	}
}

func TestHasInstalledHooks_DetectsRealHooks(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create hooks directory with real hooks
	hooksDir := filepath.Join(tmpDir, "hooks")
	os.MkdirAll(hooksDir, 0755)
	os.WriteFile(filepath.Join(hooksDir, "pre-commit.sample"), []byte("#!/bin/sh"), 0644)
	os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte("#!/bin/sh\nexit 0"), 0755)

	ctx := context.Background()
	result := hasInstalledHooks(ctx, hooksDir)
	if !result {
		t.Error("Expected true when real hooks exist")
	}
}

func TestHasInstalledHooks_DetectsVariousHookTypes(t *testing.T) {
	hookTypes := []string{
		"pre-commit",
		"commit-msg",
		"pre-push",
		"prepare-commit-msg",
		"post-commit",
		"pre-rebase",
		"post-checkout",
		"post-merge",
	}

	for _, hookType := range hookTypes {
		t.Run(hookType, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(originalDir)

			hooksDir := filepath.Join(tmpDir, "hooks")
			os.MkdirAll(hooksDir, 0755)
			os.WriteFile(filepath.Join(hooksDir, hookType), []byte("#!/bin/sh"), 0755)

			ctx := context.Background()
			result := hasInstalledHooks(ctx, hooksDir)
			if !result {
				t.Errorf("Expected true when %s hook exists", hookType)
			}
		})
	}
}

func TestMissingGitHooksSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_MISSING_GIT_HOOKS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_MISSING_GIT_HOOKS")

	signal := NewMissingGitHooksSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
