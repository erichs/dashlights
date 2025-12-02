package main

//go:generate go run gen_repo_url.go

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/erichs/dashlights/src/signals"
	"github.com/fatih/color"
)

// Version information (set by GoReleaser via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type dashlight struct {
	Name        string
	Glyph       string
	Diagnostic  string
	Color       *color.Color
	UnsetString string
}

type cliArgs struct {
	DetailsMode     bool `arg:"-d,--details,help:Show detailed diagnostic information for detected issues."`
	VerboseMode     bool `arg:"-v,--verbose,help:Verbose mode: show documentation links in diagnostic output."`
	ListCustomMode  bool `arg:"-l,--list-custom,help:List supported color attributes and emoji aliases for custom lights."`
	ClearCustomMode bool `arg:"-c,--clear-custom,help:Shell code to clear custom DASHLIGHT_ environment variables."`
}

// Version returns the version string for --version flag
func (cliArgs) Version() string {
	return fmt.Sprintf("dashlights %s (commit: %s, built: %s)", version, commit, date)
}

var args cliArgs
var lights []dashlight

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

func main() {
	// start watchdog timer for 10.1ms, this will exit the program and 'fail open'
	// if any security signals are not respecting context cancellation for any reason
	watchdog := time.AfterFunc(time.Duration(10.1*float64(time.Millisecond)), func() {
		// fmt.Println("Timeout!") TODO: add debug logging if this occurs
		os.Exit(0)
	})
	defer watchdog.Stop() // Cancel watchdog on normal completion

	arg.MustParse(&args)

	// Run security signal checks with a tight timeout for performance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var (
		wg      sync.WaitGroup
		results []signals.Result
		envRaw  []string
	)

	wg.Add(2)
	// Parse DASHLIGHT_ environment variables
	go func() {
		defer wg.Done()
		envRaw = os.Environ()
		// we pass a local slice to avoid locking: we'll assign to global
		// 'lights' after parsing
		localLights := []dashlight{}
		parseEnviron(envRaw, &localLights)
		lights = localLights
	}()

	// Security signal checks
	go func() {
		defer wg.Done()
		allSignals := signals.GetAllSignals()
		results = signals.CheckAll(ctx, allSignals)
	}()

	wg.Wait() // Wait for both goroutines to complete,
	// if we timed out, the watchdog will exit here

	display(os.Stdout, &lights, results)
}

func parseEnviron(environ []string, lights *[]dashlight) {
	for _, env := range environ {
		parseDashlightFromEnv(lights, env)
	}
}

func display(w io.Writer, lights *[]dashlight, results []signals.Result) {
	if args.ListCustomMode {
		displayColorList(w)
		flexPrintln(w, "")
		displayEmojiList(w)
		return
	}
	if args.ClearCustomMode {
		displayClearCodes(w, lights)
		return
	}

	// New default output: 🚨 {count} {DASHLIGHT_runes}
	if args.DetailsMode {
		// Details mode: show detailed signal information
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
		// Use gray color for count to be legible on both light and dark backgrounds
		gray := color.New(color.FgHiBlack)
		flexPrintf(w, "🚨 %s", gray.Sprintf("%d", count))
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

// signalTypeToFilename converts a signal type name to its documentation filename
// Example: "*signals.AWSAliasHijackSignal" -> "aws_alias_hijack"
func signalTypeToFilename(sig signals.Signal) string {
	// Get the type name using reflection
	typeName := reflect.TypeOf(sig).String()

	// Remove package prefix and pointer indicator
	// Example: "*signals.AWSAliasHijackSignal" -> "AWSAliasHijackSignal"
	re := regexp.MustCompile(`\*?signals\.(.+)Signal`)
	matches := re.FindStringSubmatch(typeName)
	if len(matches) < 2 {
		return ""
	}

	name := matches[1]

	// Convert from PascalCase to snake_case
	// Handle consecutive uppercase letters (e.g., "AWS" -> "aws", not "a_w_s")
	var result strings.Builder
	runes := []rune(name)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Add underscore before uppercase letter if:
		// 1. Not the first character
		// 2. Previous character is lowercase OR
		// 3. Next character is lowercase (end of acronym)
		if i > 0 && r >= 'A' && r <= 'Z' {
			prevIsLower := runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextIsLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'

			if prevIsLower || nextIsLower {
				result.WriteRune('_')
			}
		}

		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
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

		// Show documentation link in verbose mode
		if args.VerboseMode {
			filename := signalTypeToFilename(sig)
			if filename != "" {
				docURL := fmt.Sprintf("%s/blob/main/docs/signals/%s.md", RepositoryURL, filename)
				flexPrintf(w, "   📖 Documentation: %s\n", docURL)
			}
		}

		flexPrintln(w, "")
	}

	// Show breadcrumb footer in non-verbose mode
	if !args.VerboseMode {
		flexPrintln(w, "💡 Tip: Use -v flag for detailed documentation links")
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
