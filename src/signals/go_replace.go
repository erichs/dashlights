package signals

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// GoReplaceSignal checks for replace directives in go.mod
// These often indicate local debugging paths that break builds for others
type GoReplaceSignal struct {
	foundReplace string
}

// NewGoReplaceSignal creates a GoReplaceSignal.
func NewGoReplaceSignal() Signal {
	return &GoReplaceSignal{}
}

// Name returns the human-readable name of the signal.
func (s *GoReplaceSignal) Name() string {
	return "Go Replace Directive"
}

// Emoji returns the emoji associated with the signal.
func (s *GoReplaceSignal) Emoji() string {
	return "🔄" // Counterclockwise arrows (replace/swap)
}

// Diagnostic returns a description of the detected replace directive in go.mod.
func (s *GoReplaceSignal) Diagnostic() string {
	if s.foundReplace != "" {
		return "go.mod contains replace directive: " + s.foundReplace + " (breaks builds on other machines)"
	}
	return "go.mod contains replace directive (breaks builds on other machines)"
}

// Remediation returns guidance on how to remove or adjust replace directives.
func (s *GoReplaceSignal) Remediation() string {
	return "Remove replace directives from go.mod before committing"
}

// Check scans go.mod for replace directives that may break reproducible builds.
func (s *GoReplaceSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_GO_REPLACE") != "" {
		return false
	}

	// Check if go.mod exists in current directory
	file, err := os.Open("go.mod")
	if err != nil {
		// No go.mod file - not a Go project or not in project root
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		// Check for replace directive
		if strings.HasPrefix(line, "replace ") {
			// Extract the replace statement for diagnostic
			s.foundReplace = strings.TrimPrefix(line, "replace ")
			return true
		}
	}

	return false
}
