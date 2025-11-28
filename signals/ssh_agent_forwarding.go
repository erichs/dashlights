package signals

import (
	"context"
	"os"
	"strings"
)

// SSHAgentForwardingSignal checks for SSH agent forwarding
type SSHAgentForwardingSignal struct{}

func NewSSHAgentForwardingSignal() *SSHAgentForwardingSignal {
	return &SSHAgentForwardingSignal{}
}

func (s *SSHAgentForwardingSignal) Name() string {
	return "Ghost Agent"
}

func (s *SSHAgentForwardingSignal) Emoji() string {
	return "👻"
}

func (s *SSHAgentForwardingSignal) Diagnostic() string {
	return "SSH agent forwarding detected (remote socket)"
}

func (s *SSHAgentForwardingSignal) Remediation() string {
	return "Avoid using 'ssh -A' on untrusted servers; use ProxyJump instead"
}

func (s *SSHAgentForwardingSignal) Check(ctx context.Context) bool {
	authSock := os.Getenv("SSH_AUTH_SOCK")
	if authSock == "" {
		return false
	}
	
	// Forwarded SSH agent sockets typically contain patterns like:
	// /tmp/ssh-XXXXXX/agent.NNNN (remote forwarded)
	// vs local sockets which might be in different locations
	
	// Check for SSH_CONNECTION which indicates we're in an SSH session
	sshConnection := os.Getenv("SSH_CONNECTION")
	if sshConnection == "" {
		return false // Not in an SSH session
	}
	
	// If we're in an SSH session AND have SSH_AUTH_SOCK, 
	// it's likely forwarded (especially if in /tmp/ssh-*)
	if strings.Contains(authSock, "/tmp/ssh-") {
		return true
	}
	
	return false
}

