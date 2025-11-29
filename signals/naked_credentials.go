package signals

import (
	"context"
	"os"
	"strings"
)

// NakedCredentialsSignal checks for raw secrets in environment variables
type NakedCredentialsSignal struct {
	foundVars []string
}

func NewNakedCredentialsSignal() *NakedCredentialsSignal {
	return &NakedCredentialsSignal{}
}

func (s *NakedCredentialsSignal) Name() string {
	return "Naked Credential"
}

func (s *NakedCredentialsSignal) Emoji() string {
	return "🩲"
}

func (s *NakedCredentialsSignal) Diagnostic() string {
	if len(s.foundVars) == 0 {
		return "Raw secrets detected in environment variables"
	}
	return "Raw secrets in environment: " + strings.Join(s.foundVars, ", ")
}

func (s *NakedCredentialsSignal) Remediation() string {
	return "Use credential helpers, keychains, or secret management tools instead"
}

func (s *NakedCredentialsSignal) Check(ctx context.Context) bool {
	s.foundVars = []string{}

	// Patterns that indicate secrets
	secretPatterns := []string{
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"GOOGLE_APPLICATION_CREDENTIALS",
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"DOCKER_PASSWORD",
		"NPM_TOKEN",
		"SLACK_TOKEN",
		"STRIPE_SECRET_KEY",
		"TWILIO_AUTH_TOKEN",
	}

	// Suffix patterns
	secretSuffixes := []string{
		"_TOKEN",
		"_SECRET",
		"_KEY",
		"_PASSWORD",
		"_APIKEY",
		"_API_KEY",
	}

	environ := os.Environ()
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		varName := parts[0]
		varValue := parts[1]

		// Skip empty values
		if varValue == "" {
			continue
		}

		// Skip DASHLIGHT_ variables (those are ours)
		if strings.HasPrefix(varName, "DASHLIGHT_") {
			continue
		}

		// Check exact matches
		for _, pattern := range secretPatterns {
			if varName == pattern {
				s.foundVars = append(s.foundVars, varName)
				break
			}
		}

		// Check suffix matches (but avoid common false positives)
		for _, suffix := range secretSuffixes {
			if strings.HasSuffix(varName, suffix) {
				// Filter out some common false positives
				if varName == "PATH" || varName == "HOME" ||
					varName == "SHELL" || varName == "TERM" ||
					strings.HasPrefix(varName, "XDG_") {
					continue
				}
				s.foundVars = append(s.foundVars, varName)
				break
			}
		}
	}

	return len(s.foundVars) > 0
}
