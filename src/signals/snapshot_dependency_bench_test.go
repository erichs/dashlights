package signals

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Benchmark the optimized version (reading .git directly)
func BenchmarkIsReleaseContext_Optimized(b *testing.B) {
	// Create a temporary git repo
	tmpDir := b.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create a tag
	exec.Command("git", "tag", "v1.0.0").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReleaseContext(context.Background())
	}
}

// Benchmark the old version (shelling out to git)
func BenchmarkIsReleaseContext_ShellOut(b *testing.B) {
	// Create a temporary git repo
	tmpDir := b.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create a tag
	exec.Command("git", "tag", "v1.0.0").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isReleaseContextShellOut()
	}
}

// Old implementation using shell commands
func isReleaseContextShellOut() bool {
	// Check if current branch/tag looks like a release
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// We're on a tag - this is a release
		return true
	}

	// Check current branch name
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	branch := strings.TrimSpace(string(output))
	// Check for common release branch patterns
	if strings.HasPrefix(branch, "release/") ||
		strings.HasPrefix(branch, "releases/") ||
		branch == "main" ||
		branch == "master" {
		return true
	}

	return false
}

// Benchmark individual helper functions
func BenchmarkGetCurrentHeadSHA(b *testing.B) {
	tmpDir := b.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getCurrentHeadSHA()
	}
}

func BenchmarkGetCurrentBranch(b *testing.B) {
	tmpDir := b.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getCurrentBranch()
	}
}

func BenchmarkIsHeadOnTag(b *testing.B) {
	tmpDir := b.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Initialize git repo and create tags
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	os.WriteFile("test.txt", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create multiple tags to simulate realistic scenario
	for i := 0; i < 10; i++ {
		exec.Command("git", "tag", filepath.Join("v1.0.", string(rune('0'+i)))).Run()
	}

	headSHA, _ := getCurrentHeadSHA()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isHeadOnTag(context.Background(), headSHA)
	}
}
