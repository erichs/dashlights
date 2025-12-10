package signals

import (
	"context"
	"os"
	"testing"
)

func TestCargoPathDepsSignal_NoCargo(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when Cargo.toml doesn't exist")
	}
}

func TestCargoPathDepsSignal_NoPathDeps(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml without path dependencies
	cargo := `[package]
name = "myproject"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when Cargo.toml has no path dependencies")
	}
}

func TestCargoPathDepsSignal_WithPathDep(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml with path dependency
	cargo := `[package]
name = "myproject"
version = "0.1.0"

[dependencies]
serde = "1.0"
mylib = { path = "../mylib" }
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when Cargo.toml has path dependency")
	}
}

func TestCargoPathDepsSignal_DevDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml with path dependency in dev-dependencies
	cargo := `[package]
name = "myproject"
version = "0.1.0"

[dev-dependencies]
test-utils = { path = "../test-utils" }
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when dev-dependencies has path dependency")
	}
}

func TestCargoPathDepsSignal_BuildDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml with path dependency in build-dependencies
	cargo := `[package]
name = "myproject"
version = "0.1.0"

[build-dependencies]
build-script = { path = "../build-script" }
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when build-dependencies has path dependency")
	}
}

func TestCargoPathDepsSignal_CommentedPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml with commented path dependency
	cargo := `[package]
name = "myproject"
version = "0.1.0"

[dependencies]
serde = "1.0"
# mylib = { path = "../mylib" }
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when path dependency is commented out")
	}
}

func TestCargoPathDepsSignal_PathInVersion(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create Cargo.toml where "path" appears in a different context
	cargo := `[package]
name = "myproject"
version = "0.1.0"

[dependencies]
serde = "1.0"
serde_json = "1.0"
`
	os.WriteFile("Cargo.toml", []byte(cargo), 0644)

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when 'path' doesn't indicate a path dependency")
	}
}

func TestCargoPathDepsSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_CARGO_PATH_DEPS", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_CARGO_PATH_DEPS")

	signal := NewCargoPathDepsSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
