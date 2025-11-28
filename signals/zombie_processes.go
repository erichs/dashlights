package signals

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ZombieProcessesSignal checks for excessive zombie processes
type ZombieProcessesSignal struct {
	count int
}

func NewZombieProcessesSignal() *ZombieProcessesSignal {
	return &ZombieProcessesSignal{}
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

	// Read /proc/stat to count zombie processes
	data, err := os.ReadFile("/proc/stat")
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
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is numeric (PID)
		name := entry.Name()
		if len(name) == 0 || name[0] < '0' || name[0] > '9' {
			continue
		}

		// Read /proc/PID/stat
		statPath := "/proc/" + name + "/stat"
		statData, err := os.ReadFile(statPath)
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
