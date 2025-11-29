package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRootKubeContextSignal_NoKubeConfig(t *testing.T) {
	// Create temp home directory without .kube/config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewRootKubeContextSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no kube config exists")
	}
}

func TestRootKubeContextSignal_KubeSystemNamespace(t *testing.T) {
	// Create temp home directory with kube config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create .kube directory
	kubeDir := filepath.Join(tmpDir, ".kube")
	os.MkdirAll(kubeDir, 0755)

	// Create config with kube-system namespace
	kubeConfig := `apiVersion: v1
kind: Config
current-context: prod-cluster
contexts:
- context:
    cluster: prod-cluster
    namespace: kube-system
    user: admin
  name: prod-cluster
clusters:
- cluster:
    server: https://kubernetes.example.com
  name: prod-cluster
users:
- name: admin
  user:
    token: fake-token
`
	configPath := filepath.Join(kubeDir, "config")
	err := os.WriteFile(configPath, []byte(kubeConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create kube config: %v", err)
	}

	signal := NewRootKubeContextSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when namespace is kube-system")
	}
}

func TestRootKubeContextSignal_SafeNamespace(t *testing.T) {
	// Create temp home directory with kube config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create .kube directory
	kubeDir := filepath.Join(tmpDir, ".kube")
	os.MkdirAll(kubeDir, 0755)

	// Create config with safe namespace
	kubeConfig := `apiVersion: v1
kind: Config
current-context: dev-cluster
contexts:
- context:
    cluster: dev-cluster
    namespace: default
    user: developer
  name: dev-cluster
clusters:
- cluster:
    server: https://kubernetes.example.com
  name: dev-cluster
users:
- name: developer
  user:
    token: fake-token
`
	configPath := filepath.Join(kubeDir, "config")
	err := os.WriteFile(configPath, []byte(kubeConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create kube config: %v", err)
	}

	signal := NewRootKubeContextSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when namespace is not kube-system")
	}
}

func TestRootKubeContextSignal_NoNamespace(t *testing.T) {
	// Create temp home directory with kube config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create .kube directory
	kubeDir := filepath.Join(tmpDir, ".kube")
	os.MkdirAll(kubeDir, 0755)

	// Create config without namespace specified
	kubeConfig := `apiVersion: v1
kind: Config
current-context: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: tester
  name: test-cluster
clusters:
- cluster:
    server: https://kubernetes.example.com
  name: test-cluster
users:
- name: tester
  user:
    token: fake-token
`
	configPath := filepath.Join(kubeDir, "config")
	err := os.WriteFile(configPath, []byte(kubeConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create kube config: %v", err)
	}

	signal := NewRootKubeContextSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no namespace is specified")
	}
}

func TestRootKubeContextSignal_Metadata(t *testing.T) {
	signal := NewRootKubeContextSignal()

	if signal.Name() == "" {
		t.Error("Name should not be empty")
	}

	if signal.Emoji() == "" {
		t.Error("Emoji should not be empty")
	}

	if signal.Diagnostic() == "" {
		t.Error("Diagnostic should not be empty")
	}

	if signal.Remediation() == "" {
		t.Error("Remediation should not be empty")
	}
}
