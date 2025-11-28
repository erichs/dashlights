package signals

import (
	"context"
	"os/exec"
	"strings"
)

// GitEmailMismatchSignal checks for git email/remote mismatch
type GitEmailMismatchSignal struct {
	email  string
	remote string
}

func NewGitEmailMismatchSignal() *GitEmailMismatchSignal {
	return &GitEmailMismatchSignal{}
}

func (s *GitEmailMismatchSignal) Name() string {
	return "Mixed Identity"
}

func (s *GitEmailMismatchSignal) Emoji() string {
	return "🎭"
}

func (s *GitEmailMismatchSignal) Diagnostic() string {
	return "Git email '" + s.email + "' may not match remote '" + s.remote + "'"
}

func (s *GitEmailMismatchSignal) Remediation() string {
	return "Verify git user.email matches the repository context (work vs personal)"
}

func (s *GitEmailMismatchSignal) Check(ctx context.Context) bool {
	// Get git user email
	emailCmd := exec.CommandContext(ctx, "git", "config", "user.email")
	emailOut, err := emailCmd.Output()
	if err != nil {
		return false // Not in a git repo or git not configured
	}
	s.email = strings.TrimSpace(string(emailOut))

	// Get git remote URL
	remoteCmd := exec.CommandContext(ctx, "git", "config", "remote.origin.url")
	remoteOut, err := remoteCmd.Output()
	if err != nil {
		return false // No remote configured
	}
	s.remote = strings.TrimSpace(string(remoteOut))

	// Simple heuristic: check for common mismatches
	// Personal email with work remote or vice versa
	emailLower := strings.ToLower(s.email)
	remoteLower := strings.ToLower(s.remote)

	// Check if email is personal (gmail, yahoo, etc.) but remote is work
	isPersonalEmail := strings.Contains(emailLower, "gmail") ||
		strings.Contains(emailLower, "yahoo") ||
		strings.Contains(emailLower, "hotmail") ||
		strings.Contains(emailLower, "icloud") ||
		strings.Contains(emailLower, "aol") ||
		strings.Contains(emailLower, "live") ||
		strings.Contains(emailLower, "msn") ||
		strings.Contains(emailLower, "mail") ||
		strings.Contains(emailLower, "comcast") ||
		strings.Contains(emailLower, "verizon") ||
		strings.Contains(emailLower, "att") ||
		strings.Contains(emailLower, "sbcglobal") ||
		strings.Contains(emailLower, "proton") ||
		strings.Contains(emailLower, "zoho") ||
		strings.Contains(emailLower, "gmx") ||
		strings.Contains(emailLower, "outlook")

	isWorkRemote := strings.Contains(remoteLower, "github.com") &&
		!strings.Contains(remoteLower, "github.com/"+extractUsername(emailLower))

	// This is a simple check - could be enhanced
	if isPersonalEmail && isWorkRemote {
		return true
	}

	return false
}

func extractUsername(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
