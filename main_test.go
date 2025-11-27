package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

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
	if 0 != len(lights) {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// missing utf8 hex string...
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO=foo")
	if 0 != len(lights) {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// invalid utf8 hex strings...
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO_ZZDA9=")
	if 0 != len(lights) {
		t.Error("Expected length of 0, got ", len(lights))
	}
	parseDashlightFromEnv(&lights, "DASHLIGHT_FOO_X=")
	if 0 != len(lights) {
		t.Error("Expected length of 0, got ", len(lights))
	}
	// invalid colormap codes are ignored...
	parseDashlightFromEnv(&lights, "DASHLIGHT_NOCODETEST_0021_NOTACODE=")
	if 1 != len(lights) {
		t.Error("Expected length of 1, got ", len(lights))
	}
	parseDashlightFromEnv(&lights, "DASHLIGHT_VALIDCODETEST_0021_BGWHITE=")
	if 2 != len(lights) {
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
	if "*color.Color" != typeof(light.Color) {
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
	if args.ObdMode {
		t.Error("Diagnostic mode should not start enabled!")
	}
	if args.ListMode {
		t.Error("List color mode should not start enabled!")
	}
	if args.ClearMode {
		t.Error("Clear mode should not start enabled!")
	}
}

func TestListColorModeDisplay(t *testing.T) {
	args.ListMode = true
	defer func() { args.ListMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_LCM_0021=")

	display(&b, &lights)
	if !strings.Contains(b.String(), "BGWHITE") {
		t.Errorf("List mode should contain color attribute keys, found: %s", b.String())
	}
}

func TestClearModeDisplay(t *testing.T) {
	args.ClearMode = true
	defer func() { args.ClearMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_CM_0021=")

	display(&b, &lights)
	expectStr := "unset DASHLIGHT_CM_0021"
	if !strings.Contains(b.String(), expectStr) {
		t.Errorf("Clear mode should '%s', found: %s", expectStr, b.String())
	}
}

func TestDiagModeDisplay(t *testing.T) {
	args.ObdMode = true
	defer func() { args.ObdMode = false }()

	var b bytes.Buffer
	lights := make([]dashlight, 0)
	parseDashlightFromEnv(&lights, "DASHLIGHT_DM_0021=bar diagnostic")

	display(&b, &lights)
	if !strings.Contains(b.String(), lights[0].Glyph) {
		t.Errorf("Diag mode should lead with dashlight display containing '%s', found: '%s'", lights[0].Glyph, b.String())
	}

	expectStr := " DM - bar diagnostic"
	if !strings.Contains(b.String(), expectStr) {
		t.Errorf("Expected to see '%s' in:\n%s", expectStr, b.String())
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
