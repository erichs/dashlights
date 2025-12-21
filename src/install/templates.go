package install

// Sentinel markers for idempotency detection.
const (
	SentinelBegin = "# BEGIN dashlights"
	SentinelEnd   = "# END dashlights"
)

// Shell prompt templates wrapped with sentinel markers.
const (
	// BashTemplate is the prompt integration for bash.
	BashTemplate = `# BEGIN dashlights
__dashlights_prompt() {
    local dl_out
    dl_out=$(dashlights 2>/dev/null)
    if [ -n "$dl_out" ]; then
        echo -n "$dl_out "
    fi
}
PS1='$(__dashlights_prompt)'"$PS1"
# END dashlights
`

	// ZshTemplate is the prompt integration for zsh.
	ZshTemplate = `# BEGIN dashlights
setopt prompt_subst
__dashlights_prompt() {
    local dl_out
    dl_out=$(dashlights 2>/dev/null)
    if [[ -n "$dl_out" ]]; then
        echo -n "$dl_out "
    fi
}
PROMPT='$(__dashlights_prompt)'"$PROMPT"
# END dashlights
`

	// FishTemplate is the prompt integration for fish.
	FishTemplate = `# BEGIN dashlights
function __dashlights_prompt --on-event fish_prompt
    set -l dl_out (dashlights 2>/dev/null)
    if test -n "$dl_out"
        echo -n "$dl_out "
    end
end
# END dashlights
`

	// P10kTemplate is the prompt segment function for Powerlevel10k.
	P10kTemplate = `# BEGIN dashlights
function prompt_dashlights() {
    local dl_out
    dl_out=$(dashlights 2>/dev/null)
    if [[ -n "$dl_out" ]]; then
        p10k segment -f 208 -t "$dl_out"
    fi
}
# END dashlights
`
)

// GetShellTemplate returns the appropriate template for a template type.
func GetShellTemplate(templateType TemplateType) string {
	switch templateType {
	case TemplateBash:
		return BashTemplate
	case TemplateZsh:
		return ZshTemplate
	case TemplateFish:
		return FishTemplate
	case TemplateP10k:
		return P10kTemplate
	default:
		return ""
	}
}

// Agent configuration templates.
const (
	// ClaudeHookJSON is the hook configuration to add to Claude's settings.
	// This is not the full file, but the structure to merge in.
	ClaudeHookJSON = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "dashlights --agentic"
          }
        ]
      }
    ]
  }
}`

	// CursorHookJSON is the hook configuration for Cursor.
	CursorHookJSON = `{
  "beforeShellExecution": {
    "command": "dashlights --agentic"
  }
}`
)

// DashlightsCommand is the command to run for agentic mode.
const DashlightsCommand = "dashlights --agentic"
