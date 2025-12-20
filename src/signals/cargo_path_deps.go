package signals

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/fileutil"
)

// CargoPathDepsSignal checks for path dependencies in Cargo.toml
// These break builds on other machines (like Go replace directives)
type CargoPathDepsSignal struct {
	foundDep string
}

// NewCargoPathDepsSignal creates a CargoPathDepsSignal.
func NewCargoPathDepsSignal() Signal {
	return &CargoPathDepsSignal{}
}

// Name returns the human-readable name of the signal.
func (s *CargoPathDepsSignal) Name() string {
	return "Cargo Path Dependencies"
}

// Emoji returns the emoji associated with the signal.
func (s *CargoPathDepsSignal) Emoji() string {
	return "🦀" // Crab (Rust)
}

// Diagnostic returns a description of the detected path dependencies.
func (s *CargoPathDepsSignal) Diagnostic() string {
	if s.foundDep != "" {
		return "Cargo.toml contains path dependency: " + s.foundDep + " (breaks builds on other machines)"
	}
	return "Cargo.toml contains path dependencies (breaks builds on other machines)"
}

// Remediation returns guidance on how to replace path dependencies.
func (s *CargoPathDepsSignal) Remediation() string {
	return "Replace path dependencies with crates.io versions before committing"
}

// Check scans Cargo.toml for path dependencies that harm reproducibility.
func (s *CargoPathDepsSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_CARGO_PATH_DEPS") != "" {
		return false
	}

	// Check if Cargo.toml exists in current directory
	const maxCargoTomlBytes = 256 * 1024

	data, err := fileutil.ReadFileLimitedString("Cargo.toml", maxCargoTomlBytes)
	if err != nil {
		// No Cargo.toml file - not a Rust project
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(data))
	inDependenciesSection := false

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := strings.TrimSpace(scanner.Text())

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Track if we're in a dependencies section
		if strings.HasPrefix(line, "[dependencies") ||
			strings.HasPrefix(line, "[dev-dependencies") ||
			strings.HasPrefix(line, "[build-dependencies") {
			inDependenciesSection = true
			continue
		}

		// Exit dependencies section when we hit another section
		if strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[dependencies") &&
			!strings.HasPrefix(line, "[dev-dependencies") && !strings.HasPrefix(line, "[build-dependencies") {
			inDependenciesSection = false
			continue
		}

		// Check for path = "..." in dependencies
		if inDependenciesSection && strings.Contains(line, "path") && strings.Contains(line, "=") {
			// Extract the dependency name if possible
			if strings.Contains(line, "path") {
				s.foundDep = line
				return true
			}
		}
	}

	return false
}
