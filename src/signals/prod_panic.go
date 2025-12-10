package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/homedirutil"
)

// ProdPanicSignal checks for production context in kubectl or AWS
type ProdPanicSignal struct {
	context string
	source  string
}

// NewProdPanicSignal creates a ProdPanicSignal.
func NewProdPanicSignal() *ProdPanicSignal {
	return &ProdPanicSignal{}
}

// Name returns the human-readable name of the signal.
func (s *ProdPanicSignal) Name() string {
	return "Prod Panic"
}

// Emoji returns the emoji associated with the signal.
func (s *ProdPanicSignal) Emoji() string {
	return "🚨"
}

// Diagnostic returns details about the detected production context.
func (s *ProdPanicSignal) Diagnostic() string {
	return "Production context detected: " + s.context + " (" + s.source + ")"
}

// Remediation returns guidance on switching away from production contexts.
func (s *ProdPanicSignal) Remediation() string {
	return "Switch to non-production context before running commands"
}

// Check inspects AWS and Kubernetes configuration for production indicators.
func (s *ProdPanicSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_PROD_PANIC") != "" {
		return false
	}

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
	kubeConfig, err := homedirutil.SafeHomePath(".kube", "config")
	if err != nil {
		return false
	}

	// filepath.Clean for gosec G304 - path is already validated by SafeHomePath
	file, err := os.Open(filepath.Clean(kubeConfig))
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
