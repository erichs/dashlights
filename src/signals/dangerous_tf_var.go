package signals

import (
	"context"
	"os"
	"strings"
)

// DangerousTFVarSignal checks for dangerous Terraform variables in environment
// Passing secrets via TF_VAR_ often leaves them in shell history
type DangerousTFVarSignal struct {
	foundVars []string
}

// NewDangerousTFVarSignal creates a DangerousTFVarSignal.
func NewDangerousTFVarSignal() Signal {
	return &DangerousTFVarSignal{}
}

// Name returns the human-readable name of the signal.
func (s *DangerousTFVarSignal) Name() string {
	return "Dangerous TF_VAR"
}

// Emoji returns the emoji associated with the signal.
func (s *DangerousTFVarSignal) Emoji() string {
	return "🔐" // Locked with key (secrets)
}

// Diagnostic returns a description of the detected dangerous Terraform variables.
func (s *DangerousTFVarSignal) Diagnostic() string {
	if len(s.foundVars) > 0 {
		return "Dangerous Terraform variables in environment: " + s.foundVars[0] + " (secrets in shell history)"
	}
	return "Dangerous Terraform variables in environment (secrets in shell history)"
}

// Remediation returns guidance on safer handling of Terraform secrets.
func (s *DangerousTFVarSignal) Remediation() string {
	return "Use .tfvars files or secret management instead of TF_VAR_ environment variables for secrets"
}

// Check scans TF_VAR_ environment variables for names that likely hold secrets.
func (s *DangerousTFVarSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_DANGEROUS_TF_VAR") != "" {
		return false
	}

	s.foundVars = []string{}

	// Dangerous patterns to look for in TF_VAR_ variables
	dangerousPatterns := []string{
		"access_key",
		"secret_key",
		"password",
		"token",
		"api_key",
		"private_key",
		"secret",
		"credential",
	}

	// Scan all environment variables
	for _, env := range os.Environ() {
		// Check if it's a TF_VAR_ variable
		if !strings.HasPrefix(env, "TF_VAR_") {
			continue
		}

		// Extract variable name (before the = sign)
		parts := strings.SplitN(env, "=", 2)
		if len(parts) < 1 {
			continue
		}

		varName := parts[0]
		lowerVarName := strings.ToLower(varName)

		// Check if variable name contains dangerous patterns
		for _, pattern := range dangerousPatterns {
			if strings.Contains(lowerVarName, pattern) {
				s.foundVars = append(s.foundVars, varName)
				break
			}
		}
	}

	return len(s.foundVars) > 0
}
