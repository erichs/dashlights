package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/erichs/dashlights/src/signals"
	"github.com/fatih/color"
)

func typeof(v interface{}) string {
	return reflect.TypeOf(v).String()
}

func captureRunAgenticMode(t *testing.T, stdin string) (int, string, string) {
	t.Helper()

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	if stdin != "" {
		if _, err := stdinW.WriteString(stdin); err != nil {
			t.Fatalf("Failed to write stdin: %v", err)
		}
	}
	stdinW.Close()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Stderr = stderrW

	exitCode := runAgenticMode()

	stdoutW.Close()
	stderrW.Close()

	stdoutBytes, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}
	stderrBytes, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("Failed to read stderr: %v", err)
	}

	stdinR.Close()
	stdoutR.Close()
	stderrR.Close()

	os.Stdin = oldStdin
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return exitCode, string(stdoutBytes), string(stderrBytes)
}

func TestDisplayCodes(t *testing.T) {
	lights := make([]dashlight, 0)
	lights = append(lights, dashlight{
		Name:        "foo",
		Glyph:       "X",
		Diagnostic:  "",
		Color:       color.New(),
		UnsetString: "unset foo",
	})
	var b bytes.Buffer
	displayClearCodes(&b, &lights)
	if b.String() != "unset foo\n" {
		t.Error("Expected 'unset foo\n', got ", b.String())
	}
	lights = append(lights, dashlight{
		Name:        "bar",
		Glyph:       "Y",
		Diagnostic:  "",
		Color:       color.New(),
		UnsetString: "unset bar",
	})
	b.Reset()
	displayClearCodes(&b, &lights)
	//	fmt.Printf("output: %s\n", "X"+b.String()+"X")
	if b.String() != "unset foo\nunset bar\n" {
		t.Error("Expected 'unset foo\nunset bar\n', got ", b.String())
	}
}

func TestParseDashlightFromEnv(t *testing.T) {
	lights := make([]dashlight, 0)
	// missing namespace prefix...
	parseDashlightFromEnv(&lights, "FOO_2112_BGWHITE=foo")
	if len(lights) != 0 {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// missing utf8 hex string...
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO=foo")
	if len(lights) != 0 {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// invalid utf8 hex strings...
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO_ZZDA9=")
	if len(lights) != 0 {
		t.Error("Expected length of 0, got ", len(lights))
	}
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO_X=")
	if len(lights) != 0 {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// invalid colormap codes are ignored...
	parseDashlightFromEnv(&lights, "DASHLIGHT_NOCODETEST_0021_NOTACODE=")
	if len(lights) != 1 {
		t.Error("Expected length of 1, got ", len(lights))
	}
	parseDashlightFromEnv(&lights, "DASHLIGHT_VALIDCODETEST_0021_BGWHITE=")
	if len(lights) != 2 {
		t.Error("Expected length of 2, got ", len(lights))
	}
	light := lights[1]
	if light.Name != "VALIDCODETEST" {
		t.Error("Expected Name of 'VALIDCODETEST', got ", light.Name)
	}
	if light.Glyph != "!" {
		t.Error("Expected Glyph of '!', got ", light.Glyph)
	}
	if light.Diagnostic != "No diagnostic info provided." {
		t.Error("Expected default diagnostic string, got ", light.Diagnostic)
	}
	if typeof(light.Color) != "*color.Color" {
		t.Error("Expected color to be type *color.Color, got ", typeof(light.Color))
	}
	if light.UnsetString != "unset DASHLIGHT_VALIDCODETEST_0021_BGWHITE" {
		t.Error("Expected valid unset string, got ", light.UnsetString)
	}
}

func TestDisplayDiagnostics(t *testing.T) {
	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_SZYZYGY_2112_BGWHITE=foo diagnostic")
	displayDiagnostics(&b, &lights)
	expectedStr := " SZYZYGY - foo diagnostic"
	// contains detailed diagnostics for test light
	if !strings.Contains(b.String(), expectedStr) {
		t.Errorf("Expected to see '%s' in:\n%s", expectedStr, b.String())
	}
	// contains diagnostics header
	if !strings.Contains(b.String(), "- Diagnostics -") {
		t.Errorf("Expected to see '%s' in:\n%s", "- Diagnostics -", b.String())
	}
}

func TestDisplayDashlights(t *testing.T) {
	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_DISPLAY_0021=")
	parseDashlightFromEnv(&lights, "DASHLIGHT_BAR_25A6=")
	displayDashlights(&b, &lights)
	spcCount := strings.Count(b.String(), " ")
	if spcCount != 4 {
		t.Errorf("Expected %d spaces in output, got %d", 2, spcCount)
	}
}

func TestDefaultFlagStates(t *testing.T) {
	if args.DetailsMode {
		t.Error("Details mode should not start enabled!")
	}
	if args.ListCustomMode {
		t.Error("List custom mode should not start enabled!")
	}
	if args.ClearCustomMode {
		t.Error("Clear custom mode should not start enabled!")
	}
}

func TestVersion(t *testing.T) {
	// Test that Version() method exists and returns expected format
	versionStr := args.Version()

	// Should contain "dashlights"
	if !strings.Contains(versionStr, "dashlights") {
		t.Errorf("Version string should contain 'dashlights', got: %s", versionStr)
	}

	// Should contain version info (even if it's "dev")
	if !strings.Contains(versionStr, version) {
		t.Errorf("Version string should contain version '%s', got: %s", version, versionStr)
	}

	// Should contain commit info
	if !strings.Contains(versionStr, "commit:") {
		t.Errorf("Version string should contain 'commit:', got: %s", versionStr)
	}

	// Should contain build date info
	if !strings.Contains(versionStr, "built:") {
		t.Errorf("Version string should contain 'built:', got: %s", versionStr)
	}
}

func TestListColorModeDisplay(t *testing.T) {
	args.ListCustomMode = true
	defer func() { args.ListCustomMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_LCM_0021=")

	display(&b, &lights, nil)
	if !strings.Contains(b.String(), "BGWHITE") {
		t.Errorf("List custom mode should contain color attribute keys, found: %s", b.String())
	}
}

func TestClearModeDisplay(t *testing.T) {
	args.ClearCustomMode = true
	defer func() { args.ClearCustomMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_CM_0021=")

	display(&b, &lights, nil)
	expectStr := "unset DASHLIGHT_CM_0021"
	if !strings.Contains(b.String(), expectStr) {
		t.Errorf("Clear custom mode should '%s', found: %s", expectStr, b.String())
	}
}

func TestDiagModeDisplay(t *testing.T) {
	args.DetailsMode = true
	defer func() { args.DetailsMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_DM_0021=bar diagnostic")

	// Pass empty results for this test
	display(&b, &lights, []signals.Result{})

	// In diagnostic mode with custom lights, should show the custom light
	if !strings.Contains(b.String(), "DM - bar diagnostic") {
		t.Errorf("Expected to see custom light 'DM - bar diagnostic' in:\n%s", b.String())
	}
}

func TestParseEnviron(t *testing.T) {
	environ := []string{
		"LC_CTYPE=en_US.UTF-8",
		"DASHLIGHT_FOO_0021=",
		"PAGER=less",
	}
	lights := make([]dashlight, 0)
	parseEnviron(environ, &lights)
	if len(lights) != 1 {
		t.Error("Failed to parse from environ key=val strings.")
	}
}

func TestEmojiAliasInParsing(t *testing.T) {
	lights := make([]dashlight, 0)

	// Test emoji alias without color modifiers
	parseDashlightFromEnv(&lights, "DASHLIGHT_FIX_WRENCH=Tool needed")
	if len(lights) != 1 {
		t.Fatalf("Expected 1 light, got %d", len(lights))
	}
	if lights[0].Name != "FIX" {
		t.Errorf("Expected Name 'FIX', got '%s'", lights[0].Name)
	}
	if lights[0].Glyph != "🔧" {
		t.Errorf("Expected Glyph '🔧', got '%s'", lights[0].Glyph)
	}
	if lights[0].Diagnostic != "Tool needed" {
		t.Errorf("Expected Diagnostic 'Tool needed', got '%s'", lights[0].Diagnostic)
	}

	// Test emoji alias with color modifiers
	lights = make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_ATTACH_PAPERCLIP_BGBLUE=File attached")
	if len(lights) != 1 {
		t.Fatalf("Expected 1 light, got %d", len(lights))
	}
	if lights[0].Name != "ATTACH" {
		t.Errorf("Expected Name 'ATTACH', got '%s'", lights[0].Name)
	}
	if lights[0].Glyph != "📎" {
		t.Errorf("Expected Glyph '📎', got '%s'", lights[0].Glyph)
	}

	// Test multiple emoji aliases
	lights = make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_SUCCESS_CHECKMARK_FGGREEN=All good")
	parseDashlightFromEnv(&lights, "DASHLIGHT_ERROR_CROSSMARK_FGRED=Failed")
	parseDashlightFromEnv(&lights, "DASHLIGHT_INFO_LIGHTBULB=Tip")
	if len(lights) != 3 {
		t.Fatalf("Expected 3 lights, got %d", len(lights))
	}
	if lights[0].Glyph != "✅" {
		t.Errorf("Expected first glyph '✅', got '%s'", lights[0].Glyph)
	}
	if lights[1].Glyph != "❌" {
		t.Errorf("Expected second glyph '❌', got '%s'", lights[1].Glyph)
	}
	if lights[2].Glyph != "💡" {
		t.Errorf("Expected third glyph '💡', got '%s'", lights[2].Glyph)
	}

	// Test that hex codes still work
	lights = make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_LEGACY_1F527=Old style")
	if len(lights) != 1 {
		t.Fatalf("Expected 1 light, got %d", len(lights))
	}
	if lights[0].Glyph != "🔧" {
		t.Errorf("Expected Glyph '🔧' from hex code, got '%s'", lights[0].Glyph)
	}

	// Test emoji alias with empty name (double underscore)
	lights = make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT__LINK_FGCYAN=Connected")
	if len(lights) != 1 {
		t.Fatalf("Expected 1 light, got %d", len(lights))
	}
	if lights[0].Name != "" {
		t.Errorf("Expected empty Name, got '%s'", lights[0].Name)
	}
	if lights[0].Glyph != "🔗" {
		t.Errorf("Expected Glyph '🔗', got '%s'", lights[0].Glyph)
	}
}

func TestSignalTypeToFilename(t *testing.T) {
	tests := []struct {
		name     string
		signal   signals.Signal
		expected string
	}{
		{
			name:     "AWS Alias Hijack Signal",
			signal:   signals.NewAWSAliasHijackSignal(),
			expected: "aws_alias_hijack",
		},
		{
			name:     "Debug Enabled Signal",
			signal:   signals.NewDebugEnabledSignal(),
			expected: "debug_enabled",
		},
		{
			name:     "Docker Socket Signal",
			signal:   signals.NewDockerSocketSignal(),
			expected: "docker_socket",
		},
		{
			name:     "History Permissions Signal",
			signal:   signals.NewHistoryPermissionsSignal(),
			expected: "history_permissions",
		},
		{
			name:     "Naked Credentials Signal",
			signal:   signals.NewNakedCredentialsSignal(),
			expected: "naked_credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := signalTypeToFilename(tt.signal)
			if result != tt.expected {
				t.Errorf("signalTypeToFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDisplaySignalDiagnosticsNoIssues(t *testing.T) {
	var b bytes.Buffer
	results := []signals.Result{}
	emptyLights := []dashlight{}

	displaySignalDiagnostics(&b, results, &emptyLights)

	expected := "✅ No security issues detected"
	if !strings.Contains(b.String(), expected) {
		t.Errorf("Expected to see '%s' in:\n%s", expected, b.String())
	}
}

func TestDisplaySignalDiagnosticsWithIssues(t *testing.T) {
	var b bytes.Buffer
	emptyLights := []dashlight{}

	// Create a mock signal result
	sig := signals.NewDockerSocketSignal()
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}

	// Test non-verbose mode
	args.VerboseMode = false
	displaySignalDiagnostics(&b, results, &emptyLights)

	// Should contain the diagnostic message
	if !strings.Contains(b.String(), "Security Issues Detected:") {
		t.Errorf("Expected to see 'Security Issues Detected:' in:\n%s", b.String())
	}

	// Should contain the breadcrumb in non-verbose mode
	if !strings.Contains(b.String(), "Use -v flag for detailed documentation links") {
		t.Errorf("Expected to see breadcrumb message in:\n%s", b.String())
	}

	// Should NOT contain documentation link in non-verbose mode
	if strings.Contains(b.String(), "Documentation:") {
		t.Errorf("Should not see documentation link in non-verbose mode:\n%s", b.String())
	}
}

func TestDisplaySignalDiagnosticsVerboseMode(t *testing.T) {
	var b bytes.Buffer
	emptyLights := []dashlight{}

	// Create a mock signal result
	sig := signals.NewDockerSocketSignal()
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}

	// Test verbose mode
	args.VerboseMode = true
	defer func() { args.VerboseMode = false }()

	displaySignalDiagnostics(&b, results, &emptyLights)

	// Should contain the diagnostic message
	if !strings.Contains(b.String(), "Security Issues Detected:") {
		t.Errorf("Expected to see 'Security Issues Detected:' in:\n%s", b.String())
	}

	// Should contain documentation link in verbose mode
	if !strings.Contains(b.String(), "📖 Documentation:") {
		t.Errorf("Expected to see documentation link in:\n%s", b.String())
	}

	// Should contain the correct documentation URL
	if !strings.Contains(b.String(), "docs/signals/docker_socket.md") {
		t.Errorf("Expected to see correct documentation URL in:\n%s", b.String())
	}

	// Should NOT contain breadcrumb in verbose mode
	if strings.Contains(b.String(), "Use -v flag") {
		t.Errorf("Should not see breadcrumb message in verbose mode:\n%s", b.String())
	}
}

func TestDisplaySecurityStatusNoIssues(t *testing.T) {
	var b bytes.Buffer
	results := []signals.Result{}
	lights := make([]dashlight, 0)

	displaySecurityStatus(&b, results, &lights)

	// Should only have a newline, no siren
	if strings.Contains(b.String(), "🚨") {
		t.Errorf("Should not show siren when no issues detected:\n%s", b.String())
	}
}

func TestDisplaySecurityStatusWithIssues(t *testing.T) {
	var b bytes.Buffer

	sig := signals.NewDockerSocketSignal()
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}
	lights := make([]dashlight, 0)

	displaySecurityStatus(&b, results, &lights)

	// Should show siren with count
	if !strings.Contains(b.String(), "🚨 1") {
		t.Errorf("Expected to see '🚨 1' in:\n%s", b.String())
	}
}

func TestDisplaySecurityStatusWithIssuesAndLights(t *testing.T) {
	var b bytes.Buffer

	sig := signals.NewDockerSocketSignal()
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_TEST_0021=test")

	displaySecurityStatus(&b, results, &lights)

	// Should show siren with count and dashlight
	if !strings.Contains(b.String(), "🚨 1") {
		t.Errorf("Expected to see '🚨 1' in:\n%s", b.String())
	}
	if !strings.Contains(b.String(), "!") {
		t.Errorf("Expected to see dashlight glyph in:\n%s", b.String())
	}
}

func TestDisplaySecurityStatusOnlyLights(t *testing.T) {
	var b bytes.Buffer

	results := []signals.Result{}
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_TEST_0021=test")

	displaySecurityStatus(&b, results, &lights)

	// Should show only dashlight, no siren
	if strings.Contains(b.String(), "🚨") {
		t.Errorf("Should not show siren when no security issues:\n%s", b.String())
	}
	if !strings.Contains(b.String(), "!") {
		t.Errorf("Expected to see dashlight glyph in:\n%s", b.String())
	}
}

// mockSignal is a simple signal for testing
type mockSignal struct {
	name       string
	detected   bool
	checkDelay time.Duration
}

func (m *mockSignal) Check(ctx context.Context) bool {
	if m.checkDelay > 0 {
		select {
		case <-time.After(m.checkDelay):
			return m.detected
		case <-ctx.Done():
			return false
		}
	}
	return m.detected
}

func (m *mockSignal) Name() string        { return m.name }
func (m *mockSignal) Emoji() string       { return "🔍" }
func (m *mockSignal) Diagnostic() string  { return "Test diagnostic" }
func (m *mockSignal) Remediation() string { return "Test remediation" }

type panicSignal struct {
	name string
}

func (p *panicSignal) Check(_ context.Context) bool { panic("boom") }
func (p *panicSignal) Name() string                 { return p.name }
func (p *panicSignal) Emoji() string                { return "💥" }
func (p *panicSignal) Diagnostic() string           { return "Panic signal" }
func (p *panicSignal) Remediation() string          { return "Handle panic" }

func TestCheckAllWithTimingEmptySignals(t *testing.T) {
	ctx := context.Background()
	results, debugResults, completed := checkAllWithTiming(ctx, []signals.Signal{})

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
	if len(debugResults) != 0 {
		t.Errorf("Expected 0 debug results, got %d", len(debugResults))
	}
	if !completed {
		t.Error("Expected completed to be true for empty signals")
	}
}

func TestCheckAllWithTimingSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sigs := []signals.Signal{
		&mockSignal{name: "Signal1", detected: true},
		&mockSignal{name: "Signal2", detected: false},
		&mockSignal{name: "Signal3", detected: true},
	}

	results, debugResults, completed := checkAllWithTiming(ctx, sigs)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	if len(debugResults) != 3 {
		t.Errorf("Expected 3 debug results, got %d", len(debugResults))
	}
	if !completed {
		t.Error("Expected completed to be true")
	}

	// Verify detection status
	detectedCount := 0
	for _, r := range results {
		if r.Detected {
			detectedCount++
		}
	}
	if detectedCount != 2 {
		t.Errorf("Expected 2 detected signals, got %d", detectedCount)
	}

	// Verify debug results have duration
	for _, dr := range debugResults {
		if dr.Result.Signal == nil {
			t.Error("Debug result should have signal reference")
		}
	}
}

func TestCheckAllWithTimingTimeout(t *testing.T) {
	// Create a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sigs := []signals.Signal{
		&mockSignal{name: "FastSignal", detected: true, checkDelay: 0},
		&mockSignal{name: "SlowSignal", detected: true, checkDelay: 100 * time.Millisecond},
	}

	results, debugResults, completed := checkAllWithTiming(ctx, sigs)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if len(debugResults) != 2 {
		t.Errorf("Expected 2 debug results, got %d", len(debugResults))
	}
	if completed {
		t.Error("Expected completed to be false due to timeout")
	}
}

func TestCheckAllWithTimingPanicRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sigs := []signals.Signal{
		&panicSignal{name: "PanicSignal"},
	}

	results, debugResults, completed := checkAllWithTiming(ctx, sigs)

	if !completed {
		t.Error("Expected completed to be true even with panic recovery")
	}
	if len(results) != 1 || len(debugResults) != 1 {
		t.Fatalf("Expected 1 result/debug result, got %d/%d", len(results), len(debugResults))
	}
	if results[0].Signal == nil || results[0].Signal.Name() != "PanicSignal" {
		t.Errorf("Expected recovered signal to be present, got %+v", results[0].Signal)
	}
}

func TestDisplayDebugInfoNoLights(t *testing.T) {
	var b bytes.Buffer

	envStart := time.Now()
	envEnd := envStart.Add(100 * time.Microsecond)
	sigStart := envEnd
	sigEnd := sigStart.Add(5 * time.Millisecond)
	total := sigEnd.Sub(envStart)

	emptyLights := []dashlight{}
	results := []signals.Result{
		{Signal: &mockSignal{name: "TestSignal1", detected: false}, Detected: false},
		{Signal: &mockSignal{name: "TestSignal2", detected: true}, Detected: true},
	}
	debugResults := []debugResult{
		{Result: results[0], Duration: 1 * time.Millisecond},
		{Result: results[1], Duration: 2 * time.Millisecond},
	}

	displayDebugInfo(&b, envStart, envEnd, sigStart, sigEnd, total, &emptyLights, results, debugResults)

	output := b.String()

	// Check for expected sections
	if !strings.Contains(output, "DEBUG INFORMATION") {
		t.Errorf("Expected DEBUG INFORMATION header in:\n%s", output)
	}
	if !strings.Contains(output, "PHASE TIMING") {
		t.Errorf("Expected PHASE TIMING section in:\n%s", output)
	}
	if !strings.Contains(output, "No DASHLIGHT_ variables found") {
		t.Errorf("Expected 'No DASHLIGHT_ variables found' in:\n%s", output)
	}
	if !strings.Contains(output, "SIGNAL CHECK RESULTS") {
		t.Errorf("Expected SIGNAL CHECK RESULTS section in:\n%s", output)
	}
	if !strings.Contains(output, "1 detected") {
		t.Errorf("Expected '1 detected' in:\n%s", output)
	}
}

func TestDisplayDebugInfoWithLights(t *testing.T) {
	var b bytes.Buffer

	envStart := time.Now()
	envEnd := envStart.Add(100 * time.Microsecond)
	sigStart := envEnd
	sigEnd := sigStart.Add(5 * time.Millisecond)
	total := sigEnd.Sub(envStart)

	lights := []dashlight{}
	parseDashlightFromEnv(&lights, "DASHLIGHT_TEST_0021=test diagnostic")

	results := []signals.Result{}
	debugResults := []debugResult{}

	displayDebugInfo(&b, envStart, envEnd, sigStart, sigEnd, total, &lights, results, debugResults)

	output := b.String()

	// Check for custom dashlight in output
	if !strings.Contains(output, "Found 1 custom dashlight") {
		t.Errorf("Expected 'Found 1 custom dashlight' in:\n%s", output)
	}
	if !strings.Contains(output, "TEST") {
		t.Errorf("Expected 'TEST' dashlight name in:\n%s", output)
	}
}

func TestDisplayDebugInfoSlowSignals(t *testing.T) {
	var b bytes.Buffer

	now := time.Now()

	lights := []dashlight{}
	results := []signals.Result{
		{Signal: &mockSignal{name: "SlowSignal1"}, Detected: false},
		{Signal: &mockSignal{name: "SlowSignal2"}, Detected: false},
		{Signal: &mockSignal{name: "SlowSignal3"}, Detected: false},
		{Signal: &mockSignal{name: "FastSignal"}, Detected: false},
	}
	debugResults := []debugResult{
		{Result: results[0], Duration: 6 * time.Millisecond},
		{Result: results[1], Duration: 7 * time.Millisecond},
		{Result: results[2], Duration: 8 * time.Millisecond},
		{Result: results[3], Duration: 1 * time.Millisecond},
	}

	displayDebugInfo(&b, now, now.Add(time.Microsecond), now, now.Add(time.Millisecond), time.Millisecond, &lights, results, debugResults)

	output := b.String()

	// Check for slow signal warning
	if !strings.Contains(output, "exceeded 5ms threshold") {
		t.Errorf("Expected slow signal warning in:\n%s", output)
	}
	// Check for top 3 slowest
	if !strings.Contains(output, "Top 3 slowest") {
		t.Errorf("Expected 'Top 3 slowest' in:\n%s", output)
	}
}

// nonMatchingSignal doesn't match the expected type name pattern
type nonMatchingSignal struct{}

func (n *nonMatchingSignal) Check(_ context.Context) bool { return false }
func (n *nonMatchingSignal) Name() string                 { return "NonMatching" }
func (n *nonMatchingSignal) Emoji() string                { return "❓" }
func (n *nonMatchingSignal) Diagnostic() string           { return "Test" }
func (n *nonMatchingSignal) Remediation() string          { return "Test" }

func TestSignalTypeToFilenameNoMatch(t *testing.T) {
	// This signal type name won't match "*signals.XxxSignal" pattern
	sig := &nonMatchingSignal{}
	result := signalTypeToFilename(sig)

	if result != "" {
		t.Errorf("Expected empty string for non-matching signal type, got: %s", result)
	}
}

// mockVerboseSignal implements VerboseRemediator interface
type mockVerboseSignal struct {
	mockSignal
	verboseRemediation string
}

func (m *mockVerboseSignal) VerboseRemediation() string { return m.verboseRemediation }

func TestDisplaySignalDiagnosticsVerboseRemediator(t *testing.T) {
	var b bytes.Buffer

	// Create a mock signal that implements VerboseRemediator
	sig := &mockVerboseSignal{
		mockSignal:         mockSignal{name: "VerboseTest", detected: true},
		verboseRemediation: "Run this command: fix-it-all",
	}
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}
	emptyLights := []dashlight{}

	// Enable verbose mode
	args.VerboseMode = true
	defer func() { args.VerboseMode = false }()

	displaySignalDiagnostics(&b, results, &emptyLights)

	output := b.String()

	// Should contain the verbose remediation
	if !strings.Contains(output, "🔧 Run this command: fix-it-all") {
		t.Errorf("Expected verbose remediation in:\n%s", output)
	}
}

func TestDisplaySignalDiagnosticsEmptyVerboseRemediation(t *testing.T) {
	var b bytes.Buffer

	// Create a mock signal with empty verbose remediation
	sig := &mockVerboseSignal{
		mockSignal:         mockSignal{name: "EmptyVerbose", detected: true},
		verboseRemediation: "",
	}
	results := []signals.Result{
		{Signal: sig, Detected: true},
	}
	emptyLights := []dashlight{}

	// Enable verbose mode
	args.VerboseMode = true
	defer func() { args.VerboseMode = false }()

	displaySignalDiagnostics(&b, results, &emptyLights)

	output := b.String()

	// Should NOT contain the verbose remediation prefix when empty
	if strings.Contains(output, "🔧 \n") {
		t.Errorf("Should not show empty verbose remediation in:\n%s", output)
	}
}

func TestRunAgenticModeDisabled(t *testing.T) {
	t.Setenv("DASHLIGHTS_DISABLE_AGENTIC", "1")

	// Need to provide valid input since stdin is read before checking disabled
	input := `{"tool_name":"Read","tool_input":{"file_path":"test.txt"}}`
	exitCode, stdout, stderr := captureRunAgenticMode(t, input)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "\"permissionDecision\":\"allow\"") {
		t.Errorf("Expected allow decision in stdout, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Rule of Two: disabled") {
		t.Errorf("Expected disabled reason in stdout, got: %s", stdout)
	}
}

func TestRunAgenticModeEmptyInput(t *testing.T) {
	t.Setenv("DASHLIGHTS_DISABLE_AGENTIC", "")

	exitCode, stdout, stderr := captureRunAgenticMode(t, "")

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "no input provided") {
		t.Errorf("Expected no input error, got: %s", stderr)
	}
}

func TestRunAgenticModeInvalidJSON(t *testing.T) {
	t.Setenv("DASHLIGHTS_DISABLE_AGENTIC", "")

	exitCode, stdout, stderr := captureRunAgenticMode(t, "{bad")

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Error parsing JSON") {
		t.Errorf("Expected JSON parsing error, got: %s", stderr)
	}
}

func TestRunAgenticModeCriticalThreatBlock(t *testing.T) {
	t.Setenv("DASHLIGHTS_DISABLE_AGENTIC", "")

	input := `{"tool_name":"Write","tool_input":{"file_path":"CLAUDE.md","content":"x"}}`
	exitCode, stdout, stderr := captureRunAgenticMode(t, input)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Blocked: Attempted write to agent configuration") {
		t.Errorf("Expected blocked message, got: %s", stderr)
	}
}

func TestRunAgenticModeCriticalThreatAsk(t *testing.T) {
	t.Setenv("DASHLIGHTS_DISABLE_AGENTIC", "")
	t.Setenv("DASHLIGHTS_AGENTIC_MODE", "ask")

	input := "{\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"echo \\u200B\"}}"
	exitCode, stdout, stderr := captureRunAgenticMode(t, input)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "\"permissionDecision\":\"ask\"") {
		t.Errorf("Expected ask decision in stdout, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Invisible Unicode detected") {
		t.Errorf("Expected invisible unicode reason, got: %s", stdout)
	}
}
