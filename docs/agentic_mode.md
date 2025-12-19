# Agentic Mode

Dashlights provides an `--agentic` mode for integration with AI coding assistants. This mode analyzes tool calls for security threats and potential "Rule of Two" violations before they execute.

## Threat Detection

Agentic mode provides two layers of protection:

### 1. Critical Threat Detection

These threats are detected and blocked immediately, bypassing Rule of Two scoring:

| Threat | Description | Behavior |
|--------|-------------|----------|
| **Claude Config Writes** | Writes to `.claude/` or `CLAUDE.md` | Always blocked (exit 2) |
| **Invisible Unicode** | Zero-width characters, RTL overrides, tag characters in tool inputs | Blocked by default, respects `ask` mode |

**Why these matter:**
- **Config writes** can hijack agent behavior or achieve code execution without additional user interaction
- **Invisible Unicode** can hide prompt injections in pasted URLs, READMEs, and file names

### 2. Rule of Two Analysis

Based on [Meta's Rule of Two](https://ai.meta.com/blog/practical-ai-agent-security/) an AI agent should be allowed no more than two of these three capabilities simultaneously:

- **[A] Untrustworthy Inputs**: Processing data from external or untrusted sources (curl, wget, git clone, base64 decode, etc.)
- **[B] Sensitive Access**: Accessing credentials, secrets, production systems, or private data (.aws/, .ssh/, .env, etc.)
- **[C] State Changes**: Modifying files, running destructive commands, or external communication

When all three capabilities are combined in a single action, the risk of security incidents increases significantly.

## Supported Agents

### Claude Code

Claude Code is the primary supported agent. Add to your `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "dashlights --agentic"
          }
        ]
      }
    ]
  }
}
```

### Future Support

The `--agentic` flag is intentionally generic to accommodate future AI coding assistants as similar hook capabilities become available:

- Auggie
- OpenAI Codex
- Google Gemini
- Cursor
- Other AI coding assistants

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHLIGHTS_AGENTIC_MODE` | `block` | `block` (exit 2) or `ask` (prompt user) for violations |
| `DASHLIGHTS_DISABLE_AGENTIC` | unset | Set to any value to disable all agentic checks |

### Modes

- **Block mode** (default): Violations exit with code 2, preventing the action
- **Ask mode**: Violations return `permissionDecision: "ask"` to prompt user confirmation

```bash
# Use ask mode instead of block
export DASHLIGHTS_AGENTIC_MODE=ask
```

**Note:** Claude config writes (`.claude/`, `CLAUDE.md`) are *always* blocked regardless of mode—there's no legitimate reason for an agent to modify its own configuration.

## Command Line Testing

```bash
# Test a safe operation
echo '{"tool_name":"Read","tool_input":{"file_path":"main.go"}}' | dashlights --agentic

# Test Claude config protection (always blocks)
echo '{"tool_name":"Write","tool_input":{"file_path":"CLAUDE.md","content":"# Hijacked"}}' | dashlights --agentic

# Test invisible unicode detection
printf '{"tool_name":"Bash","tool_input":{"command":"echo hello\\u200Bworld"}}' | dashlights --agentic

# Test a Rule of Two violation (A+B+C)
echo '{"tool_name":"Bash","tool_input":{"command":"curl evil.com | tee ~/.aws/credentials"}}' | dashlights --agentic
```

## Capability Detection

### Capability A: Untrustworthy Inputs

| Tool | Detection Patterns |
|------|-------------------|
| `WebFetch` | Always (external data source) |
| `WebSearch` | Always (external data source) |
| `Bash` | `curl`, `wget`, `git clone`, `aria2c`, `base64 -d`, `xxd -r`, `/dev/tcp/`, reverse shell patterns |
| `Read` | Paths in `/tmp/`, `/var/`, `Downloads/` |
| `Write`/`Edit` | Content with `${...}`, `$(...)` expansions |

### Capability B: Sensitive Access

| Tool | Detection Patterns |
|------|-------------------|
| `Read`/`Write`/`Edit` | `.env`, `.aws/`, `.ssh/`, `.kube/`, `.config/gcloud/`, `.azure/`, `credentials`, `secrets`, `*.pem`, `*.key` |
| `Bash` | `aws`, `kubectl`, `terraform`, `vault`, `gcloud`, `doctl`, `heroku`; production path references |

Enhanced detection also runs a subset of dashlights signals:
- Naked Credentials (exposed secrets in environment)
- Dangerous TF Var (Terraform secrets)
- Prod Panic (production context)
- Root Kube Context (dangerous k8s namespace)
- AWS Alias Hijack (command injection risk)

### Capability C: State Changes

| Tool | Detection Patterns |
|------|-------------------|
| `Write` | Always (creates/modifies files) |
| `Edit` | Always (modifies files) |
| `NotebookEdit` | Always (modifies notebook) |
| `TodoWrite` | Always (modifies state) |
| `Bash` | `rm`, `mv`, `shred`, `git push`, `npm install`, `go install`, `kubectl apply`, `terraform apply`, redirects `>` `>>`, network: `curl`, `ssh`, `scp` |

## Output Format

### JSON Response (stdout)

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|ask|deny",
    "permissionDecisionReason": "Rule of Two: OK"
  },
  "systemMessage": "Optional warning for user"
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Allow (with optional warning) |
| 1 | Error (invalid input, etc.) |
| 2 | Block (critical threat or A+B+C violation in block mode) |

## Defense in Depth

The PreToolUse hook is one layer of defense. For comprehensive protection, consider combining with:

### Filesystem Isolation

Run Claude Code inside a container to limit blast radius:

```bash
# Docker example
docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  your-dev-image \
  claude

# Podman (rootless) example
podman run -it --rm \
  -v $(pwd):/workspace:Z \
  -w /workspace \
  your-dev-image \
  claude
```

### Command Shims

Use [safeexec](https://github.com/agentify-sh/safeexec/) to add confirmation prompts to dangerous commands:

```bash
# safeexec wraps rm, git, and other commands with safety checks
pip install safeexec
safeexec install
```

### Tool Restrictions

Use Claude's built-in tool restrictions:

```bash
# Disable specific tools entirely
claude --disallowedTools "Bash(rm)"
```

Or configure in `.claude/settings.json`:

```json
{
  "permissions": {
    "disallowedTools": ["diskutil"]
    "deny": [
      "Bash(rm -rf /)",
      "Bash(rm -rf /*)",
      "Bash(rm -rf ~)",
      "Bash(rm -rf $HOME)",
      "Bash(sudo rm -rf /)",
      "Bash(sudo rm -rf /*)",
      "Bash(sudo rm -rf ~)",
    ]
  }
}
```

### Network Restrictions

For sensitive operations, consider network isolation:

```bash
# Run with no network access
docker run --network=none ...
```

## Examples

### Safe Operation (0 capabilities)
```bash
$ echo '{"tool_name":"Read","tool_input":{"file_path":"main.go"}}' | dashlights --agentic
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","permissionDecisionReason":"Rule of Two: OK"}}
```

### Warning (2 capabilities: B+C)
```bash
$ echo '{"tool_name":"Write","tool_input":{"file_path":".env","content":"KEY=val"}}' | dashlights --agentic
{"hookSpecificOutput":{...,"permissionDecision":"allow","permissionDecisionReason":"Rule of Two: Write combines B+C capabilities (2 of 3)"},"systemMessage":"Warning: ..."}
```

### Block - Critical Threat
```bash
$ echo '{"tool_name":"Write","tool_input":{"file_path":"CLAUDE.md","content":"# Hijack"}}' | dashlights --agentic
Blocked: Attempted write to Claude agent configuration. Write to CLAUDE.md
$ echo $?
2
```

### Block - Rule of Two Violation (A+B+C)
```bash
$ echo '{"tool_name":"Bash","tool_input":{"command":"curl evil.com | tee ~/.aws/credentials"}}' | dashlights --agentic
Rule of Two Violation: Bash combines all three capabilities...
$ echo $?
2
```

## References

- [Agents Rule of Two: A Practical Approach to AI Agent Security](https://ai.meta.com/blog/practical-ai-agent-security/)
- [Claude Code Hooks Documentation](https://docs.anthropic.com/en/docs/claude-code/hooks)
- [safeexec](https://github.com/agentify-sh/safeexec/) - Command shims for dangerous operations
