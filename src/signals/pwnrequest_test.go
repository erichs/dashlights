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
