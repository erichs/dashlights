package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// UnsafeWorkflowSignal detects dangerous GitHub Actions workflow patterns
// including pwn request vulnerabilities and expression injection attacks.
type UnsafeWorkflowSignal struct {
	pwnRequestFiles []string
	exprInjections  []exprInjectionFinding
}

// exprInjectionFinding stores details about an expression injection vulnerability
type exprInjectionFinding struct {
	file       string
	expression string
}

// Untrusted GitHub context patterns that should not be used directly in run: blocks
var untrustedContextPatterns = []string{
	"github.event.issue.title",
	"github.event.issue.body",
	"github.event.pull_request.title",
	"github.event.pull_request.body",
	"github.event.comment.body",
	"github.event.review.body",
	"github.event.head_commit.message",
	"github.event.head_commit.author.email",
	"github.event.head_commit.author.name",
	"github.event.commits",
	"github.event.pages",
	"github.event.pull_request.head.ref",
	"github.event.pull_request.head.label",
	"github.event.pull_request.head.repo.default_branch",
	"github.head_ref",
}

// exprPattern matches ${{ ... }} expressions
var exprPattern = regexp.MustCompile(`\$\{\{\s*([^}]+)\s*\}\}`)

// NewUnsafeWorkflowSignal creates an UnsafeWorkflowSignal.
func NewUnsafeWorkflowSignal() Signal {
	return &UnsafeWorkflowSignal{}
}

// Name returns the human-readable name of the signal.
func (s *UnsafeWorkflowSignal) Name() string {
	return "Unsafe Workflow"
}

// Emoji returns the emoji associated with the signal.
func (s *UnsafeWorkflowSignal) Emoji() string {
	return "🎬"
}

// Diagnostic returns a description of detected GitHub Actions workflow issues.
func (s *UnsafeWorkflowSignal) Diagnostic() string {
	var parts []string
	if len(s.pwnRequestFiles) > 0 {
		parts = append(parts, "pwn request in "+s.pwnRequestFiles[0])
	}
	if len(s.exprInjections) > 0 {
		finding := s.exprInjections[0]
		parts = append(parts, "expression injection in "+finding.file+" (uses "+finding.expression+" in run: block)")
	}
	if len(parts) > 0 {
		return "GitHub Actions: " + strings.Join(parts, "; ")
	}
	return "GitHub Actions workflow contains security vulnerabilities"
}

// Remediation returns guidance on how to fix unsafe workflow patterns.
func (s *UnsafeWorkflowSignal) Remediation() string {
	if len(s.pwnRequestFiles) > 0 && len(s.exprInjections) > 0 {
		return "Use pull_request trigger instead of pull_request_target; set untrusted inputs to env: variables before using in run: blocks"
	}
	if len(s.pwnRequestFiles) > 0 {
		return "Use pull_request trigger instead, or add persist-credentials: false and avoid checking out PR head"
	}
	return "Set untrusted inputs to env: variables before using in run: blocks"
}

// Check scans GitHub Actions workflows for pwn request and expression injection vulnerabilities.
func (s *UnsafeWorkflowSignal) Check(ctx context.Context) bool {
	workflowsDir := ".github/workflows"

	absWorkflowsDir, err := filepath.Abs(workflowsDir)
	if err != nil {
		return false
	}

	info, err := os.Stat(absWorkflowsDir)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(absWorkflowsDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return s.hasFindings()
		default:
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		if strings.Contains(name, "..") || strings.Contains(name, "/") {
			continue
		}

		filePath, err := safeJoinPath(absWorkflowsDir, name)
		if err != nil {
			continue
		}

		s.checkWorkflowFile(ctx, filePath, name)
	}

	return s.hasFindings()
}

func (s *UnsafeWorkflowSignal) hasFindings() bool {
	return len(s.pwnRequestFiles) > 0 || len(s.exprInjections) > 0
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

func (s *UnsafeWorkflowSignal) checkWorkflowFile(ctx context.Context, filePath, name string) {
	hasPRT := s.quickScanForPullRequestTarget(filePath)
	hasExpr := s.quickScanForUntrustedExpr(filePath)

	if !hasPRT && !hasExpr {
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	s.parseAndCheckWorkflow(ctx, filePath, name, hasPRT, hasExpr)
}

// quickScanForPullRequestTarget does a fast line-by-line scan
func (s *UnsafeWorkflowSignal) quickScanForPullRequestTarget(filePath string) bool {
	cleanPath := filepath.Clean(filePath)
	file, err := os.Open(cleanPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "pull_request_target") {
			return true
		}
	}
	return false
}

// quickScanForUntrustedExpr does a fast scan for untrusted expressions
func (s *UnsafeWorkflowSignal) quickScanForUntrustedExpr(filePath string) bool {
	cleanPath := filepath.Clean(filePath)
	file, err := os.Open(cleanPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range untrustedContextPatterns {
			if strings.Contains(line, pattern) {
				return true
			}
		}
	}
	return false
}

// parseAndCheckWorkflow performs full YAML parsing to detect vulnerable patterns
func (s *UnsafeWorkflowSignal) parseAndCheckWorkflow(ctx context.Context, filePath, name string, checkPRT, checkExpr bool) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	cleanPath := filepath.Clean(filePath)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return
	}

	var workflow WorkflowExt
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return
	}

	// Check for pwn request vulnerability
	if checkPRT && workflow.hasPullRequestTargetTrigger() && workflow.hasVulnerableCheckout() {
		s.pwnRequestFiles = append(s.pwnRequestFiles, name)
	}

	// Check for expression injection vulnerability
	if checkExpr {
		if finding := s.checkExpressionInjection(&workflow, name); finding != nil {
			s.exprInjections = append(s.exprInjections, *finding)
		}
	}
}

// checkExpressionInjection checks for untrusted expressions in run: blocks
func (s *UnsafeWorkflowSignal) checkExpressionInjection(w *WorkflowExt, filename string) *exprInjectionFinding {
	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			if step.Run == "" {
				continue
			}

			// Find all expressions in the run block
			matches := exprPattern.FindAllStringSubmatch(step.Run, -1)
			for _, match := range matches {
				if len(match) < 2 {
					continue
				}
				exprContent := strings.TrimSpace(match[1])

				// Check if expression contains untrusted context
				for _, pattern := range untrustedContextPatterns {
					if strings.Contains(exprContent, pattern) {
						return &exprInjectionFinding{
							file:       filename,
							expression: pattern,
						}
					}
				}
			}
		}
	}
	return nil
}

// WorkflowExt extends Workflow with additional fields for expression injection detection
type WorkflowExt struct {
	On   interface{}       `yaml:"on"`
	Jobs map[string]JobExt `yaml:"jobs"`
}

// JobExt extends Job with additional fields
type JobExt struct {
	Steps []StepExt `yaml:"steps"`
}

// StepExt extends Step with run field for expression injection detection
type StepExt struct {
	Uses string                 `yaml:"uses"`
	With map[string]interface{} `yaml:"with"`
	Run  string                 `yaml:"run"`
	Env  map[string]string      `yaml:"env"`
}

// hasPullRequestTargetTriggerExt checks if workflow has pull_request_target trigger
func (w *WorkflowExt) hasPullRequestTargetTrigger() bool {
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
func (w *WorkflowExt) hasVulnerableCheckout() bool {
	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			if isCheckoutActionExt(step.Uses) {
				if isVulnerableCheckoutStepExt(step) {
					return true
				}
			}
		}
	}
	return false
}

// isCheckoutActionExt checks if the step uses actions/checkout
func isCheckoutActionExt(uses string) bool {
	return strings.HasPrefix(uses, "actions/checkout")
}

// isVulnerableCheckoutStepExt checks if a checkout step has vulnerable configuration
func isVulnerableCheckoutStepExt(step StepExt) bool {
	if step.With == nil {
		return false
	}

	if ref, ok := step.With["ref"]; ok {
		refStr, isString := ref.(string)
		if isString && isPRHeadRefExt(refStr) {
			return !hasPersistCredentialsFalseExt(step)
		}
	}

	return false
}

// isPRHeadRefExt checks if the ref value references the PR head
func isPRHeadRefExt(ref string) bool {
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

// hasPersistCredentialsFalseExt checks if persist-credentials is explicitly false
func hasPersistCredentialsFalseExt(step StepExt) bool {
	if pc, ok := step.With["persist-credentials"]; ok {
		switch v := pc.(type) {
		case bool:
			return !v
		case string:
			return strings.ToLower(v) == "false"
		}
	}
	return false
}
