package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPwnRequestSignal_Name(t *testing.T) {
	signal := NewPwnRequestSignal()
	if signal.Name() != "Pwn Request Risk" {
		t.Errorf("Expected 'Pwn Request Risk', got '%s'", signal.Name())
	}
}

func TestPwnRequestSignal_Emoji(t *testing.T) {
	signal := NewPwnRequestSignal()
	if signal.Emoji() != "🎣" {
		t.Errorf("Expected '🎣', got '%s'", signal.Emoji())
	}
}

func TestPwnRequestSignal_Diagnostic_Default(t *testing.T) {
	signal := NewPwnRequestSignal()
	expected := "GitHub Actions workflow contains pull_request_target with unsafe checkout"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestPwnRequestSignal_Diagnostic_WithFile(t *testing.T) {
	signal := &PwnRequestSignal{vulnerableFiles: []string{"ci.yml"}}
	expected := "GitHub Actions workflow vulnerable to pwn request: ci.yml"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestPwnRequestSignal_Remediation(t *testing.T) {
	signal := NewPwnRequestSignal()
	expected := "Use pull_request trigger instead, or add persist-credentials: false and avoid checking out PR head"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

// Test cases for the Check method
type pwnRequestTestCase struct {
	name           string
	workflowYAML   string
	expectedResult bool
	description    string
}

var pwnRequestTestCases = []pwnRequestTestCase{
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
		expectedResult: false,
		description:    "pull_request trigger is safe",
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
		expectedResult: false,
		description:    "pull_request_target without checkout is safe",
	},
	{
		name: "safe_pull_request_target_default_checkout",
		workflowYAML: `name: Safe Target CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "safe"
`,
		expectedResult: false,
		description:    "pull_request_target with default checkout (no ref) is safe",
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
		expectedResult: true,
		description:    "pull_request_target with PR head checkout is vulnerable",
	},
	{
		name: "vulnerable_pr_head_ref_checkout",
		workflowYAML: `name: Vulnerable CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.ref }}
      - run: npm test
`,
		expectedResult: true,
		description:    "pull_request_target with PR head ref checkout is vulnerable",
	},
	{
		name: "safe_with_persist_credentials_false",
		workflowYAML: `name: Safe CI with persist-credentials
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: false
      - run: cat README.md
`,
		expectedResult: false,
		description:    "pull_request_target with persist-credentials: false is safe",
	},
	{
		name: "vulnerable_array_trigger",
		workflowYAML: `name: Multi-trigger
on: [push, pull_request_target]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
		expectedResult: true,
		description:    "pull_request_target in array with PR head checkout is vulnerable",
	},
	{
		name: "vulnerable_map_trigger",
		workflowYAML: `name: Map trigger
on:
  pull_request_target:
    types: [opened, synchronize]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
		expectedResult: true,
		description:    "pull_request_target as map key with PR head checkout is vulnerable",
	},
	{
		name: "safe_push_only",
		workflowYAML: `name: Push CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
		expectedResult: false,
		description:    "push trigger only is safe",
	},
}

func TestPwnRequestSignal_Check(t *testing.T) {
	for _, tc := range pwnRequestTestCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
			if err := os.MkdirAll(workflowsDir, 0755); err != nil {
				t.Fatal(err)
			}

			workflowFile := filepath.Join(workflowsDir, "test.yml")
			if err := os.WriteFile(workflowFile, []byte(tc.workflowYAML), 0644); err != nil {
				t.Fatal(err)
			}

			oldCwd, _ := os.Getwd()
			defer os.Chdir(oldCwd)
			os.Chdir(tmpDir)

			signal := NewPwnRequestSignal()
			ctx := context.Background()
			result := signal.Check(ctx)

			if result != tc.expectedResult {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expectedResult, result)
			}
		})
	}
}

func TestPwnRequestSignal_NoWorkflowsDir(t *testing.T) {
	tmpDir := t.TempDir()

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when .github/workflows directory does not exist")
	}
}

func TestPwnRequestSignal_EmptyWorkflowsDir(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when workflows directory is empty")
	}
}

func TestPwnRequestSignal_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create malformed YAML that contains trigger string but is invalid
	malformedYAML := `name: Broken
on: pull_request_target
jobs:
  build:
    runs-on: [
      - invalid yaml here
`
	workflowFile := filepath.Join(workflowsDir, "broken.yml")
	if err := os.WriteFile(workflowFile, []byte(malformedYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should not panic and should return false for malformed YAML
	if signal.Check(ctx) {
		t.Error("Expected false when YAML is malformed")
	}
}

func TestPwnRequestSignal_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid workflow file
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	workflowFile := filepath.Join(workflowsDir, "ci.yml")
	if err := os.WriteFile(workflowFile, []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return false due to cancelled context
	if signal.Check(ctx) {
		t.Error("Expected false when context is cancelled")
	}
}

func TestPwnRequestSignal_Performance(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create multiple workflow files
	for i := 0; i < 10; i++ {
		workflowYAML := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
		filename := filepath.Join(workflowsDir, "workflow"+string(rune('0'+i))+".yml")
		if err := os.WriteFile(filename, []byte(workflowYAML), 0644); err != nil {
			t.Fatal(err)
		}
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	start := time.Now()
	signal.Check(ctx)
	elapsed := time.Since(start)

	// Should complete well within 10ms budget
	if elapsed > 10*time.Millisecond {
		t.Errorf("Check took %v, expected < 10ms", elapsed)
	}
}

func TestPwnRequestSignal_NonYAMLFiles(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a non-YAML file that mentions pull_request_target
	content := `This is a README about pull_request_target vulnerabilities`
	if err := os.WriteFile(filepath.Join(workflowsDir, "README.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when only non-YAML files exist")
	}
}

func TestPwnRequestSignal_MultipleJobs(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with multiple jobs, only one vulnerable
	workflowYAML := `name: Multi-job
on: pull_request_target
jobs:
  safe-job:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/labeler@v4
  vulnerable-job:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "multi.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when at least one job has vulnerable checkout")
	}
}

func TestPwnRequestSignal_YAMLExtension(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Use .yaml extension instead of .yml
	workflowYAML := `name: YAML extension
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yaml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true for .yaml extension files")
	}
}

func TestPwnRequestSignal_SubdirectoryInWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory inside workflows (should be skipped)
	subDir := filepath.Join(workflowsDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid safe workflow
	workflowYAML := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - subdirectory should be skipped, safe workflow is safe
	if signal.Check(ctx) {
		t.Error("Expected false - subdirectories should be skipped")
	}
}

func TestPwnRequestSignal_UnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a workflow file with pull_request_target
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	workflowFile := filepath.Join(workflowsDir, "ci.yml")
	if err := os.WriteFile(workflowFile, []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Make file unreadable (only works if not running as root)
	if err := os.Chmod(workflowFile, 0000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions on cleanup
	defer os.Chmod(workflowFile, 0644)

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false because file can't be read
	// (quickScanForTrigger will fail to open the file)
	result := signal.Check(ctx)
	// On systems where we're root, the file may still be readable
	// so we just verify the check doesn't panic
	_ = result
}

func TestPwnRequestSignal_PersistCredentialsAsString(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test persist-credentials as string "false" (should be safe)
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: "false"
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false - persist-credentials: 'false' (string) should mitigate")
	}
}

func TestPwnRequestSignal_PersistCredentialsAsTrue(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test persist-credentials explicitly true (should be vulnerable)
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: true
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true - persist-credentials: true should be vulnerable")
	}
}

func TestPwnRequestSignal_RefAsNonString(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test ref as a number (unusual but valid YAML)
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: 12345
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - ref is not a string, not a vulnerable pattern
	if signal.Check(ctx) {
		t.Error("Expected false - non-string ref should not be detected as vulnerable")
	}
}

func TestPwnRequestSignal_NoJobs(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with pull_request_target but no jobs
	workflowYAML := `name: Empty
on: pull_request_target
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - no jobs means no vulnerable checkout
	if signal.Check(ctx) {
		t.Error("Expected false - workflow with no jobs can't be vulnerable")
	}
}

func TestPwnRequestSignal_NonCheckoutAction(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with pull_request_target but using a different action with ref
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - only checkout action is vulnerable
	if signal.Check(ctx) {
		t.Error("Expected false - only actions/checkout is vulnerable")
	}
}

func TestPwnRequestSignal_CheckoutWithoutWith(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with checkout but no 'with' block at all
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello"
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - no 'with' means default checkout which is safe
	if signal.Check(ctx) {
		t.Error("Expected false - checkout without 'with' is safe")
	}
}

func TestPwnRequestSignal_WorkflowIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	githubDir := filepath.Join(tmpDir, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .github/workflows as a file, not a directory
	workflowsFile := filepath.Join(githubDir, "workflows")
	if err := os.WriteFile(workflowsFile, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - workflows is not a directory
	if signal.Check(ctx) {
		t.Error("Expected false when .github/workflows is a file, not directory")
	}
}

func TestSafeJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		baseDir  string
		filename string
		wantErr  bool
	}{
		{"valid simple", "/base", "file.yml", false},
		{"valid nested name", "/base", "ci.workflow.yml", false},
		{"reject dotdot", "/base", "..", true},
		{"reject dotdot prefix", "/base", "../etc/passwd", true},
		{"reject slash", "/base", "sub/file.yml", true},
		{"reject backslash", "/base", "sub\\file.yml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safeJoinPath(tt.baseDir, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("safeJoinPath(%q, %q) error = %v, wantErr %v",
					tt.baseDir, tt.filename, err, tt.wantErr)
			}
		})
	}
}

func TestPwnRequestSignal_QuickScanNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow without pull_request_target - quick scan should return false
	workflowYAML := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - quick scan won't find pull_request_target
	if signal.Check(ctx) {
		t.Error("Expected false - no pull_request_target in file")
	}
}

func TestPwnRequestSignal_UnreadableDirAfterStat(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a workflow file
	workflowYAML := `name: CI
on: push
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Make directory unreadable after it exists (to trigger ReadDir error)
	if err := os.Chmod(workflowsDir, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(workflowsDir, 0755)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - can't read directory
	result := signal.Check(ctx)
	_ = result // Just verify no panic
}

func TestPwnRequestSignal_ContextCancelledDuringParse(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workflow that passes quick scan (has pull_request_target)
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()

	// Create a context that we cancel immediately
	// The context check in parseAndCheckWorkflow should catch it
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return false due to cancelled context
	if signal.Check(ctx) {
		t.Error("Expected false when context is cancelled")
	}
}

func TestPwnRequestSignal_HasPRTargetButNoVulnerableCheckout(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Workflow with pull_request_target but checkout has persist-credentials: false
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: false
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	signal := NewPwnRequestSignal()
	ctx := context.Background()

	// Should return false - has persist-credentials: false
	if signal.Check(ctx) {
		t.Error("Expected false - persist-credentials: false mitigates")
	}
}

func TestPwnRequestSignal_TriggerVariants(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected bool
	}{
		{
			name: "trigger as single event with types",
			yaml: `name: CI
on:
  pull_request_target:
    types: [opened]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
			expected: true,
		},
		{
			name: "trigger in array with other events",
			yaml: `name: CI
on: [push, pull_request_target]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
			expected: true,
		},
		{
			name: "trigger with workflow_dispatch",
			yaml: `name: CI
on:
  workflow_dispatch:
  pull_request_target:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
			if err := os.MkdirAll(workflowsDir, 0755); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(tt.yaml), 0644); err != nil {
				t.Fatal(err)
			}

			oldCwd, _ := os.Getwd()
			defer os.Chdir(oldCwd)
			os.Chdir(tmpDir)

			signal := NewPwnRequestSignal()
			ctx := context.Background()

			if signal.Check(ctx) != tt.expected {
				t.Errorf("Expected %v for %s", tt.expected, tt.name)
			}
		})
	}
}

func TestPwnRequestSignal_ParseAndCheckWorkflow_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "ci.yml")

	// Create a valid vulnerable workflow
	workflowYAML := `name: CI
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
`
	if err := os.WriteFile(workflowFile, []byte(workflowYAML), 0644); err != nil {
		t.Fatal(err)
	}

	signal := &PwnRequestSignal{}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Call parseAndCheckWorkflow directly with cancelled context
	result := signal.parseAndCheckWorkflow(ctx, workflowFile)

	// Should return false due to cancelled context
	if result {
		t.Error("Expected false when context is cancelled")
	}
}
