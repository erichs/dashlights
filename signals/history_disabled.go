package signals

import (
	"context"
	"os"
)

// HistoryDisabledSignal checks if shell history is disabled
type HistoryDisabledSignal struct {
	reason string
}

func NewHistoryDisabledSignal() *HistoryDisabledSignal {
	return &HistoryDisabledSignal{}
}

func (s *HistoryDisabledSignal) Name() string {
	return "Blind Spot"
}

func (s *HistoryDisabledSignal) Emoji() string {
	return "🕶️"
}

func (s *HistoryDisabledSignal) Diagnostic() string {
	return s.reason
}

func (s *HistoryDisabledSignal) Remediation() string {
	return "Re-enable shell history for audit trail and incident response"
}

func (s *HistoryDisabledSignal) Check(ctx context.Context) bool {
	// Check if HISTFILE is unset or set to /dev/null
	histfile := os.Getenv("HISTFILE")
	if histfile == "/dev/null" {
		s.reason = "HISTFILE set to /dev/null (history disabled)"
		return true
	}
	
	// Note: We can't reliably detect if HISTFILE is unset vs empty string
	// in Go's os.Getenv, so we check HISTCONTROL instead
	
	// Check if HISTCONTROL contains ignorespace or ignoreboth
	histcontrol := os.Getenv("HISTCONTROL")
	if histcontrol != "" {
		if histcontrol == "ignorespace" || histcontrol == "ignoreboth" {
			s.reason = "HISTCONTROL set to '" + histcontrol + "' (commands with leading space ignored)"
			return true
		}
	}
	
	return false
}

