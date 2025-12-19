package agentic

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/erichs/dashlights/src/signals"
)

// AnalysisResult captures the complete Rule of Two analysis for a tool call.
type AnalysisResult struct {
	ToolName    string
	CapabilityA CapabilityResult // Untrustworthy inputs
	CapabilityB CapabilityResult // Sensitive access
	CapabilityC CapabilityResult // State change/external comms
	SignalHits  []string         // Which dashlights signals also triggered
}

// CapabilityCount returns how many capabilities were detected.
func (r *AnalysisResult) CapabilityCount() int {
	count := 0
	if r.CapabilityA.Detected {
		count++
	}
	if r.CapabilityB.Detected {
		count++
	}
	if r.CapabilityC.Detected {
		count++
	}
	return count
}

// ViolatesRuleOfTwo returns true if all three capabilities are detected.
func (r *AnalysisResult) ViolatesRuleOfTwo() bool {
	return r.CapabilityCount() >= 3
}

// AllReasons collects all detection reasons across capabilities.
func (r *AnalysisResult) AllReasons() []string {
	var reasons []string
	reasons = append(reasons, r.CapabilityA.Reasons...)
	reasons = append(reasons, r.CapabilityB.Reasons...)
	reasons = append(reasons, r.CapabilityC.Reasons...)
	return reasons
}

// CapabilityString returns a string like "A+B" or "A+B+C" for detected capabilities.
func (r *AnalysisResult) CapabilityString() string {
	var caps []string
	if r.CapabilityA.Detected {
		caps = append(caps, "A")
	}
	if r.CapabilityB.Detected {
		caps = append(caps, "B")
	}
	if r.CapabilityC.Detected {
		caps = append(caps, "C")
	}
	return strings.Join(caps, "+")
}

// Analyzer performs Rule of Two analysis on tool calls.
type Analyzer struct {
	// RunSignals controls whether to run dashlights signals for enhanced detection.
	RunSignals bool
	// SignalTimeout is the timeout for running signals (default 5ms).
	SignalTimeout time.Duration
}

// NewAnalyzer creates an Analyzer with default settings.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		RunSignals:    true,
		SignalTimeout: 5 * time.Millisecond,
	}
}

// Analyze performs Rule of Two analysis on a hook input.
func (a *Analyzer) Analyze(input *HookInput) *AnalysisResult {
	result := &AnalysisResult{
		ToolName: input.ToolName,
	}

	// Run heuristic detection for each capability
	result.CapabilityA = DetectCapabilityA(input.ToolName, input.ToolInput, input.Cwd)
	result.CapabilityB = DetectCapabilityB(input.ToolName, input.ToolInput)
	result.CapabilityC = DetectCapabilityC(input.ToolName, input.ToolInput)

	// Optionally run relevant signals for enhanced B-capability detection
	if a.RunSignals {
		signalHits := a.runRelevantSignals()
		result.SignalHits = signalHits

		// If signals detected sensitive issues, enhance B detection
		if len(signalHits) > 0 && !result.CapabilityB.Detected {
			result.CapabilityB.Detected = true
			for _, hit := range signalHits {
				result.CapabilityB.Reasons = append(result.CapabilityB.Reasons,
					"signal detected: "+hit)
			}
		}
	}

	return result
}

// runRelevantSignals runs a subset of dashlights signals relevant to agentic context.
// Returns names of signals that detected issues.
func (a *Analyzer) runRelevantSignals() []string {
	ctx, cancel := context.WithTimeout(context.Background(), a.SignalTimeout)
	defer cancel()

	// Only run signals relevant to detecting sensitive access (Capability B)
	relevantSignals := []signals.Signal{
		signals.NewNakedCredentialsSignal(),
		signals.NewDangerousTFVarSignal(),
		signals.NewProdPanicSignal(),
		signals.NewRootKubeContextSignal(),
		signals.NewAWSAliasHijackSignal(),
	}

	var hits []string
	for _, sig := range relevantSignals {
		// Check if signal is disabled
		disableVar := "DASHLIGHTS_DISABLE_" + strings.ToUpper(strings.ReplaceAll(sig.Name(), " ", "_"))
		if os.Getenv(disableVar) != "" {
			continue
		}

		if sig.Check(ctx) {
			hits = append(hits, sig.Name())
		}
	}

	return hits
}
