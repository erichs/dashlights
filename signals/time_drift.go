package signals

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TimeDriftSignal checks for clock synchronization issues
type TimeDriftSignal struct {
	driftSeconds int
}

func NewTimeDriftSignal() *TimeDriftSignal {
	return &TimeDriftSignal{}
}

func (s *TimeDriftSignal) Name() string {
	return "Time Drift"
}

func (s *TimeDriftSignal) Emoji() string {
	return "🕰️"
}

func (s *TimeDriftSignal) Diagnostic() string {
	return "System clock may be out of sync"
}

func (s *TimeDriftSignal) Remediation() string {
	return "Enable NTP synchronization or check time service"
}

func (s *TimeDriftSignal) Check(ctx context.Context) bool {
	// Try to check NTP sync status
	if runtime.GOOS == "linux" {
		return s.checkLinuxTimeSync(ctx)
	} else if runtime.GOOS == "darwin" {
		return s.checkMacOSTimeSync(ctx)
	}
	
	return false
}

func (s *TimeDriftSignal) checkLinuxTimeSync(ctx context.Context) bool {
	// Try timedatectl first (systemd systems)
	cmd := exec.CommandContext(ctx, "timedatectl", "status")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		// Check if NTP is synchronized
		if strings.Contains(outputStr, "NTP synchronized: no") ||
		   strings.Contains(outputStr, "System clock synchronized: no") {
			return true
		}
		return false
	}
	
	// Fallback: check if ntpd or chronyd is running
	// This is a simplified check
	return false
}

func (s *TimeDriftSignal) checkMacOSTimeSync(ctx context.Context) bool {
	// Check if time sync is enabled on macOS
	cmd := exec.CommandContext(ctx, "systemsetup", "-getusingnetworktime")
	output, err := cmd.Output()
	if err == nil {
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "off") {
			return true
		}
	}
	
	// Alternative: check sntp offset
	cmd = exec.CommandContext(ctx, "sntp", "-t", "1", "time.apple.com")
	output, err = cmd.Output()
	if err == nil {
		// Parse offset from output
		// Format: "time.apple.com: +0.123456 +/- 0.001234"
		outputStr := string(output)
		if strings.Contains(outputStr, "+") || strings.Contains(outputStr, "-") {
			// Extract offset value
			fields := strings.Fields(outputStr)
			for i, field := range fields {
				if (strings.HasPrefix(field, "+") || strings.HasPrefix(field, "-")) && i > 0 {
					offsetStr := strings.TrimPrefix(strings.TrimPrefix(field, "+"), "-")
					offset, err := strconv.ParseFloat(offsetStr, 64)
					if err == nil {
						// Flag if offset is more than 5 seconds
						if offset > 5.0 || offset < -5.0 {
							s.driftSeconds = int(offset)
							return true
						}
					}
					break
				}
			}
		}
	}
	
	return false
}

// Simple fallback: compare file modification time
func (s *TimeDriftSignal) checkFileTimestamp() bool {
	// Create a temp file and check if its timestamp is reasonable
	// This is a very basic check
	now := time.Now()
	
	// If system time is wildly off (more than 1 year), flag it
	if now.Year() < 2020 || now.Year() > 2030 {
		return true
	}
	
	return false
}

