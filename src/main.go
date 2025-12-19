package main

//go:generate go run gen_repo_url.go

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/erichs/dashlights/src/agentic"
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

type debugResult struct {
	Result   signals.Result
	Duration time.Duration
}

type cliArgs struct {
	DetailsMode     bool `arg:"-d,--details,help:Show detailed diagnostic information for detected issues."`
	VerboseMode     bool `arg:"-v,--verbose,help:Verbose mode: show documentation links in diagnostic output."`
	ListCustomMode  bool `arg:"-l,--list-custom,help:List supported color attributes and emoji aliases for custom lights."`
	ClearCustomMode bool `arg:"-c,--clear-custom,help:Shell code to clear custom DASHLIGHT_ environment variables."`
	DebugMode       bool `arg:"--debug,help:Debug mode: disable timeouts and show detailed execution timing."`
	AgenticMode     bool `arg:"--agentic,help:Agentic mode for AI coding assistants (reads JSON from stdin)."`
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
	arg.MustParse(&args)

	// Agentic mode: completely different execution path for AI coding assistant hooks
	if args.AgenticMode {
		os.Exit(runAgenticMode())
	}

	startTime := time.Now()
	var envParseStart, envParseEnd time.Time
	var signalsStart, signalsEnd time.Time

	// Phase 1: Parse DASHLIGHT_ environment variables FIRST (microseconds)
	// This ensures custom emojis are always available, even on timeout
	if args.DebugMode {
		envParseStart = time.Now()
	}
	envRaw := os.Environ()
	localLights := []dashlight{}
	parseEnviron(envRaw, &localLights)
	lights = localLights
	if args.DebugMode {
		envParseEnd = time.Now()
	}

	// Phase 2: Set up context with timeout for signal checks
	var ctx context.Context
	var cancel context.CancelFunc
	if args.DebugMode {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		flexPrintln(os.Stderr, "🐛 Debug mode: watchdog timer disabled")
		flexPrintf(os.Stderr, "🐛 Debug mode: timeout set to 30 seconds\n\n")
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	}
	defer cancel()

	// Safety watchdog - backup in case signals don't respect context cancellation
	// Set slightly longer than context timeout to allow graceful collection first
	if !args.DebugMode {
		watchdog := time.AfterFunc(time.Duration(10.5*float64(time.Millisecond)), func() {
			// Emergency exit with timeout code - signals didn't respect context
			os.Exit(124)
		})
		defer watchdog.Stop()
	}

	// Phase 3: Run signal checks (returns partial results on timeout)
	var results []signals.Result
	var debugResults []debugResult
	var completed bool

	if args.DebugMode {
		signalsStart = time.Now()
	}
	allSignals := signals.GetAllSignals()
	if args.DebugMode {
		results, debugResults, completed = checkAllWithTiming(ctx, allSignals)
		signalsEnd = time.Now()
		totalDuration := time.Since(startTime)
		displayDebugInfo(os.Stderr, envParseStart, envParseEnd, signalsStart, signalsEnd, totalDuration, &lights, results, debugResults)
	} else {
		results, completed = signals.CheckAll(ctx, allSignals)
	}

	// Phase 4: Display results (partial or complete) with custom emojis
	display(os.Stdout, &lights, results)

	// Phase 5: Exit with timeout code if incomplete (and not in debug mode)
	if !completed && !args.DebugMode {
		os.Exit(124)
	}
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
		displaySignalDiagnostics(w, results, lights)
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
func displaySignalDiagnostics(w io.Writer, results []signals.Result, lights *[]dashlight) {
	detected := signals.GetDetected(results)

	if len(detected) == 0 && len(*lights) == 0 {
		flexPrintln(w, "✅ No security issues detected")
		return
	}

	if len(detected) > 0 {
		flexPrintln(w, "Security Issues Detected:")
		flexPrintln(w, "")

		for _, result := range detected {
			sig := result.Signal
			flexPrintf(w, "%s %s\n", sig.Emoji(), sig.Diagnostic())
			flexPrintf(w, "   → Fix: %s\n", sig.Remediation())

			// Show verbose remediation and documentation link in verbose mode
			if args.VerboseMode {
				// Check if the signal implements VerboseRemediator interface
				if vr, ok := sig.(signals.VerboseRemediator); ok {
					if verboseRem := vr.VerboseRemediation(); verboseRem != "" {
						flexPrintln(w, "")
						flexPrintf(w, "   🔧 %s\n", verboseRem)
					}
				}

				filename := signalTypeToFilename(sig)
				if filename != "" {
					docURL := fmt.Sprintf("%s/blob/main/docs/signals/%s.md", RepositoryURL, filename)
					flexPrintf(w, "   📖 Documentation: %s\n", docURL)
				}
			}

			flexPrintln(w, "")
		}
	}

	// Show custom DASHLIGHT_ emojis if any
	if len(*lights) > 0 {
		for _, light := range *lights {
			flexPrintf(w, "%s %s - %s\n", light.Glyph, light.Name, light.Diagnostic)
		}
		flexPrintln(w, "")
	}

	// Show breadcrumb footer in non-verbose mode (only if there were security signals)
	if len(detected) > 0 && !args.VerboseMode {
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
	kv := strings.SplitN(env, "=", 2)
	if len(kv) < 2 {
		return
	}
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

// indexedDebugResult pairs a debug result with its index for channel-based collection
type indexedDebugResult struct {
	idx         int
	result      signals.Result
	debugResult debugResult
}

// checkAllWithTiming runs all signal checks concurrently and tracks timing for each.
// Returns (results, debugResults, completed) where completed is false if timeout occurred.
func checkAllWithTiming(ctx context.Context, sigs []signals.Signal) ([]signals.Result, []debugResult, bool) {
	if len(sigs) == 0 {
		return []signals.Result{}, []debugResult{}, true
	}

	results := make([]signals.Result, len(sigs))
	debugResults := make([]debugResult, len(sigs))
	resultChan := make(chan indexedDebugResult, len(sigs))

	for i, sig := range sigs {
		go func(idx int, s signals.Signal) {
			start := time.Now()

			// Use a panic recovery to ensure one bad signal doesn't crash everything
			defer func() {
				if r := recover(); r != nil {
					duration := time.Since(start)
					res := signals.Result{
						Signal:   s,
						Detected: false,
						Error:    nil,
					}
					resultChan <- indexedDebugResult{
						idx:    idx,
						result: res,
						debugResult: debugResult{
							Result:   res,
							Duration: duration,
						},
					}
				}
			}()

			detected := s.Check(ctx)
			duration := time.Since(start)

			res := signals.Result{
				Signal:   s,
				Detected: detected,
				Error:    nil,
			}
			resultChan <- indexedDebugResult{
				idx:    idx,
				result: res,
				debugResult: debugResult{
					Result:   res,
					Duration: duration,
				},
			}
		}(i, sig)
	}

	// Collect results until context expires or all complete
	completed := 0
	for completed < len(sigs) {
		select {
		case ir := <-resultChan:
			results[ir.idx] = ir.result
			debugResults[ir.idx] = ir.debugResult
			completed++
		case <-ctx.Done():
			// Drain any results that arrived just before/during timeout
		drainLoop:
			for {
				select {
				case ir := <-resultChan:
					results[ir.idx] = ir.result
					debugResults[ir.idx] = ir.debugResult
					completed++
				default:
					break drainLoop
				}
			}

			// Mark unreceived signals with empty results
			for i, sig := range sigs {
				if results[i].Signal == nil {
					results[i] = signals.Result{Signal: sig, Detected: false, Error: nil}
					debugResults[i] = debugResult{Result: results[i], Duration: 0}
				}
			}
			// Note: Goroutines that haven't completed will finish and send to the
			// buffered channel (non-blocking). Since this is a CLI that exits
			// immediately after displaying results, os.Exit() cleans up any
			// remaining goroutines - no explicit cancellation needed.
			return results, debugResults, false // Partial results
		}
	}

	return results, debugResults, true // All complete
}

// runAgenticMode handles the --agentic flag for AI coding assistant integration.
// It reads a tool call JSON from stdin, performs critical threat and Rule of Two
// analysis, and outputs appropriate JSON/exit code. Supports both Claude Code
// (PreToolUse hook) and Cursor (beforeShellExecution hook).
func runAgenticMode() int {
	const maxAgenticInputBytes = 1 * 1024 * 1024

	// Read JSON from stdin first (needed for agent detection)
	input, err := io.ReadAll(io.LimitReader(os.Stdin, maxAgenticInputBytes+1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		return 1
	}
	if len(input) > maxAgenticInputBytes {
		fmt.Fprintf(os.Stderr, "Error: input exceeds %d bytes\n", maxAgenticInputBytes)
		return 1
	}

	// Handle empty input gracefully
	if len(input) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input provided on stdin")
		return 1
	}

	// Detect agent type from environment, fall back to input format detection
	agentType := agentic.DetectAgent()
	if agentType == agentic.AgentUnknown {
		agentType = agentic.DetectAgentFromInput(input)
	}

	// Check if disabled - output format depends on agent type
	if agentic.IsDisabled() {
		return outputDisabled(agentType)
	}

	// Parse hook input based on agent type
	var hookInput *agentic.HookInput
	switch agentType {
	case agentic.AgentCursor:
		hookInput, err = agentic.ParseCursorInput(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing Cursor input: %v\n", err)
			return 1
		}
	default:
		// Claude Code format (default)
		hookInput = &agentic.HookInput{}
		if err := json.Unmarshal(input, hookInput); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			return 1
		}
	}

	// Check for critical threats BEFORE Rule of Two analysis
	// These bypass the capability scoring and are handled immediately
	if threat := agentic.DetectCriticalThreat(hookInput); threat != nil {
		return outputThreat(agentType, threat)
	}

	// Analyze for Rule of Two violations
	analyzer := agentic.NewAnalyzer()
	result := analyzer.Analyze(hookInput)

	// Generate output based on agent type
	return outputResult(agentType, result)
}

// outputDisabled outputs the appropriate "disabled" response for the agent type.
func outputDisabled(agentType agentic.AgentType) int {
	switch agentType {
	case agentic.AgentCursor:
		jsonOut, exitCode, _ := agentic.GenerateCursorDisabledOutput()
		fmt.Println(string(jsonOut))
		return exitCode
	default:
		// Claude Code format
		output := agentic.HookOutput{
			HookSpecificOutput: &agentic.HookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       "allow",
				PermissionDecisionReason: "Rule of Two: disabled",
			},
		}
		jsonOut, err := json.Marshal(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
			return 1
		}
		fmt.Println(string(jsonOut))
		return 0
	}
}

// outputThreat outputs the appropriate threat response for the agent type.
func outputThreat(agentType agentic.AgentType, threat *agentic.CriticalThreat) int {
	var jsonOut []byte
	var exitCode int
	var stderrMsg string

	switch agentType {
	case agentic.AgentCursor:
		jsonOut, exitCode, stderrMsg = agentic.GenerateCursorThreatOutput(threat)
	default:
		// Claude Code format
		var output *agentic.HookOutput
		output, exitCode, stderrMsg = agentic.GenerateThreatOutput(threat)
		if output != nil {
			var err error
			jsonOut, err = json.Marshal(output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
				return 1
			}
		}
	}

	if exitCode == 2 {
		fmt.Fprintln(os.Stderr, stderrMsg)
	}
	if jsonOut != nil {
		fmt.Println(string(jsonOut))
	}
	return exitCode
}

// outputResult outputs the appropriate analysis result for the agent type.
func outputResult(agentType agentic.AgentType, result *agentic.AnalysisResult) int {
	var jsonOut []byte
	var exitCode int
	var stderrMsg string

	switch agentType {
	case agentic.AgentCursor:
		jsonOut, exitCode, stderrMsg = agentic.GenerateCursorOutput(result)
	default:
		// Claude Code format
		var output *agentic.HookOutput
		output, exitCode, stderrMsg = agentic.GenerateOutput(result)
		if output != nil {
			var err error
			jsonOut, err = json.Marshal(output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
				return 1
			}
		}
	}

	if exitCode == 2 {
		fmt.Fprintln(os.Stderr, stderrMsg)
	}
	if jsonOut != nil {
		fmt.Println(string(jsonOut))
	}
	return exitCode
}

// displayDebugInfo outputs detailed debug information to stderr
func displayDebugInfo(w io.Writer, envStart, envEnd, sigStart, sigEnd time.Time, total time.Duration, lights *[]dashlight, results []signals.Result, debugResults []debugResult) {
	flexPrintln(w, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	flexPrintln(w, "🐛 DEBUG INFORMATION")
	flexPrintln(w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Phase timing
	envDuration := envEnd.Sub(envStart)
	sigDuration := sigEnd.Sub(sigStart)

	flexPrintln(w, "⏱️  PHASE TIMING:")
	flexPrintf(w, "   Environment parsing: %v\n", envDuration)
	flexPrintf(w, "   Signal checks:       %v\n", sigDuration)
	flexPrintf(w, "   Total execution:     %v\n\n", total)

	// Environment parsing results
	flexPrintln(w, "📦 ENVIRONMENT PARSING:")
	if len(*lights) == 0 {
		flexPrintln(w, "   No DASHLIGHT_ variables found\n")
	} else {
		flexPrintf(w, "   Found %d custom dashlight(s):\n", len(*lights))
		for _, light := range *lights {
			flexPrintf(w, "      %s %s - %s\n", light.Glyph, light.Name, light.Diagnostic)
		}
		flexPrintln(w, "")
	}

	// Signal check results
	detectedCount := signals.CountDetected(results)
	flexPrintf(w, "🔍 SIGNAL CHECK RESULTS (%d detected, %d total):\n\n", detectedCount, len(results))

	// Sort debug results by duration (slowest first)
	sortedDebug := make([]debugResult, len(debugResults))
	copy(sortedDebug, debugResults)
	// Simple bubble sort for top performers
	for i := 0; i < len(sortedDebug); i++ {
		for j := i + 1; j < len(sortedDebug); j++ {
			if sortedDebug[j].Duration > sortedDebug[i].Duration {
				sortedDebug[i], sortedDebug[j] = sortedDebug[j], sortedDebug[i]
			}
		}
	}

	// Show all signals with timing
	for _, dr := range sortedDebug {
		status := "  "
		if dr.Result.Detected {
			status = "🚨"
		}
		flexPrintf(w, "   %s %-35s %8v\n", status, dr.Result.Signal.Name(), dr.Duration)
	}

	// Performance summary
	flexPrintln(w, "\n📊 PERFORMANCE SUMMARY:")
	if len(sortedDebug) >= 3 {
		flexPrintln(w, "   Top 3 slowest checks:")
		for i := 0; i < 3 && i < len(sortedDebug); i++ {
			flexPrintf(w, "      %d. %-35s %v\n", i+1, sortedDebug[i].Result.Signal.Name(), sortedDebug[i].Duration)
		}
	}

	// Check if any signals exceeded thresholds
	slow := 0
	for _, dr := range debugResults {
		if dr.Duration > 5*time.Millisecond {
			slow++
		}
	}
	if slow > 0 {
		flexPrintf(w, "\n   ⚠️  %d signal(s) exceeded 5ms threshold\n", slow)
	}

	flexPrintln(w, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
