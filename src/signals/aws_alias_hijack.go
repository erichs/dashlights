// Package signals defines security signal implementations used by dashlights.
package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/erichs/dashlights/src/signals/internal/homedirutil"
)

// AWSAliasHijackSignal detects potentially malicious AWS CLI aliases
// that override core AWS commands, which could indicate command injection attacks
type AWSAliasHijackSignal struct{}

// NewAWSAliasHijackSignal creates an AWSAliasHijackSignal.
func NewAWSAliasHijackSignal() Signal {
	return &AWSAliasHijackSignal{}
}

// Name returns the human-readable name of the signal.
func (s *AWSAliasHijackSignal) Name() string {
	return "AWS CLI Alias Hijacking"
}

// Emoji returns the emoji associated with the signal.
func (s *AWSAliasHijackSignal) Emoji() string {
	return "🪝" // Hook emoji - represents hijacking
}

// Diagnostic returns a description of the detected AWS CLI alias issues.
func (s *AWSAliasHijackSignal) Diagnostic() string {
	return "AWS CLI aliases override core commands or have insecure permissions"
}

// Remediation returns guidance on reviewing and hardening AWS CLI aliases.
func (s *AWSAliasHijackSignal) Remediation() string {
	return "Review ~/.aws/cli/alias for suspicious overrides and set permissions to 0600"
}

// Check inspects the AWS CLI alias file for dangerous overrides or permissions issues.
func (s *AWSAliasHijackSignal) Check(ctx context.Context) bool {
	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_AWS_ALIAS_HIJACK") != "" {
		return false
	}

	// Core AWS commands that should never be aliased (potential hijacking)
	coreCommands := map[string]bool{
		"cloudformation": true,
		"cloudtrail":     true,
		"cloudwatch":     true,
		"configure":      true,
		"dynamodb":       true,
		"ec2":            true,
		"ecr":            true,
		"eks":            true,
		"iam":            true,
		"kms":            true,
		"lambda":         true,
		"login":          true,
		"logs":           true,
		"rds":            true,
		"s3":             true,
		"secretsmanager": true,
		"ssm":            true,
		"sso":            true,
		"sts":            true,
	}

	aliasPath, err := homedirutil.SafeHomePath(".aws", "cli", "alias")
	if err != nil {
		return false
	}

	// Check if alias file exists
	// filepath.Clean for gosec G304 - path is already validated by SafeHomePath
	fileInfo, err := os.Stat(filepath.Clean(aliasPath))
	if err != nil {
		// File doesn't exist - no issue
		if os.IsNotExist(err) {
			return false
		}
		return false
	}

	// Check file permissions - should be 0600 (owner read/write only)
	mode := fileInfo.Mode()
	if mode.Perm() != 0600 {
		return true // Insecure permissions
	}

	// Parse alias file for suspicious aliases
	// filepath.Clean for gosec G304 - path is already validated by SafeHomePath
	file, err := os.Open(filepath.Clean(aliasPath))
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Parse alias definition: aliasname = command [--options]
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		aliasName := strings.TrimSpace(parts[0])

		// Check if this alias overrides a core command
		if coreCommands[aliasName] {
			return true
		}
	}

	return false
}
