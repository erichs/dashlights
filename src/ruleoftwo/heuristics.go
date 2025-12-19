package ruleoftwo

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Capability represents one of the three Rule of Two capabilities.
type Capability int

const (
	// CapabilityA represents processing untrustworthy inputs.
	CapabilityA Capability = iota
	// CapabilityB represents access to sensitive systems or data.
	CapabilityB
	// CapabilityC represents state changes or external communication.
	CapabilityC
)

// String returns a human-readable name for the capability.
func (c Capability) String() string {
	switch c {
	case CapabilityA:
		return "A (untrustworthy input)"
	case CapabilityB:
		return "B (sensitive access)"
	case CapabilityC:
		return "C (state change/external comms)"
	default:
		return "unknown"
	}
}

// CapabilityResult holds the detection result for a single capability.
type CapabilityResult struct {
	Detected bool
	Reasons  []string
}

// untrustedPathPatterns are paths that typically contain untrusted data.
var untrustedPathPatterns = []string{
	"/tmp/",
	"/var/tmp/",
	"/dev/shm/",
	"/downloads/",
	"/Downloads/",
	"~/Downloads/",
}

// untrustedContentMarkers indicate content from external/untrusted sources.
var untrustedContentMarkers = []string{
	"${", // variable expansion
	"$(", // command substitution
	"`",  // backtick command substitution
	"eval(",
}

// externalDataCommands are bash commands that fetch external data.
var externalDataCommands = []string{
	"curl",
	"wget",
	"fetch",
	"http",
	"nc ",
	"netcat",
	// Version control fetching
	"git clone",
	"git pull",
	"git fetch",
	"svn checkout",
	"svn update",
	"hg clone",
	"hg pull",
	// Alternative downloaders
	"aria2c",
	"lynx -source",
	"w3m -dump",
}

// obfuscationPatterns indicate encoded/obfuscated command execution.
// These are treated as Capability A (untrustworthy input) because they
// could be hiding malicious commands.
var obfuscationPatterns = []string{
	"base64 -d",
	"base64 --decode",
	"xxd -r",
	"| bash",
	"| sh",
	"| zsh",
	"| /bin/bash",
	"| /bin/sh",
	"eval ",
	"source <(",
	". <(",
}

// reverseShellPatterns indicate attempts to establish reverse shells.
// These combine Capability A (external) + C (state change/comms).
var reverseShellPatterns = []string{
	"/dev/tcp/",
	"/dev/udp/",
	"nc -e",
	"nc -c",
	"ncat -e",
	"ncat -c",
	"socat exec:",
	"bash -i >",
	"sh -i >",
	"mkfifo",
	"0<&1",
	">&0 2>&0",
}

// sensitivePathPatterns indicate access to sensitive data.
var sensitivePathPatterns = []string{
	".env",
	".aws/",
	".ssh/",
	".kube/",
	".gnupg/",
	".npmrc",
	".pypirc",
	".netrc",
	".docker/config.json",
	"credentials",
	"secrets",
	"id_rsa",
	"id_ed25519",
	"id_ecdsa",
	"id_dsa",
	"known_hosts",
	"authorized_keys",
	// Additional cloud provider configs
	".config/gcloud/",
	".azure/",
	".config/doctl/",
	".oci/",
	".config/gh/",
	".config/hub",
	// Language package manager credentials
	".gem/credentials",
	".cargo/credentials",
	".gradle/gradle.properties",
	".m2/settings.xml",
	".composer/auth.json",
	".terraform.d/credentials",
	".terraformrc",
	// Database credentials
	".pgpass",
	".my.cnf",
	".mysql_history",
	// Git config (may contain creds)
	".git/config",
	".gitconfig",
	// Web auth
	".htpasswd",
}

// sensitiveFileExtensions indicate sensitive key/certificate files.
var sensitiveFileExtensions = []string{
	".pem",
	".key",
	".p12",
	".pfx",
	".crt",
	".cer",
}

// sensitiveCommands are bash commands that access sensitive systems.
var sensitiveCommands = []string{
	"aws ",
	"kubectl ",
	"gcloud ",
	"az ",
	"terraform ",
	"vault ",
	"op ",   // 1Password CLI
	"pass ", // password-store
	"gpg ",
	"ssh-add",
	"ssh-keygen",
	// Additional cloud CLIs
	"doctl ",
	"linode-cli ",
	"heroku ",
	"oci ",
	"ibmcloud ",
	"flyctl ",
	// Container runtimes
	"podman ",
	"buildah ",
	// Orchestration
	"helm ",
	"oc ", // OpenShift
	"nomad ",
	"consul ",
	// Config management
	"ansible ",
	"ansible-playbook ",
	// Database access
	"psql ",
	"mysql ",
	"mongo ",
	"mongosh ",
	"redis-cli ",
}

// productionIndicators suggest access to production systems.
var productionIndicators = []string{
	"/prod/",
	"/production/",
	"prd-",
	"prod-",
	"-prod",
	"-prd",
	".prod.",
	".production.",
}

// stateChangingCommands modify filesystem or system state.
var stateChangingCommands = []string{
	"rm ",
	"rm\t",
	"rmdir ",
	"mv ",
	"cp ",
	"chmod ",
	"chown ",
	"touch ",
	"mkdir ",
	"ln ",
	"install ",
	"git commit",
	"git push",
	"git checkout",
	"git reset",
	"git rebase",
	"git merge",
	"npm install",
	"npm publish",
	"npm update",
	"yarn add",
	"yarn install",
	"pip install",
	"pip uninstall",
	"docker run",
	"docker exec",
	"docker build",
	"docker push",
	"kubectl apply",
	"kubectl delete",
	"kubectl exec",
	"kubectl create",
	"kubectl patch",
	"terraform apply",
	"terraform destroy",
	"terraform import",
	"make ",
	"make\t",
	// Alternative deletion/modification
	"shred ",
	"truncate ",
	"dd if=",
	// In-place editors
	"sed -i",
	"perl -i",
	// Process control
	"kill ",
	"killall ",
	"pkill ",
	"systemctl ",
	// Additional package managers
	"go install",
	"go get ",
	"cargo install",
	"gem install",
	"composer install",
	"composer update",
	"brew install",
	"brew uninstall",
	"apt install",
	"apt-get install",
	"apt remove",
	"yum install",
	"dnf install",
	"pacman -S",
	"snap install",
	// Container variants
	"podman run",
	"podman exec",
	"podman build",
	"docker-compose up",
	"docker-compose down",
	// IaC tools
	"ansible-playbook ",
	"pulumi up",
	"pulumi destroy",
	// Sync/transfer tools
	"rclone ",
	"s3cmd ",
	"gsutil ",
	"az storage ",
}

// externalCommPatterns indicate external network communication.
var externalCommPatterns = []string{
	"curl",
	"wget",
	"ssh ",
	"scp ",
	"rsync ",
	"sftp ",
	"ftp ",
	"nc ",
	"netcat ",
	"ncat ",
	"telnet ",
	"nmap ",
	"socat ",
	// Reverse shell indicators (network + state change)
	"/dev/tcp/",
	"/dev/udp/",
}

// redirectPatterns indicate output redirection (state change).
var redirectPatterns = []string{
	" > ",
	" >> ",
	" >| ",
	" 2> ",
	" 2>> ",
	" &> ",
	" &>> ",
}

// pipePatterns that may indicate processing external data.
var pipeFromExternalPattern = regexp.MustCompile(`(curl|wget|nc|netcat)\s+[^|]*\|`)

// DetectCapabilityA checks for untrustworthy input processing.
func DetectCapabilityA(toolName string, input map[string]interface{}, cwd string) CapabilityResult {
	result := CapabilityResult{Detected: false, Reasons: []string{}}

	switch toolName {
	case "WebFetch":
		// WebFetch always involves external data
		result.Detected = true
		webInput := ParseWebFetchInput(input)
		result.Reasons = append(result.Reasons, "fetching external URL: "+truncate(webInput.URL, 50))

	case "WebSearch":
		// WebSearch involves external data
		result.Detected = true
		result.Reasons = append(result.Reasons, "web search returns external data")

	case "Bash":
		bashInput := ParseBashInput(input)
		cmd := bashInput.Command
		cmdLower := strings.ToLower(cmd)

		// Check for commands that fetch external data
		for _, extCmd := range externalDataCommands {
			if strings.Contains(cmdLower, strings.ToLower(extCmd)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "command fetches external data: "+extCmd)
				break
			}
		}

		// Check for piping from external commands
		if pipeFromExternalPattern.MatchString(cmd) {
			result.Detected = true
			result.Reasons = append(result.Reasons, "piping data from external source")
		}

		// Check for obfuscation patterns (treated as untrusted input)
		for _, pattern := range obfuscationPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "obfuscated/encoded command: "+pattern)
				break
			}
		}

		// Check for reverse shell patterns (external connection attempt)
		for _, pattern := range reverseShellPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "reverse shell pattern: "+pattern)
				break
			}
		}

	case "Read":
		readInput := ParseReadInput(input)
		path := readInput.FilePath

		// Check if reading from untrusted locations
		for _, pattern := range untrustedPathPatterns {
			if strings.Contains(strings.ToLower(path), strings.ToLower(pattern)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "reading from untrusted path: "+pattern)
				break
			}
		}

		// Reading files outside cwd could be untrusted
		if cwd != "" && !strings.HasPrefix(path, cwd) && filepath.IsAbs(path) {
			// Allow home directory reads as they're typically trusted
			if !strings.HasPrefix(path, "/Users/") && !strings.HasPrefix(path, "/home/") {
				result.Detected = true
				result.Reasons = append(result.Reasons, "reading file outside project directory")
			}
		}

	case "Write", "Edit":
		// Check if content contains untrusted markers
		var content string
		if toolName == "Write" {
			writeInput := ParseWriteInput(input)
			content = writeInput.Content
		} else {
			editInput := ParseEditInput(input)
			content = editInput.NewString
		}

		for _, marker := range untrustedContentMarkers {
			if strings.Contains(content, marker) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "content contains dynamic expansion: "+marker)
				break
			}
		}
	}

	return result
}

// DetectCapabilityB checks for access to sensitive systems or data.
func DetectCapabilityB(toolName string, input map[string]interface{}) CapabilityResult {
	result := CapabilityResult{Detected: false, Reasons: []string{}}

	// Get file path based on tool type
	var filePath string
	switch toolName {
	case "Read":
		filePath = ParseReadInput(input).FilePath
	case "Write":
		filePath = ParseWriteInput(input).FilePath
	case "Edit":
		filePath = ParseEditInput(input).FilePath
	case "Glob":
		filePath = ParseGlobInput(input).Path
	case "Grep":
		filePath = ParseGrepInput(input).Path
	}

	// Check path-based tools for sensitive access
	if filePath != "" {
		pathLower := strings.ToLower(filePath)

		// Check sensitive path patterns
		for _, pattern := range sensitivePathPatterns {
			if strings.Contains(pathLower, strings.ToLower(pattern)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "accessing sensitive path: "+pattern)
				break
			}
		}

		// Check sensitive file extensions
		for _, ext := range sensitiveFileExtensions {
			if strings.HasSuffix(pathLower, ext) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "accessing sensitive file type: "+ext)
				break
			}
		}

		// Check production indicators
		for _, indicator := range productionIndicators {
			if strings.Contains(pathLower, strings.ToLower(indicator)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "accessing production path: "+indicator)
				break
			}
		}
	}

	// Check Bash commands for sensitive operations
	if toolName == "Bash" {
		bashInput := ParseBashInput(input)
		cmd := bashInput.Command
		cmdLower := strings.ToLower(cmd)

		// Check for sensitive commands
		for _, sensitiveCmd := range sensitiveCommands {
			if strings.Contains(cmdLower, strings.ToLower(sensitiveCmd)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "running sensitive command: "+strings.TrimSpace(sensitiveCmd))
				break
			}
		}

		// Check for sensitive paths in command (e.g., tee ~/.aws/credentials)
		for _, pattern := range sensitivePathPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "command accesses sensitive path: "+pattern)
				break
			}
		}

		// Check for sensitive file extensions in command
		for _, ext := range sensitiveFileExtensions {
			if strings.Contains(cmdLower, ext) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "command accesses sensitive file type: "+ext)
				break
			}
		}

		// Check for production indicators in command
		for _, indicator := range productionIndicators {
			if strings.Contains(cmdLower, strings.ToLower(indicator)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "command references production: "+indicator)
				break
			}
		}
	}

	return result
}

// DetectCapabilityC checks for state changes or external communication.
func DetectCapabilityC(toolName string, input map[string]interface{}) CapabilityResult {
	result := CapabilityResult{Detected: false, Reasons: []string{}}

	switch toolName {
	case "Write":
		// Write always changes state
		result.Detected = true
		writeInput := ParseWriteInput(input)
		result.Reasons = append(result.Reasons, "writing file: "+truncate(writeInput.FilePath, 50))

	case "Edit":
		// Edit always changes state
		result.Detected = true
		editInput := ParseEditInput(input)
		result.Reasons = append(result.Reasons, "editing file: "+truncate(editInput.FilePath, 50))

	case "NotebookEdit":
		// NotebookEdit always changes state
		result.Detected = true
		result.Reasons = append(result.Reasons, "modifying notebook")

	case "TodoWrite":
		// TodoWrite changes state
		result.Detected = true
		result.Reasons = append(result.Reasons, "modifying todo list state")

	case "Bash":
		bashInput := ParseBashInput(input)
		cmd := bashInput.Command
		cmdLower := strings.ToLower(cmd)

		// Check for state-changing commands
		for _, stateCmd := range stateChangingCommands {
			if strings.Contains(cmdLower, strings.ToLower(stateCmd)) {
				result.Detected = true
				result.Reasons = append(result.Reasons, "state-changing command: "+strings.TrimSpace(stateCmd))
				break
			}
		}

		// Check for external communication
		if !result.Detected {
			for _, extComm := range externalCommPatterns {
				if strings.Contains(cmdLower, strings.ToLower(extComm)) {
					result.Detected = true
					result.Reasons = append(result.Reasons, "external communication: "+strings.TrimSpace(extComm))
					break
				}
			}
		}

		// Check for output redirection
		if !result.Detected {
			for _, redirect := range redirectPatterns {
				if strings.Contains(cmd, redirect) {
					result.Detected = true
					result.Reasons = append(result.Reasons, "output redirection to file")
					break
				}
			}
		}
	}

	return result
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
