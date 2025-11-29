package signals

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSSHAgentBloatSignal_NoAgent(t *testing.T) {
	// Unset SSH_AUTH_SOCK
	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when SSH_AUTH_SOCK is not set")
	}
}

func TestSSHAgentBloatSignal_InvalidSocket(t *testing.T) {
	// Set SSH_AUTH_SOCK to non-existent path
	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", "/tmp/nonexistent_socket_12345")
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when socket doesn't exist")
	}
}

func TestSSHAgentBloatSignal_FewKeys(t *testing.T) {
	// Create a mock SSH agent with 3 keys
	sockPath := createMockAgent(t, 3)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent has 3 keys (under threshold)")
	}
}

func TestSSHAgentBloatSignal_ExactlyFiveKeys(t *testing.T) {
	// Create a mock SSH agent with exactly 5 keys (threshold)
	sockPath := createMockAgent(t, 5)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent has exactly 5 keys (at threshold)")
	}
}

func TestSSHAgentBloatSignal_SixKeys(t *testing.T) {
	// Create a mock SSH agent with 6 keys (over threshold)
	sockPath := createMockAgent(t, 6)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when agent has 6 keys (over threshold)")
	}
}

func TestSSHAgentBloatSignal_ManyKeys(t *testing.T) {
	// Create a mock SSH agent with 20 keys (way over threshold)
	sockPath := createMockAgent(t, 20)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when agent has 20 keys (way over threshold)")
	}
}

func TestSSHAgentBloatSignal_ZeroKeys(t *testing.T) {
	// Create a mock SSH agent with 0 keys
	sockPath := createMockAgent(t, 0)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent has 0 keys")
	}
}

// createMockAgent creates a Unix socket that mimics SSH agent protocol
func createMockAgent(t *testing.T, keyCount uint32) string {
	// Use /tmp directly to avoid long paths (macOS has 104 char limit for Unix sockets)
	sockPath := filepath.Join("/tmp", fmt.Sprintf("test_agent_%d.sock", time.Now().UnixNano()))

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create mock agent socket: %v", err)
	}

	// Clean up socket on test completion
	t.Cleanup(func() {
		listener.Close()
		os.Remove(sockPath)
	})

	// Start a goroutine to handle connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleMockAgentConnection(conn, keyCount)
		}
	}()

	// Give the listener time to start
	time.Sleep(5 * time.Millisecond)

	return sockPath
}

// handleMockAgentConnection handles a single connection to the mock agent
func handleMockAgentConnection(conn net.Conn, keyCount uint32) {
	defer conn.Close()

	// Read the request
	header := make([]byte, 5)
	_, err := conn.Read(header)
	if err != nil {
		return
	}

	// Verify it's a REQUEST_IDENTITIES message
	if header[4] != msgRequestIdentities {
		return
	}

	// Send IDENTITIES_ANSWER response
	// Format: [length:4][type:1][count:4]
	// Length = 1 (type) + 4 (count) = 5
	response := make([]byte, 9)
	binary.BigEndian.PutUint32(response[0:4], 5)        // Length
	response[4] = msgIdentitiesAnswer                   // Type
	binary.BigEndian.PutUint32(response[5:9], keyCount) // Count

	conn.Write(response)
}

func TestSSHAgentBloatSignal_IncompleteHeader(t *testing.T) {
	// Create a mock agent that sends incomplete header (less than 5 bytes)
	sockPath := createMockAgentWithIncompleteHeader(t)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	// Should return false (error handled gracefully)
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent sends incomplete header")
	}
}

func TestSSHAgentBloatSignal_IncompleteCount(t *testing.T) {
	// Create a mock agent that sends incomplete count (less than 4 bytes)
	sockPath := createMockAgentWithIncompleteCount(t)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	// Should return false (error handled gracefully)
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent sends incomplete count")
	}
}

func TestSSHAgentBloatSignal_WrongMessageType(t *testing.T) {
	// Create a mock agent that sends wrong message type
	sockPath := createMockAgentWithWrongMessageType(t)

	originalSock := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", sockPath)
	defer func() {
		if originalSock != "" {
			os.Setenv("SSH_AUTH_SOCK", originalSock)
		} else {
			os.Unsetenv("SSH_AUTH_SOCK")
		}
	}()

	signal := NewSSHAgentBloatSignal()
	ctx := context.Background()

	// Should return false (error handled gracefully)
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when agent sends wrong message type")
	}
}

// createMockAgentWithIncompleteHeader creates a mock agent that sends only 3 bytes
func createMockAgentWithIncompleteHeader(t *testing.T) string {
	sockPath := filepath.Join("/tmp", fmt.Sprintf("test_agent_incomplete_header_%d.sock", time.Now().UnixNano()))

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create mock agent socket: %v", err)
	}

	t.Cleanup(func() {
		listener.Close()
		os.Remove(sockPath)
	})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// Read request
				header := make([]byte, 5)
				c.Read(header)
				// Send incomplete response (only 3 bytes instead of 5)
				c.Write([]byte{0, 0, 0})
			}(conn)
		}
	}()

	time.Sleep(5 * time.Millisecond)
	return sockPath
}

// createMockAgentWithIncompleteCount creates a mock agent that sends incomplete count
func createMockAgentWithIncompleteCount(t *testing.T) string {
	sockPath := filepath.Join("/tmp", fmt.Sprintf("test_agent_incomplete_count_%d.sock", time.Now().UnixNano()))

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create mock agent socket: %v", err)
	}

	t.Cleanup(func() {
		listener.Close()
		os.Remove(sockPath)
	})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// Read request
				header := make([]byte, 5)
				c.Read(header)
				// Send valid header but incomplete count (only 2 bytes instead of 4)
				response := make([]byte, 7)
				binary.BigEndian.PutUint32(response[0:4], 5)
				response[4] = msgIdentitiesAnswer
				// Only send 7 bytes total (5 for header + 2 for partial count)
				c.Write(response[:7])
			}(conn)
		}
	}()

	time.Sleep(5 * time.Millisecond)
	return sockPath
}

// createMockAgentWithWrongMessageType creates a mock agent that sends wrong message type
func createMockAgentWithWrongMessageType(t *testing.T) string {
	sockPath := filepath.Join("/tmp", fmt.Sprintf("test_agent_wrong_type_%d.sock", time.Now().UnixNano()))

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("Failed to create mock agent socket: %v", err)
	}

	t.Cleanup(func() {
		listener.Close()
		os.Remove(sockPath)
	})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				// Read request
				header := make([]byte, 5)
				c.Read(header)
				// Send response with wrong message type (99 instead of msgIdentitiesAnswer)
				response := make([]byte, 9)
				binary.BigEndian.PutUint32(response[0:4], 5)
				response[4] = 99 // Wrong message type
				binary.BigEndian.PutUint32(response[5:9], 10)
				c.Write(response)
			}(conn)
		}
	}()

	time.Sleep(5 * time.Millisecond)
	return sockPath
}
