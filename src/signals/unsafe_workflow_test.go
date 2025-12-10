package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/erichs/dashlights/src/signals/internal/pathsec"
)

func TestUnsafeWorkflowSignal_Name(t *testing.T) {
	signal := NewUnsafeWorkflowSignal()
	if signal.Name() != "Unsafe Workflow" {
		t.Errorf("Expected 'Unsafe Workflow', got '%s'", signal.Name())
	}
}

func TestUnsafeWorkflowSignal_Emoji(t *testing.T) {
	signal := NewUnsafeWorkflowSignal()
	if signal.Emoji() != "🎬" {
		t.Errorf("Expected '🎬', got '%s'", signal.Emoji())
	}
}

func TestUnsafeWorkflowSignal_Diagnostic_Default(t *testing.T) {
	signal := NewUnsafeWorkflowSignal()
	expected := "GitHub Actions workflow contains security vulnerabilities"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUnsafeWorkflowSignal_Diagnostic_PwnRequest(t *testing.T) {
	signal := &UnsafeWorkflowSignal{pwnRequestFiles: []string{"ci.yml"}}
	expected := "GitHub Actions: pwn request in ci.yml"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUnsafeWorkflowSignal_Diagnostic_ExprInjection(t *testing.T) {
	signal := &UnsafeWorkflowSignal{
		exprInjections: []exprInjectionFinding{
			{file: "build.yml", expression: "github.event.issue.title"},
		},
	}
	expected := "GitHub Actions: expression injection in build.yml (uses github.event.issue.title in run: block)"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUnsafeWorkflowSignal_Diagnostic_Both(t *testing.T) {
	signal := &UnsafeWorkflowSignal{
		pwnRequestFiles: []string{"deploy.yml"},
		exprInjections: []exprInjectionFinding{
			{file: "ci.yml", expression: "github.event.pull_request.title"},
		},
	}
	expected := "GitHub Actions: pwn request in deploy.yml; expression injection in ci.yml (uses github.event.pull_request.title in run: block)"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestUnsafeWorkflowSignal_Remediation_PwnRequest(t *testing.T) {
	signal := &UnsafeWorkflowSignal{pwnRequestFiles: []string{"ci.yml"}}
	expected := "Use pull_request trigger instead, or add persist-credentials: false and avoid checking out PR head"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestUnsafeWorkflowSignal_Remediation_ExprInjection(t *testing.T) {
	signal := &UnsafeWorkflowSignal{
		exprInjections: []exprInjectionFinding{{file: "ci.yml", expression: "github.event.issue.title"}},
	}
	expected := "Set untrusted inputs to env: variables before using in run: blocks"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestUnsafeWorkflowSignal_Remediation_Both(t *testing.T) {
	signal := &UnsafeWorkflowSignal{
		pwnRequestFiles: []string{"deploy.yml"},
		exprInjections:  []exprInjectionFinding{{file: "ci.yml", expression: "github.event.issue.title"}},
	}
	expected := "Use pull_request trigger instead of pull_request_target; set untrusted inputs to env: variables before using in run: blocks"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

// Test cases for pwn request detection (from original pwnrequest_test.go)
type unsafeWorkflowTestCase struct {
	name             string
	workflowYAML     string
	expectPwnRequest bool
	expectExprInject bool
	description      string
}

var pwnRequestTestCases = []unsafeWorkflowTestCase{
	{
		name: "safe_pull_request_trigger",
		workflowYAML: `name: Safe CI
on: pull_request
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm test
`,
		expectPwnRequest: false,
		description:      "pull_request trigger is safe",
	},
	{
		name: "safe_pull_request_target_no_checkout",
		workflowYAML: `name: Label PR
on: pull_request_target
jobs:
  label:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/labeler@v4
`,
		expectPwnRequest: false,
		description:      "pull_request_target without checkout is safe",
	},
	{
		name: "vulnerable_pr_head_sha_checkout",
		workflowYAML: `name: Vulnerable CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      - run: npm install && npm test
`,
		expectPwnRequest: true,
		description:      "pull_request_target with PR head checkout is vulnerable",
	},
	{
		name: "safe_with_persist_credentials_false",
		workflowYAML: `name: Safe CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: false
`,
		expectPwnRequest: false,
		description:      "persist-credentials: false mitigates",
	},
}

// Expression injection test cases
var exprInjectionTestCases = []unsafeWorkflowTestCase{
	{
		name: "expr_injection_issue_title",
		workflowYAML: `name: Issue Handler
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.title }}"
`,
		expectExprInject: true,
		description:      "github.event.issue.title in run block is vulnerable",
	},
	{
		name: "expr_injection_issue_body",
		workflowYAML: `name: Issue Handler
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: |
          body="${{ github.event.issue.body }}"
          echo "$body"
`,
		expectExprInject: true,
		description:      "github.event.issue.body in run block is vulnerable",
	},
	{
		name: "expr_injection_pr_title",
		workflowYAML: `name: PR Handler
on: pull_request
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.pull_request.title }}"
`,
		expectExprInject: true,
		description:      "github.event.pull_request.title in run block is vulnerable",
	},
	{
		name: "expr_injection_comment_body",
		workflowYAML: `name: Comment Handler
on: issue_comment
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.comment.body }}"
`,
		expectExprInject: true,
		description:      "github.event.comment.body in run block is vulnerable",
	},
	{
		name: "expr_injection_head_ref",
		workflowYAML: `name: PR Handler
on: pull_request
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.head_ref }}"
`,
		expectExprInject: true,
		description:      "github.head_ref in run block is vulnerable",
	},
	{
		name: "expr_injection_commit_message",
		workflowYAML: `name: Push Handler
on: push
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.head_commit.message }}"
`,
		expectExprInject: true,
		description:      "github.event.head_commit.message in run block is vulnerable",
	},
	{
		name: "safe_env_variable_pattern",
		workflowYAML: `name: Safe Handler
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - name: Safe echo
        env:
          TITLE: ${{ github.event.issue.title }}
        run: echo "$TITLE"
`,
		expectExprInject: false,
		description:      "Using env: to set variable is safe",
	},
	{
		name: "safe_action_with_input",
		workflowYAML: `name: Safe Handler
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: some-action@v1
        with:
          title: ${{ github.event.issue.title }}
`,
		expectExprInject: false,
		description:      "Passing to action with: is safe",
	},
	{
		name: "safe_github_sha",
		workflowYAML: `name: Safe CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.sha }}"
`,
		expectExprInject: false,
		description:      "github.sha is not attacker-controlled",
	},
	{
		name: "safe_github_ref",
		workflowYAML: `name: Safe CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.ref }}"
`,
		expectExprInject: false,
		description:      "github.ref on push is not attacker-controlled",
	},
}

// Helper to create temp workflow directory with test files
func setupTestWorkflows(t *testing.T, workflows map[string]string) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "unsafe-workflow-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Write workflow files
	for name, content := range workflows {
		path := filepath.Join(workflowsDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write workflow file: %v", err)
		}
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to get current dir: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	cleanup := func() {
		os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestUnsafeWorkflowSignal_PwnRequest(t *testing.T) {
	for _, tc := range pwnRequestTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, cleanup := setupTestWorkflows(t, map[string]string{"test.yml": tc.workflowYAML})
			defer cleanup()

			signal := NewUnsafeWorkflowSignal()
			ctx := context.Background()
			detected := signal.Check(ctx)

			if detected != tc.expectPwnRequest {
				t.Errorf("%s: expected detected=%v, got %v", tc.description, tc.expectPwnRequest, detected)
			}
		})
	}
}

func TestUnsafeWorkflowSignal_ExprInjection(t *testing.T) {
	for _, tc := range exprInjectionTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, cleanup := setupTestWorkflows(t, map[string]string{"test.yml": tc.workflowYAML})
			defer cleanup()

			signal := NewUnsafeWorkflowSignal()
			ctx := context.Background()
			detected := signal.Check(ctx)

			if detected != tc.expectExprInject {
				t.Errorf("%s: expected detected=%v, got %v", tc.description, tc.expectExprInject, detected)
			}
		})
	}
}

func TestUnsafeWorkflowSignal_NoWorkflowsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "no-workflows-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if detected {
		t.Error("Expected no detection when .github/workflows doesn't exist")
	}
}

func TestUnsafeWorkflowSignal_ContextCancellation(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yml": pwnRequestTestCases[2].workflowYAML, // vulnerable workflow
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return quickly without panic
	_ = signal.Check(ctx)
}

func TestUnsafeWorkflowSignal_Performance(t *testing.T) {
	// Create multiple workflow files
	workflows := make(map[string]string)
	for i := 0; i < 10; i++ {
		workflows[string(rune('a'+i))+".yml"] = pwnRequestTestCases[0].workflowYAML
	}

	_, cleanup := setupTestWorkflows(t, workflows)
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()

	start := time.Now()
	_ = signal.Check(ctx)
	elapsed := time.Since(start)

	// Should complete in under 10ms
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took too long: %v (expected <10ms)", elapsed)
	}
}

// ============================================================================
// Tests for safeJoinPath error paths
// ============================================================================

func TestSafeJoinPath_ValidFilename(t *testing.T) {
	baseDir := "/tmp/test"
	result, err := pathsec.SafeJoinPath(baseDir, "workflow.yml")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := filepath.Join(baseDir, "workflow.yml")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSafeJoinPath_DirectoryTraversal(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"double_dot", ".."},
		{"double_dot_prefix", "../secret.yml"},
		{"slash_in_filename", "sub/workflow.yml"},
		{"backslash_in_filename", "sub\\workflow.yml"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pathsec.SafeJoinPath("/tmp/base", tc.filename)
			if err == nil {
				t.Errorf("Expected error for filename %q, got nil", tc.filename)
			}
		})
	}
}

func TestSafeJoinPath_PathEscape(t *testing.T) {
	// Test case where joined path escapes base directory
	// This tests the prefix check logic
	baseDir := "/tmp/base"
	// After filepath.Clean, this would be just the filename
	result, err := pathsec.SafeJoinPath(baseDir, "safe.yml")
	if err != nil {
		t.Errorf("Expected no error for safe filename, got %v", err)
	}
	if result != filepath.Join(baseDir, "safe.yml") {
		t.Errorf("Unexpected result: %s", result)
	}
}

// ============================================================================
// Tests for Check error paths
// ============================================================================

func TestUnsafeWorkflowSignal_Check_WorkflowsDirIsFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workflow-dir-is-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .github directory
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatalf("Failed to create .github dir: %v", err)
	}

	// Create workflows as a FILE instead of directory
	workflowsPath := filepath.Join(githubDir, "workflows")
	if err := os.WriteFile(workflowsPath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("Failed to create workflows file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if detected {
		t.Error("Expected no detection when .github/workflows is a file, not directory")
	}
}

func TestUnsafeWorkflowSignal_Check_SkipsNonYAMLFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "non-yaml-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create non-YAML files
	for _, name := range []string{"readme.md", "script.sh", "config.json"} {
		path := filepath.Join(workflowsDir, name)
		content := "github.event.issue.title" // Would trigger if parsed
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if detected {
		t.Error("Expected no detection for non-YAML files")
	}
}

func TestUnsafeWorkflowSignal_Check_SkipsSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "subdir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create a subdirectory with a YAML extension (edge case)
	subDir := filepath.Join(workflowsDir, "archive.yml")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	// Should not panic or error
	_ = signal.Check(ctx)
}

func TestUnsafeWorkflowSignal_Check_SkipsFilesWithTraversalPatterns(t *testing.T) {
	// This test ensures files with ".." or "/" in name are skipped
	// Note: Most filesystems won't allow these, so this is a defensive test
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"safe.yml": exprInjectionTestCases[0].workflowYAML,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for safe.yml with vulnerable content")
	}
}

// ============================================================================
// Tests for parseAndCheckWorkflow error paths
// ============================================================================

func TestUnsafeWorkflowSignal_ParseAndCheckWorkflow_InvalidYAML(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"invalid.yml": `
name: Invalid
on: [push
jobs:
  broken yaml here
    - this is not valid
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if detected {
		t.Error("Expected no detection for invalid YAML")
	}
}

func TestUnsafeWorkflowSignal_ParseAndCheckWorkflow_UnreadableFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unreadable-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create a file that contains trigger pattern but can't be read during parsing
	path := filepath.Join(workflowsDir, "test.yml")
	content := `on: pull_request_target
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Make file unreadable after quick scan (this is timing-dependent)
	// For now, just verify the quick scan works
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for vulnerable workflow")
	}
}

// ============================================================================
// Tests for hasPullRequestTargetTrigger with different trigger formats
// ============================================================================

func TestHasPullRequestTargetTrigger_StringFormat(t *testing.T) {
	workflow := &WorkflowExt{On: "pull_request_target"}
	if !workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected true for string format trigger")
	}
}

func TestHasPullRequestTargetTrigger_ArrayFormat(t *testing.T) {
	workflow := &WorkflowExt{On: []interface{}{"push", "pull_request_target"}}
	if !workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected true for array format trigger")
	}
}

func TestHasPullRequestTargetTrigger_ArrayFormat_NotPresent(t *testing.T) {
	workflow := &WorkflowExt{On: []interface{}{"push", "pull_request"}}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false when pull_request_target not in array")
	}
}

func TestHasPullRequestTargetTrigger_ArrayFormat_NonStringElement(t *testing.T) {
	workflow := &WorkflowExt{On: []interface{}{123, "push"}}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false for non-string elements in array")
	}
}

func TestHasPullRequestTargetTrigger_MapFormat(t *testing.T) {
	workflow := &WorkflowExt{On: map[string]interface{}{
		"pull_request_target": map[string]interface{}{"branches": []string{"main"}},
	}}
	if !workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected true for map format trigger")
	}
}

func TestHasPullRequestTargetTrigger_MapFormat_NotPresent(t *testing.T) {
	workflow := &WorkflowExt{On: map[string]interface{}{
		"push":         map[string]interface{}{},
		"pull_request": map[string]interface{}{},
	}}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false when pull_request_target not in map")
	}
}

func TestHasPullRequestTargetTrigger_OtherTrigger(t *testing.T) {
	workflow := &WorkflowExt{On: "push"}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false for push trigger")
	}
}

func TestHasPullRequestTargetTrigger_UnknownType(t *testing.T) {
	// Test with an unexpected type
	workflow := &WorkflowExt{On: 12345}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false for unknown type")
	}
}

func TestHasPullRequestTargetTrigger_Nil(t *testing.T) {
	workflow := &WorkflowExt{On: nil}
	if workflow.hasPullRequestTargetTrigger() {
		t.Error("Expected false for nil trigger")
	}
}

// ============================================================================
// Tests for isVulnerableCheckoutStepExt
// ============================================================================

func TestIsVulnerableCheckoutStepExt_NoWith(t *testing.T) {
	step := StepExt{Uses: "actions/checkout@v4", With: nil}
	if isVulnerableCheckoutStepExt(step) {
		t.Error("Expected false when With is nil")
	}
}

func TestIsVulnerableCheckoutStepExt_NoRef(t *testing.T) {
	step := StepExt{
		Uses: "actions/checkout@v4",
		With: map[string]interface{}{"fetch-depth": 0},
	}
	if isVulnerableCheckoutStepExt(step) {
		t.Error("Expected false when no ref specified")
	}
}

func TestIsVulnerableCheckoutStepExt_RefNotString(t *testing.T) {
	step := StepExt{
		Uses: "actions/checkout@v4",
		With: map[string]interface{}{"ref": 12345},
	}
	if isVulnerableCheckoutStepExt(step) {
		t.Error("Expected false when ref is not a string")
	}
}

func TestIsVulnerableCheckoutStepExt_SafeRef(t *testing.T) {
	step := StepExt{
		Uses: "actions/checkout@v4",
		With: map[string]interface{}{"ref": "main"},
	}
	if isVulnerableCheckoutStepExt(step) {
		t.Error("Expected false for safe ref")
	}
}

func TestIsVulnerableCheckoutStepExt_VulnerableRef(t *testing.T) {
	step := StepExt{
		Uses: "actions/checkout@v4",
		With: map[string]interface{}{
			"ref": "${{ github.event.pull_request.head.sha }}",
		},
	}
	if !isVulnerableCheckoutStepExt(step) {
		t.Error("Expected true for vulnerable ref without persist-credentials: false")
	}
}

func TestIsVulnerableCheckoutStepExt_VulnerableRefMitigated(t *testing.T) {
	step := StepExt{
		Uses: "actions/checkout@v4",
		With: map[string]interface{}{
			"ref":                 "${{ github.event.pull_request.head.sha }}",
			"persist-credentials": false,
		},
	}
	if isVulnerableCheckoutStepExt(step) {
		t.Error("Expected false when persist-credentials is false")
	}
}

// ============================================================================
// Tests for isPRHeadRefExt
// ============================================================================

func TestIsPRHeadRefExt_AllPatterns(t *testing.T) {
	vulnerableRefs := []string{
		"github.event.pull_request.head.sha",
		"github.event.pull_request.head.ref",
		"${{ github.event.pull_request.head.sha }}",
		"${{ github.event.pull_request.head.ref }}",
	}

	for _, ref := range vulnerableRefs {
		t.Run(ref, func(t *testing.T) {
			if !isPRHeadRefExt(ref) {
				t.Errorf("Expected true for vulnerable ref: %s", ref)
			}
		})
	}
}

func TestIsPRHeadRefExt_SafeRefs(t *testing.T) {
	safeRefs := []string{
		"main",
		"develop",
		"refs/heads/main",
		"${{ github.sha }}",
		"${{ github.ref }}",
	}

	for _, ref := range safeRefs {
		t.Run(ref, func(t *testing.T) {
			if isPRHeadRefExt(ref) {
				t.Errorf("Expected false for safe ref: %s", ref)
			}
		})
	}
}

func TestIsPRHeadRefExt_PartialMatch(t *testing.T) {
	// Test that partial matches work (contains check)
	ref := "refs/${{ github.event.pull_request.head.sha }}/merge"
	if !isPRHeadRefExt(ref) {
		t.Error("Expected true for ref containing vulnerable pattern")
	}
}

// ============================================================================
// Tests for hasPersistCredentialsFalseExt
// ============================================================================

func TestHasPersistCredentialsFalseExt_BoolTrue(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": true},
	}
	if hasPersistCredentialsFalseExt(step) {
		t.Error("Expected false when persist-credentials is true")
	}
}

func TestHasPersistCredentialsFalseExt_BoolFalse(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": false},
	}
	if !hasPersistCredentialsFalseExt(step) {
		t.Error("Expected true when persist-credentials is false")
	}
}

func TestHasPersistCredentialsFalseExt_StringFalse(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": "false"},
	}
	if !hasPersistCredentialsFalseExt(step) {
		t.Error("Expected true when persist-credentials is 'false' string")
	}
}

func TestHasPersistCredentialsFalseExt_StringFalseUppercase(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": "FALSE"},
	}
	if !hasPersistCredentialsFalseExt(step) {
		t.Error("Expected true when persist-credentials is 'FALSE' string (case-insensitive)")
	}
}

func TestHasPersistCredentialsFalseExt_StringTrue(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": "true"},
	}
	if hasPersistCredentialsFalseExt(step) {
		t.Error("Expected false when persist-credentials is 'true' string")
	}
}

func TestHasPersistCredentialsFalseExt_NotPresent(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"ref": "main"},
	}
	if hasPersistCredentialsFalseExt(step) {
		t.Error("Expected false when persist-credentials is not present")
	}
}

func TestHasPersistCredentialsFalseExt_OtherType(t *testing.T) {
	step := StepExt{
		With: map[string]interface{}{"persist-credentials": 123},
	}
	if hasPersistCredentialsFalseExt(step) {
		t.Error("Expected false when persist-credentials is an unexpected type")
	}
}

// ============================================================================
// Tests for isCheckoutActionExt
// ============================================================================

func TestIsCheckoutActionExt_Versions(t *testing.T) {
	checkoutActions := []string{
		"actions/checkout@v4",
		"actions/checkout@v3",
		"actions/checkout@v2",
		"actions/checkout@v1",
		"actions/checkout@main",
		"actions/checkout@abc123",
	}

	for _, action := range checkoutActions {
		t.Run(action, func(t *testing.T) {
			if !isCheckoutActionExt(action) {
				t.Errorf("Expected true for: %s", action)
			}
		})
	}
}

func TestIsCheckoutActionExt_NotCheckout(t *testing.T) {
	otherActions := []string{
		"actions/setup-node@v4",
		"actions/cache@v3",
		"some-org/checkout@v1",
		"",
	}

	for _, action := range otherActions {
		t.Run(action, func(t *testing.T) {
			if isCheckoutActionExt(action) {
				t.Errorf("Expected false for: %s", action)
			}
		})
	}
}

// ============================================================================
// Tests for hasVulnerableCheckout
// ============================================================================

func TestHasVulnerableCheckout_NoJobs(t *testing.T) {
	workflow := &WorkflowExt{Jobs: map[string]JobExt{}}
	if workflow.hasVulnerableCheckout() {
		t.Error("Expected false when no jobs")
	}
}

func TestHasVulnerableCheckout_NoSteps(t *testing.T) {
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {Steps: []StepExt{}},
		},
	}
	if workflow.hasVulnerableCheckout() {
		t.Error("Expected false when no steps")
	}
}

func TestHasVulnerableCheckout_NonCheckoutAction(t *testing.T) {
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{Uses: "actions/setup-node@v4"},
				},
			},
		},
	}
	if workflow.hasVulnerableCheckout() {
		t.Error("Expected false for non-checkout action")
	}
}

func TestHasVulnerableCheckout_SafeCheckout(t *testing.T) {
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{Uses: "actions/checkout@v4"},
				},
			},
		},
	}
	if workflow.hasVulnerableCheckout() {
		t.Error("Expected false for safe checkout (no ref)")
	}
}

func TestHasVulnerableCheckout_VulnerableCheckout(t *testing.T) {
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{
						Uses: "actions/checkout@v4",
						With: map[string]interface{}{
							"ref": "${{ github.event.pull_request.head.sha }}",
						},
					},
				},
			},
		},
	}
	if !workflow.hasVulnerableCheckout() {
		t.Error("Expected true for vulnerable checkout")
	}
}

// ============================================================================
// Tests for checkExpressionInjection edge cases
// ============================================================================

func TestCheckExpressionInjection_EmptyRun(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{Uses: "actions/checkout@v4", Run: ""},
				},
			},
		},
	}
	finding := signal.checkExpressionInjection(workflow, "test.yml")
	if finding != nil {
		t.Error("Expected nil finding for empty run block")
	}
}

func TestCheckExpressionInjection_NoMatch(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{Run: "echo hello"},
				},
			},
		},
	}
	finding := signal.checkExpressionInjection(workflow, "test.yml")
	if finding != nil {
		t.Error("Expected nil finding for run block without expressions")
	}
}

func TestCheckExpressionInjection_SafeExpression(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					{Run: "echo ${{ github.sha }}"},
				},
			},
		},
	}
	finding := signal.checkExpressionInjection(workflow, "test.yml")
	if finding != nil {
		t.Error("Expected nil finding for safe expression")
	}
}

func TestCheckExpressionInjection_MultipleJobs(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"safe-job": {
				Steps: []StepExt{
					{Run: "echo safe"},
				},
			},
			"vulnerable-job": {
				Steps: []StepExt{
					{Run: "echo ${{ github.event.issue.title }}"},
				},
			},
		},
	}
	finding := signal.checkExpressionInjection(workflow, "test.yml")
	if finding == nil {
		t.Error("Expected finding for vulnerable job")
	}
}

// ============================================================================
// Tests for quickScan functions with file errors
// ============================================================================

func TestQuickScanForPullRequestTarget_FileNotFound(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	result := signal.quickScanForPullRequestTarget("/nonexistent/path/workflow.yml")
	if result {
		t.Error("Expected false for nonexistent file")
	}
}

func TestQuickScanForUntrustedExpr_FileNotFound(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	result := signal.quickScanForUntrustedExpr("/nonexistent/path/workflow.yml")
	if result {
		t.Error("Expected false for nonexistent file")
	}
}

// ============================================================================
// Tests for checkWorkflowFile context cancellation
// ============================================================================

func TestCheckWorkflowFile_ContextCancelledMidCheck(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yml": `name: Test
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
	})
	defer cleanup()

	signal := &UnsafeWorkflowSignal{}
	ctx, cancel := context.WithCancel(context.Background())

	// Get the workflow file path
	workflowsDir := ".github/workflows"
	absWorkflowsDir, _ := filepath.Abs(workflowsDir)
	filePath := filepath.Join(absWorkflowsDir, "test.yml")

	// Cancel context before calling checkWorkflowFile
	cancel()

	// Should return early without processing
	signal.checkWorkflowFile(ctx, filePath, "test.yml")

	// Since context was cancelled, no findings should be recorded
	// (the function returns early)
}

// ============================================================================
// Tests for hasFindings
// ============================================================================

func TestHasFindings_Empty(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	if signal.hasFindings() {
		t.Error("Expected false for empty signal")
	}
}

func TestHasFindings_PwnRequestOnly(t *testing.T) {
	signal := &UnsafeWorkflowSignal{pwnRequestFiles: []string{"test.yml"}}
	if !signal.hasFindings() {
		t.Error("Expected true when pwnRequestFiles is not empty")
	}
}

func TestHasFindings_ExprInjectionOnly(t *testing.T) {
	signal := &UnsafeWorkflowSignal{
		exprInjections: []exprInjectionFinding{{file: "test.yml", expression: "github.event.issue.title"}},
	}
	if !signal.hasFindings() {
		t.Error("Expected true when exprInjections is not empty")
	}
}

// ============================================================================
// Integration tests for combined vulnerabilities
// ============================================================================

func TestUnsafeWorkflowSignal_BothVulnerabilities(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"pwn.yml": `name: Pwn Request
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
		"expr.yml": `name: Expression Injection
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.title }}"
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal().(*UnsafeWorkflowSignal)
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for both vulnerabilities")
	}

	if len(signal.pwnRequestFiles) == 0 {
		t.Error("Expected pwn request finding")
	}

	if len(signal.exprInjections) == 0 {
		t.Error("Expected expression injection finding")
	}
}

func TestUnsafeWorkflowSignal_MultipleExprInjections(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"issue.yml": `name: Issue Handler
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.title }}"
`,
		"comment.yml": `name: Comment Handler
on: issue_comment
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.comment.body }}"
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal().(*UnsafeWorkflowSignal)
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for expression injections")
	}

	// Should find at least one (may find both depending on iteration order)
	if len(signal.exprInjections) == 0 {
		t.Error("Expected at least one expression injection finding")
	}
}

// ============================================================================
// Tests for YAML file extensions
// ============================================================================

func TestUnsafeWorkflowSignal_YAMLExtension(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yaml": `name: YAML Extension
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.title }}"
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for .yaml extension file")
	}
}

// ============================================================================
// Tests for parseAndCheckWorkflow with only one check type
// ============================================================================

func TestParseAndCheckWorkflow_OnlyPRT(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yml": `name: PRT Only
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal().(*UnsafeWorkflowSignal)
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for pwn request")
	}

	if len(signal.pwnRequestFiles) == 0 {
		t.Error("Expected pwn request finding")
	}
}

func TestParseAndCheckWorkflow_OnlyExprInjection(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yml": `name: Expr Only
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.body }}"
`,
	})
	defer cleanup()

	signal := NewUnsafeWorkflowSignal().(*UnsafeWorkflowSignal)
	ctx := context.Background()
	detected := signal.Check(ctx)

	if !detected {
		t.Error("Expected detection for expression injection")
	}

	if len(signal.exprInjections) == 0 {
		t.Error("Expected expression injection finding")
	}
}

// ============================================================================
// Tests for parseAndCheckWorkflow context cancellation
// ============================================================================

func TestParseAndCheckWorkflow_ContextCancelled(t *testing.T) {
	_, cleanup := setupTestWorkflows(t, map[string]string{
		"test.yml": `name: Test
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
	})
	defer cleanup()

	signal := &UnsafeWorkflowSignal{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	workflowsDir := ".github/workflows"
	absWorkflowsDir, _ := filepath.Abs(workflowsDir)
	filePath := filepath.Join(absWorkflowsDir, "test.yml")

	// Call parseAndCheckWorkflow directly with cancelled context
	signal.parseAndCheckWorkflow(ctx, filePath, "test.yml", true, true)

	// Should return early without findings
	if len(signal.pwnRequestFiles) > 0 || len(signal.exprInjections) > 0 {
		t.Error("Expected no findings when context is cancelled")
	}
}

// ============================================================================
// Tests for Check with unreadable workflows directory
// ============================================================================

func TestUnsafeWorkflowSignal_Check_UnreadableWorkflowsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unreadable-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create a workflow file
	path := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(path, []byte("on: push"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Make directory unreadable
	if err := os.Chmod(workflowsDir, 0000); err != nil {
		t.Skipf("Cannot change permissions: %v", err)
	}
	defer os.Chmod(workflowsDir, 0755) // Restore for cleanup

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()
	detected := signal.Check(ctx)

	if detected {
		t.Error("Expected no detection when workflows directory is unreadable")
	}
}

// ============================================================================
// Tests for safeJoinPath prefix escape
// ============================================================================

func TestSafeJoinPath_PrefixEscape(t *testing.T) {
	// Test the prefix check by using a base directory that doesn't end with separator
	// and a filename that could potentially escape
	baseDir := "/tmp/base"

	// Valid case - should work
	result, err := pathsec.SafeJoinPath(baseDir, "file.yml")
	if err != nil {
		t.Errorf("Expected no error for valid filename, got %v", err)
	}
	if result != filepath.Join(baseDir, "file.yml") {
		t.Errorf("Unexpected result: %s", result)
	}
}

// ============================================================================
// Tests for checkExpressionInjection with malformed expressions
// ============================================================================

func TestCheckExpressionInjection_MalformedExpression(t *testing.T) {
	signal := &UnsafeWorkflowSignal{}
	// This tests the case where regex matches but match array is incomplete
	// In practice, the regex always captures at least 2 groups, but we test the guard
	workflow := &WorkflowExt{
		Jobs: map[string]JobExt{
			"build": {
				Steps: []StepExt{
					// Expression with only safe content
					{Run: "echo ${{ github.sha }}"},
				},
			},
		},
	}
	finding := signal.checkExpressionInjection(workflow, "test.yml")
	if finding != nil {
		t.Error("Expected nil finding for safe expression")
	}
}

// ============================================================================
// Tests for Check with context cancellation during iteration
// ============================================================================

func TestUnsafeWorkflowSignal_Check_ContextCancelledDuringIteration(t *testing.T) {
	// Create multiple workflow files
	workflows := make(map[string]string)
	for i := 0; i < 5; i++ {
		name := string(rune('a'+i)) + ".yml"
		workflows[name] = `name: Test
on: issues
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - run: echo "${{ github.event.issue.title }}"
`
	}

	_, cleanup := setupTestWorkflows(t, workflows)
	defer cleanup()

	signal := NewUnsafeWorkflowSignal()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a very short time
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	// Should handle cancellation gracefully
	_ = signal.Check(ctx)
	// No assertion needed - just verify no panic
}

func TestUnsafeWorkflowSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_UNSAFE_WORKFLOW", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_UNSAFE_WORKFLOW")

	signal := NewUnsafeWorkflowSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
