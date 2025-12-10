package signals

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInsecureCurlPipeSignal_Metadata(t *testing.T) {
	sig := NewInsecureCurlPipeSignal().(*InsecureCurlPipeSignal)

	if sig.Name() != "Insecure Curl Pipe" {
		t.Errorf("expected Name 'Insecure Curl Pipe', got %q", sig.Name())
	}

	if sig.Emoji() == "" {
		t.Error("expected non-empty Emoji")
	}

	fallback := sig.Diagnostic()
	if !strings.Contains(fallback, "curl | bash or curl | sh") {
		t.Errorf("unexpected fallback diagnostic: %q", fallback)
	}

	rem := sig.Remediation()
	if !strings.Contains(rem, "Avoid piping curl directly") {
		t.Errorf("unexpected remediation: %q", rem)
	}
}

func TestInsecureCurlPipeSignal_NoShellEnv(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("SHELL", oldShell)
	}()

	os.Setenv("HOME", t.TempDir())
	os.Unsetenv("SHELL")

	sig := NewInsecureCurlPipeSignal()
	if sig.Check(context.Background()) {
		t.Error("expected false when SHELL is unset")
	}
}

func TestInsecureCurlPipeSignal_NoHistoryFile(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("SHELL", oldShell)
	}()

	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	os.Setenv("SHELL", "/bin/bash")

	sig := NewInsecureCurlPipeSignal()
	if sig.Check(context.Background()) {
		t.Error("expected false when history file does not exist")
	}
}

func TestInsecureCurlPipeSignal_NoMatchInLastThreeLines(t *testing.T) {
	cleanup := writeHistory(t, "bash", []string{
		"curl https://example.com/install.sh | bash", // far in history
		"echo safe1",
		"echo safe2",
		"echo safe3",
	})
	defer cleanup()

	sig := NewInsecureCurlPipeSignal()
	if sig.Check(context.Background()) {
		t.Error("expected false when only older history lines contain curl | bash")
	}
}

func TestInsecureCurlPipeSignal_DetectsCurlBashInRecentHistory(t *testing.T) {
	cleanup := writeHistory(t, "bash", []string{
		"echo before",
		"curl https://example.com/install.sh | bash",
		"echo after",
	})
	defer cleanup()

	sig := NewInsecureCurlPipeSignal().(*InsecureCurlPipeSignal)
	if !sig.Check(context.Background()) {
		t.Fatal("expected true when recent history contains curl | bash")
	}

	diag := sig.Diagnostic()
	if !strings.Contains(diag, "curl https://example.com/install.sh | bash") {
		t.Errorf("diagnostic did not include offending command: %q", diag)
	}
}

func TestInsecureCurlPipeSignal_DetectsCurlShWithFlags(t *testing.T) {
	cleanup := writeHistory(t, "zsh", []string{
		": 1700000000:0;curl -sSL https://sh.rustup.rs | sh",
	})
	defer cleanup()

	sig := NewInsecureCurlPipeSignal()
	if !sig.Check(context.Background()) {
		t.Error("expected true for curl -sSL ... | sh in zsh history format")
	}
}

func TestInsecureCurlPipeSignal_LargeHistoryTail(t *testing.T) {
	// Construct a history file larger than the 128KB tail window and ensure
	// we still correctly detect an insecure curl pipe in the final commands.
	lines := make([]string, 0, 20010)
	for i := 0; i < 20000; i++ {
		lines = append(lines, "echo safe")
	}
	lines = append(lines,
		"echo before",
		"curl https://example.com/install.sh | bash",
		"echo after",
	)

	cleanup := writeHistory(t, "bash", lines)
	defer cleanup()

	sig := NewInsecureCurlPipeSignal()
	if !sig.Check(context.Background()) {
		t.Error("expected true when large history ends with curl | bash")
	}
}

func TestInsecureCurlPipeSignal_ContextCancelled(t *testing.T) {
	cleanup := writeHistory(t, "bash", []string{
		"echo one",
		"echo two",
		"echo three",
	})
	defer cleanup()

	sig := NewInsecureCurlPipeSignal()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before scanning

	if sig.Check(ctx) {
		t.Error("expected false when context is cancelled")
	}
}

func TestTruncateHistoryLine(t *testing.T) {
	short := "echo hi"
	if got := truncateHistoryLine(short); got != short {
		t.Errorf("expected short line unchanged, got %q", got)
	}

	long := strings.Repeat("a", 200)
	got := truncateHistoryLine(long)
	if len(got) >= len(long) {
		t.Errorf("expected truncated line to be shorter, got len=%d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncated line to end with ellipsis, got %q", got)
	}
}

// writeHistory creates a temporary HOME and history file for the given shell.
func writeHistory(t *testing.T, shellName string, lines []string) func() {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "."+shellName+"_history")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write history file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	oldShell := os.Getenv("SHELL")
	os.Setenv("HOME", tmp)
	os.Setenv("SHELL", "/bin/"+shellName)

	return func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("SHELL", oldShell)
	}
}

func TestInsecureCurlPipeSignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_INSECURE_CURL_PIPE", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_INSECURE_CURL_PIPE")

	signal := NewInsecureCurlPipeSignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
