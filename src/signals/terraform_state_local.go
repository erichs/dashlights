package signals

import (
	"context"
	"os"
)

// TerraformStateLocalSignal checks for local terraform.tfstate files
// In team environments, state should be remote (S3/GCS), not local
type TerraformStateLocalSignal struct{}

// NewTerraformStateLocalSignal creates a TerraformStateLocalSignal.
func NewTerraformStateLocalSignal() Signal {
	return &TerraformStateLocalSignal{}
}

// Name returns the human-readable name of the signal.
func (s *TerraformStateLocalSignal) Name() string {
	return "Local Terraform State"
}

// Emoji returns the emoji associated with the signal.
func (s *TerraformStateLocalSignal) Emoji() string {
	return "🏗️" // Building construction (infrastructure)
}

// Diagnostic returns a description of the detected local Terraform state.
func (s *TerraformStateLocalSignal) Diagnostic() string {
	return "terraform.tfstate file exists locally (should use remote state in team environments)"
}

// Remediation returns guidance on moving Terraform state to a remote backend.
func (s *TerraformStateLocalSignal) Remediation() string {
	return "Configure remote backend (S3/GCS) and migrate state with 'terraform init -migrate-state'"
}

// Check looks for local terraform.tfstate files indicating non-remote state usage.
func (s *TerraformStateLocalSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_TERRAFORM_STATE_LOCAL") != "" {
		return false
	}

	// Check if terraform.tfstate exists in current directory
	if _, err := os.Stat("terraform.tfstate"); err == nil {
		return true
	}

	// Also check for terraform.tfstate.backup (indicates recent local state usage)
	if _, err := os.Stat("terraform.tfstate.backup"); err == nil {
		return true
	}

	return false
}
