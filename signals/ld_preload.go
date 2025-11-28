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

func NewLDPreloadSignal() *LDPreloadSignal {
	return &LDPreloadSignal{}
}

func (s *LDPreloadSignal) Name() string {
	return "Trojan Horse"
}

func (s *LDPreloadSignal) Emoji() string {
	return "🐴"
}

func (s *LDPreloadSignal) Diagnostic() string {
	return s.varName + " is set to: " + s.value
}

func (s *LDPreloadSignal) Remediation() string {
	return "Unset " + s.varName + " unless intentionally debugging"
}

func (s *LDPreloadSignal) Check(ctx context.Context) bool {
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

