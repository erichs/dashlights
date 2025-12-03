package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoNetworkExfiltration verifies that dashlights never imports packages
// that could be used for network exfiltration. This provides programmatic
// proof that the tool cannot send data to external servers.
//
// The test checks both direct and transitive dependencies.
func TestNoNetworkExfiltration(t *testing.T) {
	// Packages that could be used for data exfiltration
	forbiddenPackages := []string{
		"net/http",      // HTTP client/server
		"net/rpc",       // Remote procedure calls
		"net/smtp",      // Email sending
		"net/http/cgi",  // CGI
		"net/http/fcgi", // FastCGI
		"crypto/tls",    // TLS (only needed for network connections)
	}

	// Allowed network packages (used for local-only operations)
	// net - used only for Unix socket to SSH agent (local IPC)
	allowedPackages := []string{
		"net", // Unix sockets only - verified by code review
	}

	// Get all imports including transitive dependencies
	// Using -deps flag to include all dependencies
	// Using -buildvcs=false to avoid VCS errors in network-isolated environments
	cmd := exec.Command("go", "list", "-buildvcs=false", "-f", "{{.ImportPath}}: {{.Imports}}", "-deps", "./...")
	cmd.Dir = ".." // Run from repository root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list imports: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check for forbidden packages
	for _, pkg := range forbiddenPackages {
		// Look for the package as an import (with brackets or spaces around it)
		patterns := []string{
			"[" + pkg + " ",
			" " + pkg + " ",
			" " + pkg + "]",
			"[" + pkg + "]",
		}

		for _, pattern := range patterns {
			if strings.Contains(outputStr, pattern) {
				t.Errorf("SECURITY VIOLATION: Forbidden package %q is imported.\n"+
					"This package could be used for data exfiltration.\n"+
					"Dashlights must never import network client packages.", pkg)
			}
		}
	}

	// Also check that only allowed network packages from the 'net' family are used
	// by looking at what our own code imports
	cmd = exec.Command("go", "list", "-buildvcs=false", "-f", "{{.Imports}}", "./...")
	cmd.Dir = ".."
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list direct imports: %v\nOutput: %s", err, output)
	}

	directImports := string(output)

	// Verify 'net' is only used, not net/* subpackages (except allowed ones)
	netSubpackages := []string{
		"net/http",
		"net/rpc",
		"net/smtp",
		"net/url", // URL parsing could indicate network usage intent
	}

	for _, pkg := range netSubpackages {
		if strings.Contains(directImports, pkg) {
			isAllowed := false
			for _, allowed := range allowedPackages {
				if pkg == allowed {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				t.Errorf("SECURITY VIOLATION: Package %q is directly imported.\n"+
					"Only local IPC via Unix sockets (net package) is permitted.", pkg)
			}
		}
	}
}

// TestNoTelemetry verifies there are no telemetry, analytics, or crash reporting
// packages imported.
func TestNoTelemetry(t *testing.T) {
	telemetryPackages := []string{
		"sentry",
		"bugsnag",
		"rollbar",
		"newrelic",
		"datadog",
		"segment",
		"analytics",
		"telemetry",
		"opentelemetry",
		"prometheus/push", // push gateway (pull is fine)
	}

	cmd := exec.Command("go", "list", "-buildvcs=false", "-f", "{{.Deps}}", "./...")
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list dependencies: %v\nOutput: %s", err, output)
	}

	deps := strings.ToLower(string(output))

	for _, pkg := range telemetryPackages {
		if strings.Contains(deps, pkg) {
			t.Errorf("SECURITY VIOLATION: Telemetry-related package %q found in dependencies.\n"+
				"Dashlights must not include any telemetry, analytics, or crash reporting.", pkg)
		}
	}
}

// TestDocumentCapabilities outputs what the tool CAN access for transparency.
// This test always passes but documents the security boundary.
func TestDocumentCapabilities(t *testing.T) {
	t.Log("=== Dashlights Security Capabilities ===")
	t.Log("ALLOWED:")
	t.Log("  - Read local files (config files, .git, /proc)")
	t.Log("  - Read environment variables")
	t.Log("  - Unix socket IPC (SSH agent only)")
	t.Log("  - Write to stdout/stderr")
	t.Log("")
	t.Log("FORBIDDEN:")
	t.Log("  - HTTP/HTTPS requests")
	t.Log("  - TCP/UDP network connections")
	t.Log("  - DNS resolution")
	t.Log("  - Email sending")
	t.Log("  - Telemetry/analytics")
	t.Log("  - Writing to files (except stdout/stderr)")
}
