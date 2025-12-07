package signals

import (
	"context"
	"os"
	"testing"
)

func TestProxyActiveSignal_Name(t *testing.T) {
	signal := NewProxyActiveSignal()
	if signal.Name() != "Man in the Middle" {
		t.Errorf("Expected 'Man in the Middle', got '%s'", signal.Name())
	}
}

func TestProxyActiveSignal_Emoji(t *testing.T) {
	signal := NewProxyActiveSignal()
	if signal.Emoji() != "🕵️" {
		t.Errorf("Expected '🕵️', got '%s'", signal.Emoji())
	}
}

func TestProxyActiveSignal_Diagnostic_NoProxies(t *testing.T) {
	signal := NewProxyActiveSignal()
	signal.foundProxies = []string{}
	expected := "Proxy detected in environment"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestProxyActiveSignal_Diagnostic_WithProxies(t *testing.T) {
	signal := NewProxyActiveSignal()
	signal.foundProxies = []string{"HTTP_PROXY", "HTTPS_PROXY"}
	expected := "Proxy variables set: HTTP_PROXY"
	if signal.Diagnostic() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Diagnostic())
	}
}

func TestProxyActiveSignal_Remediation(t *testing.T) {
	signal := NewProxyActiveSignal()
	expected := "Verify proxy is expected; unset if not needed"
	if signal.Remediation() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, signal.Remediation())
	}
}

func TestProxyActiveSignal_Check_NoProxy(t *testing.T) {
	// Save and restore env vars
	proxyVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"}
	oldVals := make(map[string]string)
	for _, v := range proxyVars {
		oldVals[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range oldVals {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	signal := NewProxyActiveSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when no proxy variables are set")
	}
}

func TestProxyActiveSignal_Check_HTTPProxy(t *testing.T) {
	// Save and restore env var
	oldProxy := os.Getenv("HTTP_PROXY")
	defer func() {
		if oldProxy != "" {
			os.Setenv("HTTP_PROXY", oldProxy)
		} else {
			os.Unsetenv("HTTP_PROXY")
		}
	}()

	os.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")

	signal := NewProxyActiveSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when HTTP_PROXY is set")
	}

	found := false
	for _, v := range signal.foundProxies {
		if v == "HTTP_PROXY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected HTTP_PROXY in foundProxies")
	}
}

func TestProxyActiveSignal_Check_LowercaseProxy(t *testing.T) {
	// Save and restore env var
	oldProxy := os.Getenv("http_proxy")
	defer func() {
		if oldProxy != "" {
			os.Setenv("http_proxy", oldProxy)
		} else {
			os.Unsetenv("http_proxy")
		}
	}()

	os.Setenv("http_proxy", "http://proxy.example.com:8080")

	signal := NewProxyActiveSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when http_proxy is set")
	}

	found := false
	for _, v := range signal.foundProxies {
		if v == "http_proxy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected http_proxy in foundProxies")
	}
}

func TestProxyActiveSignal_Check_MultipleProxies(t *testing.T) {
	// Save and restore env vars
	oldHTTP := os.Getenv("HTTP_PROXY")
	oldHTTPS := os.Getenv("HTTPS_PROXY")
	defer func() {
		if oldHTTP != "" {
			os.Setenv("HTTP_PROXY", oldHTTP)
		} else {
			os.Unsetenv("HTTP_PROXY")
		}
		if oldHTTPS != "" {
			os.Setenv("HTTPS_PROXY", oldHTTPS)
		} else {
			os.Unsetenv("HTTPS_PROXY")
		}
	}()

	os.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
	os.Setenv("HTTPS_PROXY", "https://proxy.example.com:8443")

	signal := NewProxyActiveSignal()
	ctx := context.Background()

	if !signal.Check(ctx) {
		t.Error("Expected true when multiple proxy variables are set")
	}

	if len(signal.foundProxies) < 2 {
		t.Errorf("Expected at least 2 proxies found, got %d", len(signal.foundProxies))
	}
}

func TestProxyActiveSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_PROXY_ACTIVE", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_PROXY_ACTIVE")

	signal := NewProxyActiveSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
