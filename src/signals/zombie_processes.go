package signals

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// procFS is an interface for interacting with the /proc filesystem
// This allows for mocking in tests
type procFS interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

// realProcFS implements procFS using actual OS calls
type realProcFS struct{}

func (r *realProcFS) ReadFile(name string) ([]byte, error) {
	// Validate path to prevent directory traversal
	if strings.Contains(name, "..") {
		return nil, os.ErrInvalid
	}

	// Clean the path
	cleanPath := filepath.Clean(name)

	// Ensure the cleaned path doesn't contain directory traversal
	if strings.Contains(cleanPath, "..") {
		return nil, os.ErrInvalid
	}

	return os.ReadFile(cleanPath)
}

func (r *realProcFS) ReadDir(name string) ([]fs.DirEntry, error) {
	// Validate path to prevent directory traversal
	if strings.Contains(name, "..") {
		return nil, os.ErrInvalid
	}

	// Clean the path
	cleanPath := filepath.Clean(name)

	// Ensure the cleaned path doesn't contain directory traversal
	if strings.Contains(cleanPath, "..") {
		return nil, os.ErrInvalid
	}

	return os.ReadDir(cleanPath)
}

// ZombieProcessesSignal checks for excessive zombie processes
type ZombieProcessesSignal struct {
	count int
	fs    procFS
}

func NewZombieProcessesSignal() *ZombieProcessesSignal {
	return &ZombieProcessesSignal{
		fs: &realProcFS{},
	}
}

func (s *ZombieProcessesSignal) Name() string {
	return "Zombie Apocalypse"
}

func (s *ZombieProcessesSignal) Emoji() string {
	return "🧟"
}

func (s *ZombieProcessesSignal) Diagnostic() string {
	return fmt.Sprintf("Excessive zombie processes detected: %d", s.count)
}

func (s *ZombieProcessesSignal) Remediation() string {
	return "Investigate and fix parent processes not reaping children"
}

func (s *ZombieProcessesSignal) Check(ctx context.Context) bool {
	// Only applicable on Linux
	if runtime.GOOS != "linux" {
		return false
	}

	return s.checkWithFS(ctx, s.fs)
}

// checkWithFS performs the actual check using the provided filesystem interface
// This is separated to allow for testing with mocked filesystems
func (s *ZombieProcessesSignal) checkWithFS(ctx context.Context, fs procFS) bool {
	// Read /proc/stat to count zombie processes
	data, err := fs.ReadFile("/proc/stat")
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "processes") {
			// This gives us total processes, but we need zombies
			// Let's count from /proc instead
			break
		}
	}

	// Count zombie processes by checking /proc/*/stat
	zombieCount := 0
	entries, err := fs.ReadDir("/proc")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if !entry.IsDir() {
			continue
		}

		// Validate directory name is a valid PID (all numeric)
		name := entry.Name()
		if _, err := strconv.Atoi(name); err != nil {
			// Not a valid PID, skip
			continue
		}

		// Additional validation: ensure name doesn't contain directory traversal
		if strings.Contains(name, "..") || strings.Contains(name, "/") {
			continue
		}

		// Clean the name
		name = filepath.Clean(name)

		// Validate again after cleaning
		if strings.Contains(name, "..") || strings.Contains(name, "/") {
			continue
		}

		// Safely construct path to /proc/PID/stat
		statPath := filepath.Join("/proc", name, "stat")

		// Clean the final path
		statPath = filepath.Clean(statPath)

		statData, err := fs.ReadFile(statPath)
		if err != nil {
			continue
		}

		// Parse state (third field after closing paren)
		statStr := string(statData)
		// Find the last ')' to handle process names with spaces
		lastParen := strings.LastIndex(statStr, ")")
		if lastParen == -1 {
			continue
		}

		fields := strings.Fields(statStr[lastParen+1:])
		if len(fields) > 0 && fields[0] == "Z" {
			zombieCount++
		}
	}

	s.count = zombieCount

	// Threshold: more than 5 zombies
	return zombieCount > 5
}
