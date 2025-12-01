package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// ProdPanicSignal checks for production context in kubectl or AWS
type ProdPanicSignal struct {
	context string
	source  string
}

func NewProdPanicSignal() *ProdPanicSignal {
	return &ProdPanicSignal{}
}

func (s *ProdPanicSignal) Name() string {
	return "Prod Panic"
}

func (s *ProdPanicSignal) Emoji() string {
	return "🚨"
}

func (s *ProdPanicSignal) Diagnostic() string {
	return "Production context detected: " + s.context + " (" + s.source + ")"
}

func (s *ProdPanicSignal) Remediation() string {
	return "Switch to non-production context before running commands"
}

func (s *ProdPanicSignal) Check(ctx context.Context) bool {
	// Check AWS_PROFILE
	awsProfile := os.Getenv("AWS_PROFILE")
	if isProdIndicator(awsProfile) {
		s.context = awsProfile
		s.source = "AWS_PROFILE"
		return true
	}

	// Check kubectl context
	if s.checkKubeContext(ctx) {
		return true
	}

	return false
}

func (s *ProdPanicSignal) checkKubeContext(ctx context.Context) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Sanitize home directory to prevent directory traversal attacks
	// Validate that homeDir doesn't contain suspicious patterns
	if strings.Contains(homeDir, "..") {
		return false
	}

	// Clean the path to resolve any . or .. components
	sanitizedHome := filepath.Clean(homeDir)

	// Ensure the sanitized path is absolute (home directories should always be absolute)
	if !filepath.IsAbs(sanitizedHome) {
		return false
	}

	kubeConfig := filepath.Join(sanitizedHome, ".kube", "config")
	file, err := os.Open(kubeConfig)
	if err != nil {
		return false
	}
	defer file.Close()

	// Simple line-by-line parsing to find "current-context: <value>"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "current-context:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				contextName := strings.TrimSpace(parts[1])
				if isProdIndicator(contextName) {
					s.context = contextName
					s.source = "kubectl context"
					return true
				}
			}
			break
		}
	}

	return false
}

func isProdIndicator(value string) bool {
	if value == "" {
		return false
	}

	lower := strings.ToLower(value)
	prodIndicators := []string{"prod", "production", "live", "prd"}

	for _, indicator := range prodIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return false
}
