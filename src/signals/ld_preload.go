package signals

import (
	"context"
	"os"
	"runtime"
)

// LDPreloadSignal checks for LD_PRELOAD or DYLD_INSERT_LIBRARIES
type LDPreloadSignal struct {
	varName string
	value   string
}

// NewLDPreloadSignal creates an LDPreloadSignal.
func NewLDPreloadSignal() *LDPreloadSignal {
	return &LDPreloadSignal{}
}

// Name returns the human-readable name of the signal.
func (s *LDPreloadSignal) Name() string {
	return "Trojan Horse"
}

// Emoji returns the emoji associated with the signal.
func (s *LDPreloadSignal) Emoji() string {
	return "🐴"
}

// Diagnostic returns a description of the LD_PRELOAD/DYLD_INSERT_LIBRARIES setting.
func (s *LDPreloadSignal) Diagnostic() string {
	return s.varName + " is set to: " + s.value
}

// Remediation returns guidance on how to safely unset preload injection variables.
func (s *LDPreloadSignal) Remediation() string {
	return "Unset " + s.varName + " unless intentionally debugging"
}

// Check inspects environment variables for LD_PRELOAD or DYLD_INSERT_LIBRARIES.
func (s *LDPreloadSignal) Check(ctx context.Context) bool {
	_ = ctx

	// Check LD_PRELOAD on Linux
	if runtime.GOOS == "linux" {
		if val := os.Getenv("LD_PRELOAD"); val != "" {
			s.varName = "LD_PRELOAD"
			s.value = val
			return true
		}
	}

	// Check DYLD_INSERT_LIBRARIES on macOS
	if runtime.GOOS == "darwin" {
		if val := os.Getenv("DYLD_INSERT_LIBRARIES"); val != "" {
			s.varName = "DYLD_INSERT_LIBRARIES"
			s.value = val
			return true
		}
	}

	return false
}
