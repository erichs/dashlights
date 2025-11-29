package signals

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// AWSAliasHijackSignal detects potentially malicious AWS CLI aliases
// that override core AWS commands, which could indicate command injection attacks
type AWSAliasHijackSignal struct{}

func NewAWSAliasHijackSignal() Signal {
	return &AWSAliasHijackSignal{}
}

func (s *AWSAliasHijackSignal) Name() string {
	return "AWS CLI Alias Hijacking"
}

func (s *AWSAliasHijackSignal) Emoji() string {
	return "🪝" // Hook emoji - represents hijacking
}

func (s *AWSAliasHijackSignal) Diagnostic() string {
	return "AWS CLI aliases override core commands or have insecure permissions"
}

func (s *AWSAliasHijackSignal) Remediation() string {
	return "Review ~/.aws/cli/alias for suspicious overrides and set permissions to 0600"
}

func (s *AWSAliasHijackSignal) Check(ctx context.Context) bool {
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

	aliasPath := filepath.Join(sanitizedHome, ".aws", "cli", "alias")

	// Check if alias file exists
	fileInfo, err := os.Stat(aliasPath)
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
	file, err := os.Open(aliasPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
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
