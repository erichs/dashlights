package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/homedirutil"
)

// RootKubeContextSignal checks if current Kubernetes context namespace is kube-system
// You rarely want to operate directly in the system namespace
type RootKubeContextSignal struct {
	contextName string
}

// NewRootKubeContextSignal creates a RootKubeContextSignal.
func NewRootKubeContextSignal() Signal {
	return &RootKubeContextSignal{}
}

// Name returns the human-readable name of the signal.
func (s *RootKubeContextSignal) Name() string {
	return "Root Kube Context"
}

// Emoji returns the emoji associated with the signal.
func (s *RootKubeContextSignal) Emoji() string {
	return "☸️" // Kubernetes wheel
}

// Diagnostic returns a description of the risky Kubernetes context.
func (s *RootKubeContextSignal) Diagnostic() string {
	if s.contextName != "" {
		return "Kubernetes context '" + s.contextName + "' uses kube-system namespace (dangerous for operations)"
	}
	return "Kubernetes context uses kube-system namespace (dangerous for operations)"
}

// Remediation returns guidance on switching to a safer Kubernetes namespace.
func (s *RootKubeContextSignal) Remediation() string {
	return "Switch to a non-system namespace with 'kubectl config set-context --current --namespace=<namespace>'"
}

// Check parses kubeconfig to see if the current context targets the kube-system namespace.
func (s *RootKubeContextSignal) Check(ctx context.Context) bool {
	kubeConfig, err := homedirutil.SafeHomePath(".kube", "config")
	if err != nil {
		return false
	}

	// filepath.Clean for gosec G304 - path is already validated by SafeHomePath
	file, err := os.Open(filepath.Clean(kubeConfig))
	if err != nil {
		// No kube config - not using Kubernetes
		return false
	}
	defer file.Close()

	// Parse YAML to find current context and its namespace
	currentContext := ""
	inContextSection := false
	inContextEntry := false
	contextName := ""
	contextNamespace := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Find current-context
		if strings.HasPrefix(trimmed, "current-context:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				currentContext = strings.TrimSpace(parts[1])
			}
		}

		// Find contexts section
		if trimmed == "contexts:" {
			inContextSection = true
			continue
		}

		// Exit contexts section when we hit another top-level key (not indented)
		if inContextSection && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '-' {
			inContextSection = false
			continue
		}

		if !inContextSection {
			continue
		}

		// Start of a new context entry (- context: or - name:)
		if strings.HasPrefix(trimmed, "- context:") || strings.HasPrefix(trimmed, "- name:") {
			inContextEntry = true
			contextNamespace = ""
		}

		// Capture namespace (inside "context:" block)
		if inContextEntry && strings.HasPrefix(trimmed, "namespace:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				contextNamespace = strings.TrimSpace(parts[1])
			}
		}

		// Find context name (at same level as "context:")
		if inContextEntry && strings.HasPrefix(trimmed, "name:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				contextName = strings.TrimSpace(parts[1])
				// Check if this is the current context and has kube-system namespace
				if contextName == currentContext && contextNamespace == "kube-system" {
					s.contextName = currentContext
					return true
				}
			}
		}
	}

	return false
}
