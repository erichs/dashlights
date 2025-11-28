package signals

import (
	"context"
	"encoding/binary"
	"fmt"
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

func NewSSHAgentBloatSignal() Signal {
	return &SSHAgentBloatSignal{}
}

func (s *SSHAgentBloatSignal) Name() string {
	return "SSH Agent Key Bloat"
}

func (s *SSHAgentBloatSignal) Emoji() string {
	return "🔑" // Key emoji
}

func (s *SSHAgentBloatSignal) Diagnostic() string {
	return "SSH agent has too many keys loaded (causes MaxAuthTries lockouts and fingerprinting)"
}

func (s *SSHAgentBloatSignal) Remediation() string {
	return "Run: ssh-add -D && ssh-add ~/.ssh/your_key OR add 'IdentitiesOnly yes' to ~/.ssh/config"
}

func (s *SSHAgentBloatSignal) Check(ctx context.Context) bool {
	// Check if SSH_AUTH_SOCK is set
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	if sockPath == "" {
		// No agent running - not a problem
		return false
	}

	// Get the number of keys loaded in the agent
	count, err := getAgentKeyCount(sockPath)
	if err != nil {
		// Can't communicate with agent - not a problem we can detect
		return false
	}

	// Trigger if more than 5 keys (default MaxAuthTries is 6)
	return count > maxSafeKeys
}

// getAgentKeyCount queries the SSH agent socket to count loaded keys
func getAgentKeyCount(socketPath string) (uint32, error) {
	// Connect to the Unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Set a tight deadline to avoid blocking
	conn.SetDeadline(time.Now().Add(socketTimeout))

	// Send SSH_AGENTC_REQUEST_IDENTITIES message
	// Format: [length:4 bytes][type:1 byte]
	// Length = 1 (just the type byte)
	payload := []byte{0, 0, 0, 1, msgRequestIdentities}
	if _, err := conn.Write(payload); err != nil {
		return 0, err
	}

	// Read response header (5 bytes: length + type)
	header := make([]byte, 5)
	if _, err := conn.Read(header); err != nil {
		return 0, err
	}

	// Verify we got SSH_AGENT_IDENTITIES_ANSWER
	if header[4] != msgIdentitiesAnswer {
		return 0, fmt.Errorf("unexpected message type: %d", header[4])
	}

	// Read the key count (next 4 bytes)
	countBuf := make([]byte, 4)
	if _, err := conn.Read(countBuf); err != nil {
		return 0, err
	}

	// Parse as big-endian uint32
	count := binary.BigEndian.Uint32(countBuf)
	
	return count, nil
}

