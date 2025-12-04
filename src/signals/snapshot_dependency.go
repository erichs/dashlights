package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// SnapshotDependencySignal checks for SNAPSHOT dependencies on release branches
// SNAPSHOT versions should not be used in production releases
type SnapshotDependencySignal struct {
	foundSnapshot string
	fileType      string
}

// NewSnapshotDependencySignal creates a SnapshotDependencySignal.
func NewSnapshotDependencySignal() Signal {
	return &SnapshotDependencySignal{}
}

// Name returns the human-readable name of the signal.
func (s *SnapshotDependencySignal) Name() string {
	return "Snapshot Dependency"
}

// Emoji returns the emoji associated with the signal.
func (s *SnapshotDependencySignal) Emoji() string {
	return "☕" // Coffee (Java)
}

// Diagnostic returns a description of detected SNAPSHOT dependencies.
func (s *SnapshotDependencySignal) Diagnostic() string {
	if s.foundSnapshot != "" {
		return "SNAPSHOT dependency found on release branch: " + s.foundSnapshot
	}
	return "SNAPSHOT dependencies found on release branch (unstable for production)"
}

// Remediation returns guidance on replacing SNAPSHOT versions with stable releases.
func (s *SnapshotDependencySignal) Remediation() string {
	return "Replace SNAPSHOT versions with stable releases before tagging"
}

// Check scans Java build files for SNAPSHOT dependencies on release branches/tags.
func (s *SnapshotDependencySignal) Check(ctx context.Context) bool {
	// First check if we're on a release branch or tag
	if !isReleaseContext(ctx) {
		// Not on a release branch/tag, SNAPSHOT is OK
		return false
	}

	// Check pom.xml (Maven)
	if hasPomXML() {
		if s.checkPomXML() {
			s.fileType = "pom.xml"
			return true
		}
	}

	// Check build.gradle (Gradle)
	if hasBuildGradle() {
		if s.checkBuildGradle() {
			s.fileType = "build.gradle"
			return true
		}
	}

	return false
}

// isReleaseContext checks if we're on a release branch or tag
// Optimized to read .git directory directly instead of shelling out
func isReleaseContext(ctx context.Context) bool {
	// Get current HEAD SHA
	headSHA, err := getCurrentHeadSHA()
	if err != nil {
		return false
	}

	// Check if HEAD matches any tag (indicates we're on a release tag)
	if isHeadOnTag(ctx, headSHA) {
		return true
	}

	// Check current branch name
	branch, err := getCurrentBranch()
	if err != nil {
		return false
	}

	// Check for common release branch patterns
	if strings.HasPrefix(branch, "release/") ||
		strings.HasPrefix(branch, "releases/") ||
		branch == "main" ||
		branch == "master" {
		return true
	}

	return false
}

// getCurrentHeadSHA reads the current HEAD SHA from .git/HEAD
func getCurrentHeadSHA() (string, error) {
	// Read .git/HEAD
	headContent, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return "", err
	}

	headStr := strings.TrimSpace(string(headContent))

	// If HEAD is a direct SHA (detached HEAD)
	if !strings.HasPrefix(headStr, "ref:") {
		return headStr, nil
	}

	// HEAD points to a ref, read that ref
	refPath := strings.TrimPrefix(headStr, "ref: ")

	// Validate the ref path to prevent directory traversal
	if strings.Contains(refPath, "..") {
		return "", os.ErrInvalid
	}

	// Clean the path to normalize it
	refPath = filepath.Clean(refPath)

	// Additional validation: ensure the path doesn't escape .git directory
	if strings.HasPrefix(refPath, "..") || strings.Contains(refPath, "/..") {
		return "", os.ErrInvalid
	}

	refPath = filepath.Join(".git", refPath)

	// Final validation: clean the full path
	refPath = filepath.Clean(refPath)

	shaContent, err := os.ReadFile(refPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(shaContent)), nil
}

// getCurrentBranch reads the current branch name from .git/HEAD
func getCurrentBranch() (string, error) {
	headContent, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return "", err
	}

	headStr := strings.TrimSpace(string(headContent))

	// If HEAD is detached, return empty
	if !strings.HasPrefix(headStr, "ref:") {
		return "", nil
	}

	// Extract branch name from ref: refs/heads/branch-name
	refPath := strings.TrimPrefix(headStr, "ref: ")
	if strings.HasPrefix(refPath, "refs/heads/") {
		return strings.TrimPrefix(refPath, "refs/heads/"), nil
	}

	return "", nil
}

// isHeadOnTag checks if the current HEAD SHA matches any tag
func isHeadOnTag(ctx context.Context, headSHA string) bool {
	// Check .git/refs/tags/* for matching SHAs
	tagsDir := ".git/refs/tags"
	entries, err := os.ReadDir(tagsDir)
	if err != nil {
		// No tags directory or can't read it
		return false
	}

	for _, entry := range entries {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if entry.IsDir() {
			continue
		}

		// Validate entry name to prevent directory traversal
		entryName := entry.Name()
		if strings.Contains(entryName, "..") || strings.Contains(entryName, "/") {
			continue
		}

		// Clean the entry name
		entryName = filepath.Clean(entryName)

		// Additional validation after cleaning
		if strings.Contains(entryName, "..") || strings.Contains(entryName, "/") {
			continue
		}

		tagPath := filepath.Join(tagsDir, entryName)

		// Final validation: clean the full path
		tagPath = filepath.Clean(tagPath)

		tagSHA, err := os.ReadFile(tagPath)
		if err != nil {
			continue
		}

		if strings.TrimSpace(string(tagSHA)) == headSHA {
			return true
		}
	}

	return false
}

func hasPomXML() bool {
	_, err := os.Stat("pom.xml")
	return err == nil
}

func hasBuildGradle() bool {
	_, err := os.Stat("build.gradle")
	return err == nil
}

func (s *SnapshotDependencySignal) checkPomXML() bool {
	file, err := os.Open("pom.xml")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "SNAPSHOT") && strings.Contains(line, "<version>") {
			s.foundSnapshot = strings.TrimSpace(line)
			return true
		}
	}

	return false
}

func (s *SnapshotDependencySignal) checkBuildGradle() bool {
	file, err := os.Open("build.gradle")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for SNAPSHOT in dependency declarations
		if strings.Contains(line, "SNAPSHOT") &&
			(strings.Contains(line, "implementation") ||
				strings.Contains(line, "compile") ||
				strings.Contains(line, "api") ||
				strings.Contains(line, "testImplementation")) {
			s.foundSnapshot = strings.TrimSpace(line)
			return true
		}
	}

	return false
}
