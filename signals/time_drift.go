package signals

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// TimeDriftSignal detects drift between system time and filesystem mtime
// This can indicate desynchronized network drives or VMs with clock drift
type TimeDriftSignal struct{}

func NewTimeDriftSignal() Signal {
	return &TimeDriftSignal{}
}

func (s *TimeDriftSignal) Name() string {
	return "Time Drift Detected"
}

func (s *TimeDriftSignal) Emoji() string {
	return "⏰" // Alarm clock emoji
}

func (s *TimeDriftSignal) Diagnostic() string {
	return "System time and filesystem time are out of sync (network drive or VM clock drift)"
}

func (s *TimeDriftSignal) Remediation() string {
	return "Check NTP sync, VM time sync settings, or network drive mount options"
}

func (s *TimeDriftSignal) Check(ctx context.Context) bool {
	// Create a temporary file in the current directory
	tmpFile := filepath.Join(".", ".dashlights_time_check")
	
	// Record system time before creating file
	beforeTime := time.Now()
	
	// Create the file
	f, err := os.Create(tmpFile)
	if err != nil {
		// If we can't create a temp file, we can't check - return false
		return false
	}
	f.Close()
	
	// Ensure cleanup
	defer os.Remove(tmpFile)
	
	// Stat the file to get its mtime
	fileInfo, err := os.Stat(tmpFile)
	if err != nil {
		return false
	}
	
	// Get the file's modification time
	fileTime := fileInfo.ModTime()
	
	// Record system time after stat
	afterTime := time.Now()
	
	// Calculate the drift
	// The file's mtime should be between beforeTime and afterTime
	// Allow a small tolerance for filesystem timestamp granularity (typically 1-2 seconds)
	const toleranceSeconds = 5
	
	// Check if file time is too far in the past
	if fileTime.Before(beforeTime.Add(-toleranceSeconds * time.Second)) {
		return true
	}
	
	// Check if file time is too far in the future
	if fileTime.After(afterTime.Add(toleranceSeconds * time.Second)) {
		return true
	}
	
	return false
}

