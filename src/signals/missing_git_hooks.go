package signals

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/pathsec"
)

// MissingGitHooksSignal detects when a repository has hook intent (config files
// for hook managers like Husky, pre-commit, Lefthook) but no hooks are installed.
// This often indicates someone cloned a repo but forgot to run the hook install command.
type MissingGitHooksSignal struct {
	foundIntent string
}

// NewMissingGitHooksSignal creates a MissingGitHooksSignal.
func NewMissingGitHooksSignal() Signal {
	return &MissingGitHooksSignal{}
}

// Name returns the human-readable name of the signal.
func (s *MissingGitHooksSignal) Name() string {
	return "Missing Git Hooks"
}

// Emoji returns the emoji associated with the signal.
func (s *MissingGitHooksSignal) Emoji() string {
	return "⚓" // Anchor emoji - hooks should anchor your commits
}

// Diagnostic returns details about missing git hooks.
func (s *MissingGitHooksSignal) Diagnostic() string {
	if s.foundIntent != "" {
		return "Git hooks not installed (found " + s.foundIntent + " but no hooks in hooks directory)"
	}
	return "Git hooks not installed despite hook manager configuration present"
}

// Remediation returns guidance on installing configured git hooks.
func (s *MissingGitHooksSignal) Remediation() string {
	return "Run the hook installer: npm install, pre-commit install, lefthook install, or copy hooks from .githooks/"
}

// Check looks for hook configuration without corresponding installed hooks.
func (s *MissingGitHooksSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_MISSING_GIT_HOOKS") != "" {
		return false
	}

	// Check context cancellation early
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// 1. Determine hooks path (fast .git/config parse, ~12µs)
	hooksPath := getHooksPath()

	// 2. Check if ANY standard hook exists in effective hooks dir
	if hasInstalledHooks(ctx, hooksPath) {
		return false // Hooks already installed, no warning
	}

	// 3. Check for "intent" markers
	intentMarkers := []string{
		".pre-commit-config.yaml", // pre-commit
		".husky",                  // Husky
		".lefthook.yml",           // Lefthook
		"lefthook.yml",            // Lefthook alt location
		".githooks",               // Generic convention
		".git-hooks",              // Alternative convention
	}

	for _, marker := range intentMarkers {
		// Check context cancellation in loop
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if _, err := os.Stat(marker); err == nil {
			s.foundIntent = marker
			return true // Intent but no hooks = warning!
		}
	}

	return false // No intent, no hooks, that's fine
}

// getHooksPath reads .git/config to find core.hooksPath, defaulting to .git/hooks
func getHooksPath() string {
	data, err := os.ReadFile(".git/config")
	if err != nil {
		return ".git/hooks" // Default
	}

	// Parse the config file looking for hooksPath in [core] section
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inCoreSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") {
			inCoreSection = strings.HasPrefix(strings.ToLower(line), "[core]")
			continue
		}

		// Look for hooksPath = value in [core] section
		if inCoreSection && strings.HasPrefix(strings.ToLower(line), "hookspath") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				// Validate the path doesn't contain directory traversal
				if !pathsec.IsValidPath(path) {
					return ".git/hooks"
				}
				return path
			}
		}
	}

	return ".git/hooks"
}

// hasInstalledHooks checks if any standard git hooks exist in the given directory
func hasInstalledHooks(ctx context.Context, hooksPath string) bool {
	entries, err := os.ReadDir(hooksPath)
	if err != nil {
		return false // Can't read hooks dir, bail silently
	}

	standardHooks := map[string]bool{
		"pre-commit":         true,
		"commit-msg":         true,
		"pre-push":           true,
		"prepare-commit-msg": true,
		"post-commit":        true,
		"pre-rebase":         true,
		"post-checkout":      true,
		"post-merge":         true,
	}

	for _, entry := range entries {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip .sample files that git creates by default
		if strings.HasSuffix(name, ".sample") {
			continue
		}

		if standardHooks[name] {
			return true // Found at least one installed hook
		}
	}

	return false
}
