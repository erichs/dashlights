package signals

import (
	"context"
	"os"
)

// ProxyActiveSignal checks for active proxy settings
type ProxyActiveSignal struct {
	foundProxies []string
}

// NewProxyActiveSignal creates a ProxyActiveSignal.
func NewProxyActiveSignal() *ProxyActiveSignal {
	return &ProxyActiveSignal{}
}

// Name returns the human-readable name of the signal.
func (s *ProxyActiveSignal) Name() string {
	return "Man in the Middle"
}

// Emoji returns the emoji associated with the signal.
func (s *ProxyActiveSignal) Emoji() string {
	return "🕵️"
}

// Diagnostic returns a description of the detected proxy configuration.
func (s *ProxyActiveSignal) Diagnostic() string {
	if len(s.foundProxies) == 0 {
		return "Proxy detected in environment"
	}
	return "Proxy variables set: " + s.foundProxies[0]
}

// Remediation returns guidance on how to handle detected proxy settings.
func (s *ProxyActiveSignal) Remediation() string {
	return "Verify proxy is expected; unset if not needed"
}

// Check inspects proxy-related environment variables and reports if any are set.
func (s *ProxyActiveSignal) Check(ctx context.Context) bool {
	_ = ctx

	s.foundProxies = []string{}

	proxyVars := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"ALL_PROXY",
		"http_proxy",
		"https_proxy",
		"all_proxy",
	}

	for _, varName := range proxyVars {
		if val := os.Getenv(varName); val != "" {
			s.foundProxies = append(s.foundProxies, varName)
		}
	}

	return len(s.foundProxies) > 0
}
