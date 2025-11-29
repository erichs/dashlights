package signals

import (
	"context"
	"os"
)

// TerraformStateLocalSignal checks for local terraform.tfstate files
// In team environments, state should be remote (S3/GCS), not local
type TerraformStateLocalSignal struct{}

func NewTerraformStateLocalSignal() Signal {
	return &TerraformStateLocalSignal{}
}

func (s *TerraformStateLocalSignal) Name() string {
	return "Local Terraform State"
}

func (s *TerraformStateLocalSignal) Emoji() string {
	return "🏗️" // Building construction (infrastructure)
}

func (s *TerraformStateLocalSignal) Diagnostic() string {
	return "terraform.tfstate file exists locally (should use remote state in team environments)"
}

func (s *TerraformStateLocalSignal) Remediation() string {
	return "Configure remote backend (S3/GCS) and migrate state with 'terraform init -migrate-state'"
}

func (s *TerraformStateLocalSignal) Check(ctx context.Context) bool {
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

