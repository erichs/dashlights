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

func NewGoReplaceSignal() Signal {
	return &GoReplaceSignal{}
}

func (s *GoReplaceSignal) Name() string {
	return "Go Replace Directive"
}

func (s *GoReplaceSignal) Emoji() string {
	return "🔄" // Counterclockwise arrows (replace/swap)
}

func (s *GoReplaceSignal) Diagnostic() string {
	if s.foundReplace != "" {
		return "go.mod contains replace directive: " + s.foundReplace + " (breaks builds on other machines)"
	}
	return "go.mod contains replace directive (breaks builds on other machines)"
}

func (s *GoReplaceSignal) Remediation() string {
	return "Remove replace directives from go.mod before committing"
}

func (s *GoReplaceSignal) Check(ctx context.Context) bool {
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
