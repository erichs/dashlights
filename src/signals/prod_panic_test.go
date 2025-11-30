package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestProdPanicSignal_Name(t *testing.T) {
	signal := NewProdPanicSignal()
	if signal.Name() != "Prod Panic" {
		t.Errorf("Expected 'Prod Panic', got '%s'", signal.Name())
	}
}

func TestProdPanicSignal_Emoji(t *testing.T) {
	signal := NewProdPanicSignal()
	if signal.Emoji() != "🚨" {
		t.Errorf("Expected '🚨', got '%s'", signal.Emoji())
	}
}

func TestProdPanicSignal_Diagnostic(t *testing.T) {
	signal := NewProdPanicSignal()
	signal.context = "production"
	signal.source = "AWS_PROFILE"
	expected := "Production context detected: production (AWS_PROFILE)"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestProdPanicSignal_Remediation(t *testing.T) {
	signal := NewProdPanicSignal()
	expected := "Switch to non-production context before running commands"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestProdPanicSignal_Check_NoProduction(t *testing.T) {
	// Save and restore env var
	oldProfile := os.Getenv("AWS_PROFILE")
	defer func() {
		if oldProfile != "" {
			os.Setenv("AWS_PROFILE", oldProfile)
		} else {
			os.Unsetenv("AWS_PROFILE")
		}
	}()

	os.Setenv("AWS_PROFILE", "dev")

	signal := NewProdPanicSignal()
	ctx := context.Background()

	// May return true if kubectl context is production
	// Just verify it doesn't panic
	signal.Check(ctx)
}

func TestProdPanicSignal_Check_AWSProfileProduction(t *testing.T) {
	// Save and restore env var
	oldProfile := os.Getenv("AWS_PROFILE")
	defer func() {
		if oldProfile != "" {
			os.Setenv("AWS_PROFILE", oldProfile)
		} else {
			os.Unsetenv("AWS_PROFILE")
		}
	}()

	os.Setenv("AWS_PROFILE", "production")

	signal := NewProdPanicSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when AWS_PROFILE is production")
	}

	if signal.context != "production" {
		t.Errorf("Expected context 'production', got '%s'", signal.context)
	}

	if signal.source != "AWS_PROFILE" {
		t.Errorf("Expected source 'AWS_PROFILE', got '%s'", signal.source)
	}
}

func TestProdPanicSignal_isProdIndicator(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"dev", false},
		{"staging", false},
		{"test", false},
		{"prod", true},
		{"production", true},
		{"live", true},
		{"prd", true},
		{"my-prod-cluster", true},
		{"production-us-east-1", true},
		{"PRODUCTION", true},
		{"Prod", true},
		{"development", false},
	}

	for _, tt := range tests {
		result := isProdIndicator(tt.value)
		if result != tt.expected {
			t.Errorf("isProdIndicator(%q) = %v, expected %v", tt.value, result, tt.expected)
		}
	}
}

func TestProdPanicSignal_checkKubeContext_NoKubeConfig(t *testing.T) {
	// Create a temp directory without .kube/config
	tmpDir := t.TempDir()

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewProdPanicSignal()

	if signal.checkKubeContext() {
		t.Error("Expected false when .kube/config doesn't exist")
	}
}

func TestProdPanicSignal_checkKubeContext_ProductionContext(t *testing.T) {
	// Create a temp directory with .kube/config
	tmpDir := t.TempDir()
	kubeDir := filepath.Join(tmpDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatal(err)
	}

	kubeConfig := filepath.Join(kubeDir, "config")
	content := `apiVersion: v1
clusters:
- cluster:
    server: https://example.com
  name: production
contexts:
- context:
    cluster: production
    user: admin
  name: production
current-context: production
kind: Config
`
	if err := os.WriteFile(kubeConfig, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore HOME
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)

	signal := NewProdPanicSignal()

	if !signal.checkKubeContext() {
		t.Error("Expected true when current-context is production")
	}

	if signal.context != "production" {
		t.Errorf("Expected context 'production', got '%s'", signal.context)
	}

	if signal.source != "kubectl context" {
		t.Errorf("Expected source 'kubectl context', got '%s'", signal.source)
	}
}

func TestProdPanicSignal_checkKubeContext_DirectoryTraversalPrevention(t *testing.T) {
	// Test that the signal rejects home directories with ".." (directory traversal)
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Try to set a malicious HOME path with directory traversal
	maliciousPath := "/tmp/../etc"
	os.Setenv("HOME", maliciousPath)

	signal := NewProdPanicSignal()

	// The check should return false (reject the malicious path)
	result := signal.checkKubeContext()
	if result {
		t.Error("Expected false when home directory contains '..' (directory traversal attempt)")
	}
}

func TestProdPanicSignal_checkKubeContext_RelativePathPrevention(t *testing.T) {
	// Test that the signal rejects relative home directories
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Try to set a relative HOME path
	relativePath := "relative/path"
	os.Setenv("HOME", relativePath)

	signal := NewProdPanicSignal()

	// The check should return false (reject relative paths)
	result := signal.checkKubeContext()
	if result {
		t.Error("Expected false when home directory is relative (not absolute)")
	}
}

func TestProdPanicSignal_checkKubeContext_ValidAbsolutePath(t *testing.T) {
	// Test that valid absolute paths work correctly
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create .kube directory with production config
	kubeDir := filepath.Join(tmpDir, ".kube")
	os.MkdirAll(kubeDir, 0755)

	kubeConfig := filepath.Join(kubeDir, "config")
	content := `apiVersion: v1
current-context: production
kind: Config
`
	err := os.WriteFile(kubeConfig, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create kube config: %v", err)
	}

	signal := NewProdPanicSignal()

	// Should work with valid absolute path
	result := signal.checkKubeContext()
	if !result {
		t.Error("Expected true with valid absolute path and production context")
	}
}
