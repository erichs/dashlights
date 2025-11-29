package signals

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// CargoPathDepsSignal checks for path dependencies in Cargo.toml
// These break builds on other machines (like Go replace directives)
type CargoPathDepsSignal struct {
	foundDep string
}

func NewCargoPathDepsSignal() Signal {
	return &CargoPathDepsSignal{}
}

func (s *CargoPathDepsSignal) Name() string {
	return "Cargo Path Dependencies"
}

func (s *CargoPathDepsSignal) Emoji() string {
	return "🦀" // Crab (Rust)
}

func (s *CargoPathDepsSignal) Diagnostic() string {
	if s.foundDep != "" {
		return "Cargo.toml contains path dependency: " + s.foundDep + " (breaks builds on other machines)"
	}
	return "Cargo.toml contains path dependencies (breaks builds on other machines)"
}

func (s *CargoPathDepsSignal) Remediation() string {
	return "Replace path dependencies with crates.io versions before committing"
}

func (s *CargoPathDepsSignal) Check(ctx context.Context) bool {
	// Check if Cargo.toml exists in current directory
	file, err := os.Open("Cargo.toml")
	if err != nil {
		// No Cargo.toml file - not a Rust project
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDependenciesSection := false

	for scanner.Scan() {
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
