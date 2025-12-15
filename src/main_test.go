package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/erichs/dashlights/src/signals"
	"github.com/fatih/color"
)

func typeof(v interface{}) string {
	return reflect.TypeOf(v).String()
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
