package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PwnRequestSignal detects dangerous GitHub Actions workflow patterns
// that could lead to pwn request vulnerabilities (pull_request_target
// combined with explicit PR head checkout).
type PwnRequestSignal struct {
	vulnerableFiles   []string
	vulnerabilityType string
}

func NewPwnRequestSignal() Signal {
	return &PwnRequestSignal{}
}

func (s *PwnRequestSignal) Name() string {
	return "Pwn Request Risk"
}

func (s *PwnRequestSignal) Emoji() string {
	return "🎣" // Phishing/hook emoji - represents the attack vector
}

func (s *PwnRequestSignal) Diagnostic() string {
	if len(s.vulnerableFiles) > 0 {
		return "GitHub Actions workflow vulnerable to pwn request: " + s.vulnerableFiles[0]
	}
	return "GitHub Actions workflow contains pull_request_target with unsafe checkout"
}

func (s *PwnRequestSignal) Remediation() string {
	return "Use pull_request trigger instead, or add persist-credentials: false and avoid checking out PR head"
}

func (s *PwnRequestSignal) Check(ctx context.Context) bool {
	workflowsDir := ".github/workflows"

	// Get absolute path of workflows directory for validation
	absWorkflowsDir, err := filepath.Abs(workflowsDir)
	if err != nil {
		return false
	}

	// Fast path: check if workflows directory exists
	info, err := os.Stat(absWorkflowsDir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Read workflow files
	entries, err := os.ReadDir(absWorkflowsDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only process YAML files
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		// Validate filename to prevent directory traversal
		if strings.Contains(name, "..") || strings.Contains(name, "/") {
			continue
		}

		// Construct and validate file path
		filePath, err := safeJoinPath(absWorkflowsDir, name)
		if err != nil {
			continue
		}

		// Phase 1: Quick string scan for pull_request_target
		if !s.quickScanForTrigger(filePath) {
			continue
		}

		// Phase 2: Full YAML parse to verify vulnerability
		if s.parseAndCheckWorkflow(ctx, filePath) {
			s.vulnerableFiles = append(s.vulnerableFiles, name)
			return true // Found a vulnerability
		}
	}

	return false
}

// safeJoinPath safely joins a base directory and filename, ensuring the result
// stays within the base directory (prevents directory traversal attacks - G304)
func safeJoinPath(baseDir, filename string) (string, error) {
	// Clean the filename
	filename = filepath.Clean(filename)

	// Reject any path components
	if strings.ContainsAny(filename, `/\`) || filename == ".." || strings.HasPrefix(filename, "..") {
		return "", os.ErrInvalid
	}

	// Join and clean the full path
	fullPath := filepath.Join(baseDir, filename)
	fullPath = filepath.Clean(fullPath)

	// Verify the result is still within the base directory
	if !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) && fullPath != baseDir {
		return "", os.ErrInvalid
	}

	return fullPath, nil
}

// quickScanForTrigger does a fast line-by-line scan for pull_request_target
func (s *PwnRequestSignal) quickScanForTrigger(filePath string) bool {
	// filePath has already been validated by safeJoinPath in Check()
	// Clean again to satisfy static analysis
	cleanPath := filepath.Clean(filePath)
	file, err := os.Open(cleanPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "pull_request_target") {
			return true
		}
	}
	return false
}

// parseAndCheckWorkflow performs full YAML parsing to detect vulnerable patterns
func (s *PwnRequestSignal) parseAndCheckWorkflow(ctx context.Context, filePath string) bool {
	// Check context before heavy operation
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// filePath has already been validated by safeJoinPath in Check()
	// Clean again to satisfy static analysis
	cleanPath := filepath.Clean(filePath)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return false
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return false
	}

	// Check if workflow uses pull_request_target trigger
	if !workflow.hasPullRequestTargetTrigger() {
		return false
	}

	// Check all jobs for vulnerable checkout patterns
	return workflow.hasVulnerableCheckout()
}

// Workflow represents a GitHub Actions workflow file structure
type Workflow struct {
	On   interface{}    `yaml:"on"`
	Jobs map[string]Job `yaml:"jobs"`
}

// Job represents a job in a GitHub Actions workflow
type Job struct {
	Steps []Step `yaml:"steps"`
}

// Step represents a step in a GitHub Actions job
type Step struct {
	Uses string                 `yaml:"uses"`
	With map[string]interface{} `yaml:"with"`
}

// hasPullRequestTargetTrigger checks if workflow has pull_request_target trigger
func (w *Workflow) hasPullRequestTargetTrigger() bool {
	switch on := w.On.(type) {
	case string:
		return on == "pull_request_target"
	case []interface{}:
		for _, trigger := range on {
			if str, ok := trigger.(string); ok && str == "pull_request_target" {
				return true
			}
		}
	case map[string]interface{}:
		_, exists := on["pull_request_target"]
		return exists
	}
	return false
}

// hasVulnerableCheckout checks if any job has a vulnerable checkout pattern
func (w *Workflow) hasVulnerableCheckout() bool {
	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			if isCheckoutAction(step.Uses) {
				if isVulnerableCheckoutStep(step) {
					return true
				}
			}
		}
	}
	return false
}

// isCheckoutAction checks if the step uses actions/checkout
func isCheckoutAction(uses string) bool {
	return strings.HasPrefix(uses, "actions/checkout")
}

// isVulnerableCheckoutStep checks if a checkout step has vulnerable configuration
func isVulnerableCheckoutStep(step Step) bool {
	if step.With == nil {
		// No 'with' block - default checkout of target repo is safe
		return false
	}

	// Check for explicit PR head checkout (the main pwn request vulnerability)
	if ref, ok := step.With["ref"]; ok {
		refStr, isString := ref.(string)
		if isString && isPRHeadRef(refStr) {
			// Explicit checkout of PR head - this is vulnerable
			// unless persist-credentials is explicitly false
			return !hasPersistCredentialsFalse(step)
		}
	}

	// If no ref is specified but persist-credentials isn't false,
	// AND there's any 'with' configuration, check if it's safe
	// Default checkout (no ref) of target repo is safe
	return false
}

// isPRHeadRef checks if the ref value references the PR head
func isPRHeadRef(ref string) bool {
	// Common patterns for checking out PR head
	vulnerablePatterns := []string{
		"github.event.pull_request.head.sha",
		"github.event.pull_request.head.ref",
		"${{ github.event.pull_request.head.sha }}",
		"${{ github.event.pull_request.head.ref }}",
	}
	for _, pattern := range vulnerablePatterns {
		if strings.Contains(ref, pattern) {
			return true
		}
	}
	return false
}

// hasPersistCredentialsFalse checks if persist-credentials is explicitly false
func hasPersistCredentialsFalse(step Step) bool {
	if pc, ok := step.With["persist-credentials"]; ok {
		switch v := pc.(type) {
		case bool:
			return !v // true means NOT vulnerable (persist-credentials: false)
		case string:
			return strings.ToLower(v) == "false"
		}
	}
	// Default is true (persist-credentials enabled), which is vulnerable
	return false
}
