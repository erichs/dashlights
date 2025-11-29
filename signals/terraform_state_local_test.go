package signals

import (
	"context"
	"os"
	"testing"
)

func TestTerraformStateLocalSignal_NoStateFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	signal := NewTerraformStateLocalSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no terraform.tfstate exists")
	}
}

func TestTerraformStateLocalSignal_WithStateFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create terraform.tfstate
	err := os.WriteFile("terraform.tfstate", []byte(`{"version": 4}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	signal := NewTerraformStateLocalSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when terraform.tfstate exists")
	}
}

func TestTerraformStateLocalSignal_WithBackupFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create terraform.tfstate.backup (indicates recent local state usage)
	err := os.WriteFile("terraform.tfstate.backup", []byte(`{"version": 4}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	signal := NewTerraformStateLocalSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when terraform.tfstate.backup exists")
	}
}

func TestTerraformStateLocalSignal_WithBothFiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create both files
	os.WriteFile("terraform.tfstate", []byte(`{"version": 4}`), 0644)
	os.WriteFile("terraform.tfstate.backup", []byte(`{"version": 3}`), 0644)

	signal := NewTerraformStateLocalSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when both state files exist")
	}
}

func TestTerraformStateLocalSignal_WithRemoteBackend(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .terraform directory (indicates terraform init was run)
	// but no local state files (using remote backend)
	os.Mkdir(".terraform", 0755)

	signal := NewTerraformStateLocalSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when using remote backend (no local state)")
	}
}

func TestTerraformStateLocalSignal_Metadata(t *testing.T) {
	signal := NewTerraformStateLocalSignal()

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

