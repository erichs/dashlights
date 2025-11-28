package signals

import (
	"context"
	"os"
)

// ProxyActiveSignal checks for active proxy settings
type ProxyActiveSignal struct {
	foundProxies []string
}

func NewProxyActiveSignal() *ProxyActiveSignal {
	return &ProxyActiveSignal{}
}

func (s *ProxyActiveSignal) Name() string {
	return "Man in the Middle"
}

func (s *ProxyActiveSignal) Emoji() string {
	return "🕵️"
}

func (s *ProxyActiveSignal) Diagnostic() string {
	if len(s.foundProxies) == 0 {
		return "Proxy detected in environment"
	}
	return "Proxy variables set: " + s.foundProxies[0]
}

func (s *ProxyActiveSignal) Remediation() string {
	return "Verify proxy is expected; unset if not needed"
}

func (s *ProxyActiveSignal) Check(ctx context.Context) bool {
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

