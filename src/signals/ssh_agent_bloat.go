package signals

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const (
	// SSH Agent Protocol constants
	msgRequestIdentities = 11
	msgIdentitiesAnswer  = 12

	// Maximum safe number of keys before MaxAuthTries issues
	maxSafeKeys = 5

	// Timeout for agent socket communication
	socketTimeout = 10 * time.Millisecond
)

// SSHAgentBloatSignal detects too many keys loaded in SSH agent
// This can cause MaxAuthTries lockouts, fingerprinting, and increased blast radius
type SSHAgentBloatSignal struct{}

// NewSSHAgentBloatSignal creates an SSHAgentBloatSignal.
func NewSSHAgentBloatSignal() Signal {
	return &SSHAgentBloatSignal{}
}

// Name returns the human-readable name of the signal.
func (s *SSHAgentBloatSignal) Name() string {
	return "SSH Agent Key Bloat"
}

// Emoji returns the emoji associated with the signal.
func (s *SSHAgentBloatSignal) Emoji() string {
	return "🔑" // Key emoji
}

// Diagnostic returns a description of the overloaded SSH agent state.
func (s *SSHAgentBloatSignal) Diagnostic() string {
	return "SSH agent has too many keys loaded (causes MaxAuthTries lockouts and fingerprinting)"
}

// Remediation returns guidance on how to reduce keys loaded into the SSH agent.
func (s *SSHAgentBloatSignal) Remediation() string {
	return "Run: ssh-add -D && ssh-add ~/.ssh/your_key OR add 'IdentitiesOnly yes' to ~/.ssh/config"
}

// Check queries the SSH agent and reports if too many keys are loaded.
func (s *SSHAgentBloatSignal) Check(ctx context.Context) bool {
	// Early exit if context already done
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// Check if this signal is disabled via environment variable
	if os.Getenv("DASHLIGHTS_DISABLE_SSH_AGENT_BLOAT") != "" {
		return false
	}

	// Check if SSH_AUTH_SOCK is set
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	if sockPath == "" {
		// No agent running - not a problem
		return false
	}

	// Get the number of keys loaded in the agent
	count, err := getAgentKeyCount(ctx, sockPath)
	if err != nil {
		// Can't communicate with agent - not a problem we can detect
		return false
	}

	// Trigger if more than 5 keys (default MaxAuthTries is 6)
	return count > maxSafeKeys
}

// getAgentKeyCount queries the SSH agent socket to count loaded keys
func getAgentKeyCount(ctx context.Context, socketPath string) (uint32, error) {
	// Early exit if context already done
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	// Use dialer with context for cancellable connection
	dialer := net.Dialer{Timeout: socketTimeout}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Set a tight deadline to avoid blocking on read/write
	if err := conn.SetDeadline(time.Now().Add(socketTimeout)); err != nil {
		return 0, err
	}

	// Send SSH_AGENTC_REQUEST_IDENTITIES message
	// Format: [length:4 bytes][type:1 byte]
	// Length = 1 (just the type byte)
	payload := []byte{0, 0, 0, 1, msgRequestIdentities}
	if _, err := conn.Write(payload); err != nil {
		return 0, err
	}

	// Read response header (5 bytes: length + type)
	header := make([]byte, 5)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, err
	}

	// Verify we got SSH_AGENT_IDENTITIES_ANSWER
	// Bounds check satisfies gosec G602 (io.ReadFull guarantees 5 bytes on success)
	if len(header) >= 5 && header[4] != msgIdentitiesAnswer {
		return 0, fmt.Errorf("unexpected message type: %d", header[4])
	}

	// Read the key count (next 4 bytes)
	countBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, countBuf); err != nil {
		return 0, err
	}

	// Parse as big-endian uint32
	return binary.BigEndian.Uint32(countBuf), nil
}
