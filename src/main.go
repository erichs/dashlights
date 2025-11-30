package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/erichs/dashlights/src/signals"
	"github.com/fatih/color"
)

type dashlight struct {
	Name        string
	Glyph       string
	Diagnostic  string
	Color       *color.Color
	UnsetString string
}

var args struct {
	ObdMode   bool `arg:"-d,--obd,help:On-Board Diagnostics: display diagnostic info if provided."`
	ListMode  bool `arg:"-l,--list,help:List supported color attributes and emoji aliases."`
	ClearMode bool `arg:"-c,--clear,help:Shell code to clear set dashlights."`
}

func flexPrintf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format, args...)
}

func flexPrintln(w io.Writer, line string) {
	fmt.Fprintln(w, line)
}

func displayClearCodes(w io.Writer, lights *[]dashlight) {
	for _, light := range *lights {
		flexPrintln(w, light.UnsetString)
	}
}

var lights []dashlight

func init() {
	parseEnviron(os.Environ(), &lights)
}

func main() {
	arg.MustParse(&args)

	// Run security signal checks with a tight timeout for performance
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	allSignals := signals.GetAllSignals()
	results := signals.CheckAll(ctx, allSignals)

	display(os.Stdout, &lights, results)
}

func parseEnviron(environ []string, lights *[]dashlight) {
	for _, env := range environ {
		parseDashlightFromEnv(lights, env)
	}
}

func display(w io.Writer, lights *[]dashlight, results []signals.Result) {
	if args.ListMode {
		displayColorList(w)
		flexPrintln(w, "")
		displayEmojiList(w)
		return
	}
	if args.ClearMode {
		displayClearCodes(w, lights)
		return
	}

	// New default output: 🚨 {count} {DASHLIGHT_runes}
	if args.ObdMode {
		// Diagnostic mode: show detailed signal information
		displaySignalDiagnostics(w, results)
	} else {
		// Default mode: show siren, count, and DASHLIGHT runes
		displaySecurityStatus(w, results, lights)
	}
}

// displaySecurityStatus shows the default output: 🚨 {count} {DASHLIGHT_runes}
func displaySecurityStatus(w io.Writer, results []signals.Result, lights *[]dashlight) {
	// Count detected signals
	count := signals.CountDetected(results)

	// Only show siren if there are security issues
	if count > 0 {
		flexPrintf(w, "🚨 %d", count)
	}

	// Append DASHLIGHT_ runes if any
	if len(*lights) > 0 {
		if count > 0 {
			flexPrintf(w, " ")
		}
		for _, light := range *lights {
			flexPrintf(w, "%s", light.Glyph)
		}
	}

	flexPrintln(w, "")
}

// displaySignalDiagnostics shows detailed diagnostic information for detected signals
func displaySignalDiagnostics(w io.Writer, results []signals.Result) {
	detected := signals.GetDetected(results)

	if len(detected) == 0 {
		flexPrintln(w, "✅ No security issues detected")
		return
	}

	flexPrintln(w, "Security Issues Detected:")
	flexPrintln(w, "")

	for _, result := range detected {
		sig := result.Signal
		flexPrintf(w, "%s %s\n", sig.Emoji(), sig.Diagnostic())
		flexPrintf(w, "   → Fix: %s\n", sig.Remediation())
		flexPrintln(w, "")
	}
}

func displayDashlights(w io.Writer, lights *[]dashlight) {
	for _, light := range *lights {
		lamp := light.Color.SprintfFunc()("%s ", light.Glyph)
		flexPrintf(w, "%s ", lamp)
	}
	if len(*lights) > 0 {
		flexPrintln(w, "")
	}
}

func displayDiagnostics(w io.Writer, lights *[]dashlight) {
	flexPrintf(w, "\n-------- Diagnostics --------\n")
	for _, light := range *lights {
		lamp := light.Color.SprintfFunc()("%s ", light.Glyph)
		flexPrintf(w, "%s: %s - %s\n", lamp, light.Name, light.Diagnostic)
	}
}

func parseDashlightFromEnv(lights *[]dashlight, env string) {
	kv := strings.Split(env, "=")
	dashvar := kv[0]
	diagnostic := kv[1]
	if strings.Contains(dashvar, "DASHLIGHT_") {
		if diagnostic == "" {
			diagnostic = "No diagnostic info provided."
		}
		elements := strings.Split(dashvar, "_")
		if len(elements) < 3 {
			// dashvars must minimally be of form: DASHLIGHT_{name}_{utf8hex}
			return
		}
		// begin shifting elements off elements slice, ignore leading DASHLIGHT_ prefix
		name, elements := elements[1], elements[2:]
		hexstr, elements := elements[0], elements[1:]
		// Resolve emoji alias to hex code if applicable
		hexstr = resolveEmojiAlias(hexstr)
		glyph, err := utf8HexToString(string(hexstr))
		if err != nil {
			return
		}
		dashColor := color.New()
		// process any remaining elements as color additions
		for _, colorstr := range elements {
			dashColor.Add(colorMap[colorstr])
		}
		*lights = append(*lights, dashlight{
			Name:        name,
			Glyph:       glyph,
			Diagnostic:  diagnostic,
			Color:       dashColor,
			UnsetString: "unset " + dashvar,
		})
	}
}

func utf8HexToString(hex string) (string, error) {
	i, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		return "", err
	}
	return string(rune(i)), nil
}
