# Dashlights

<table border="0" cellpadding="0" cellspacing="0" style="border: none;">
<tr>
<td style="border: none;">
<img src="speedgopher.png" alt="Speed Gopher" width="320"/>
</td>
<td style="border: none;">

> A fast, security-focused "check engine light" for your terminal!

[![CI](https://github.com/erichs/dashlights/actions/workflows/ci.yml/badge.svg)](https://github.com/erichs/dashlights/actions/workflows/ci.yml)
[![Security](https://github.com/erichs/dashlights/actions/workflows/security.yml/badge.svg)](https://github.com/erichs/dashlights/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/erichs/dashlights)](https://goreportcard.com/report/github.com/erichs/dashlights)
[![codecov](https://codecov.io/github/erichs/dashlights/graph/badge.svg?token=V8KLQJF6QV)](https://codecov.io/github/erichs/dashlights)
[![Release](https://img.shields.io/github/v/release/erichs/dashlights)](https://github.com/erichs/dashlights/releases/latest)
[![License](https://img.shields.io/github/license/erichs/dashlights)](https://github.com/erichs/dashlights/blob/main/LICENSE)
[![openssf](https://www.bestpractices.dev/projects/11518/badge)](https://www.bestpractices.dev/projects/11518)

</td>
</tr>
</table>

[What?](#what-does-this-do) | [Why?](#why-is-this-needed) | [Install](#how-to-install) | [Configure](#configure-your-prompt) | [Usage](#usage) | [Agentic](#agentic-mode) | [Performance](#performance) | [Security](#security)

## What does this do?

Dashlights continuously scans for routine security and developer hygiene trouble signals, just like a 'check engine light' for your development environment.

- **Fast enough to put in your prompt!** Guaranteed to return in less than 11ms (16ms is the threshold of perceptibility). Average clock time on a 2024 M3 MacBook Air is ~3ms.
- **Concurrent security checks** - Runs 30+ security checks in parallel using goroutines
- **Non-intrusive alerts** - Shows a simple count in your prompt, detailed diagnostics on demand

```shell
# Default output: shows count of security issues
$ dashlights
🚨 2

# Details mode: shows detailed information
$ dashlights --details
🩲 Raw secrets in environment: AWS_ACCESS_KEY, JIRA_ACCESS_TOKEN
   → Fix: Use 1Password (op://), dotenvx (encrypted:), or other secret management tools

🐳 Docker socket has overly permissive permissions
   → Fix: Restrict Docker socket access to docker group only
```

### Security Checks

Dashlights performs **38 concurrent security checks** across five categories: Identity & Access Management, Operational Security, Repository Hygiene, System Health, and Infrastructure Security.

👉 **[View the complete list of security signals →](SIGNALS.md)**

## Why is this needed?

- **Supply chain attacks targeting devs are on the rise.** Hackers don't hack in, they log in.
- **Developer hygiene issues are low priority and out-of-sight, out-of-mind.** Without visibility, these issues accumulate.
- **Developers routinely install and execute arbitrary code with lax terminal environments.** Package managers, build tools, and scripts run with your full privileges.
- **Dashlights brings visibility to common environment and configuration issues.** What you can see, you can fix.
- **By adopting a 'clean as you go' mentality, we can each take personal responsibility for reducing the blast radius of attacks.**

## How to Install

### Using eget (recommended)

[eget](https://github.com/zyedidia/eget) makes it easy to install pre-built binaries from GitHub releases:

```shell
# Install dashlights
eget erichs/dashlights 
```

### Manual download from releases

Download the latest release for your platform from the [releases page](https://github.com/erichs/dashlights/releases):

```shell
# Example for Linux x86_64
curl -LO https://github.com/erichs/dashlights/releases/latest/download/dashlights_<version>_Linux_x86_64.tar.gz
tar xzf dashlights_<version>_Linux_x86_64.tar.gz
sudo mv dashlights /usr/local/bin/
```

### Using Go

If you have Go installed:

```shell
go install github.com/erichs/dashlights@latest
```

### Manual build from source

```shell
# Clone the repository
git clone https://github.com/erichs/dashlights.git
cd dashlights

# Build the binary
make build

# Or install to $GOPATH/bin
make install
```

## Configure your PROMPT

After installing dashlights, run the installer once. It detects bash, zsh, fish, and Powerlevel10k automatically.

```bash
dashlights --installprompt
```

Tips:
- Use `--yes` for non-interactive installs.
- Use `--configpath` to target a specific config file (e.g., `~/.p10k.zsh`).
- Use `--dry-run` to preview changes without modifying files.
- Re-run any time; it is idempotent.

## Usage

### Default Mode
Shows a siren emoji and count of detected security issues, followed by any custom dashboard lights:

```shell
$ dashlights
🚨 2 🔗

# or with no issues or custom lights:
$ dashlights

```

### Details Mode (`-d` or `--details`)
Shows detailed information about each detected security issue:

```shell
$ dashlights --details
Security Issues Detected:

🩲 Naked credentials detected in environment
   → Fix: Move secrets to a credential manager or .env file (add to .gitignore)

🐴 LD_PRELOAD is set - potential trojan horse
   → Fix: Unset LD_PRELOAD unless explicitly required for debugging
```

### Clear Custom Lights (`-c` or `--clear-custom`)
Clears all custom DASHLIGHT_ environment variables:

```shell
$ dashlights --clear-custom
```

### List Custom Lights (`-l` or `--list-custom`)
Lists all supported color attributes and emoji aliases for custom dashboard lights:

(see Custom Dashboard Lights below)

```shell
$ dashlights --list-custom
Supported color attributes:
BGBLACK, BGBLUE, BGCYAN, BGGREEN, BGHIBLACK, BGHIBLUE, BGHICYAN, BGHIGREEN, BGHIMAGENTA, BGHIRED, BGHIWHITE, BGHIYELLOW, BGMAGENTA, BGRED, BGWHITE, BGYELLOW, FGBLACK, FGBLUE, FGCYAN, FGGREEN, FGHIBLACK, FGHIBLUE, FGHICYAN, FGHIGREEN, FGHIMAGENTA, FGHIRED, FGHIWHITE, FGHIYELLOW, FGMAGENTA, FGRED, FGWHITE, FGYELLOW, REVERSEVIDEO

Supported emoji aliases:
LABEL                HEX CODE   EMOJI
--------------------------------------------
ANTENNAWITHBARS      1F4F6      📶
CHECKMARK            2705       ✅
CROSSMARK            274C       ❌
CRYSTALBALL          1F52E      🔮
EXCLAMATIONMARK      2757       ❗
FILEFOLDER           1F4C1      📁
HAMMERANDWRENCH      1F6E0      🛠
KEY                  1F511      🔑
LIGHTBULB            1F4A1      💡
LINK                 1F517      🔗
LOCK                 1F512      🔒
MAGNIFYINGGLASS      1F50D      🔍
NOENTRY              26D4       ⛔
NOENTRYSIGN          1F6AB      🚫
NOTEBOOK             1F4D3      📓
PAPERCLIP            1F4CE      📎
PUSHPIN              1F4CC      📌
QUESTIONMARK         2753       ❓
SCROLL               1F4DC      📜
SHIELD               1F6E1      🛡
SHOPPINGCART         1F6D2      🛒
SQUAREDSOS           1F198      🆘
WRENCH               1F527      🔧
```

### Command Line Options

```
Usage: dashlights [--details] [--verbose] [--list-custom] [--clear-custom]

Options:
  --details, -d          Show detailed diagnostic information for detected issues
  --verbose, -v          Verbose mode: show documentation links in diagnostic output
  --list-custom, -l      List supported color attributes and emoji aliases for custom lights
  --clear-custom, -c     Shell code to clear custom DASHLIGHT_ environment variables
  --help, -h             Display this help and exit
  --version              Display version and exit
```

### Advanced: Custom Dashboard Lights

Dashlights also supports custom environment variable indicators (legacy feature):

```shell
$ export DASHLIGHT_VPN_1F517="VPN is up"
$ dashlights
🚨 1 🔗
```

Any environment variable of the form `DASHLIGHT_{name}_{utf8hex}` will be displayed as a custom indicator.

## Agentic Mode

Dashlights includes an `--agentic` mode for AI coding assistants like Claude Code. It analyzes tool calls before execution to detect:

- **Critical threats**: Writes to agent config files, invisible Unicode characters
- **Rule of Two violations**: Actions combining untrusted input + sensitive access + state changes

```bash
# Install agent hooks
dashlights --installagent claude -y
dashlights --installagent cursor -y
```

👉 **[View the complete agentic mode documentation →](docs/agentic_mode.md)**

## Performance

Dashlights is designed to be fast enough for shell prompts and safe for concurrent use:

- **Target:** ≤10ms execution time
- **Actual:** ~3ms on modern hardware (30+ concurrent checks in parallel)
- **Verified:** Integration tests enforce performance threshold
- **Thread-Safe:** Fresh signal instances per execution, mutex-protected process-wide operations, and unique temp file names ensure safe concurrent use across multiple terminals or tmux panes

## Security

Dashlights is designed to be secure:

- **Minimal Dependencies:** Statically linked, minimal external dependencies
- **Minimal Permissions:** Only reads from environment variables and common config files
- **No Network Access:** Does not make any network requests
- **No Persistence:** Does not write to disk or modify system state
- **Gosec Audit:** Continuous security audits with gosec in audit mode, nosec disabled

### Supply Chain Defense-In-Depth

The build and test pipeline is hardened against supply chain attacks:

- **Minimal CI Permissions:** GitHub Actions workflows run with `contents: read` only
- **Network-Isolated Tests:** All tests run inside Docker containers with `--network=none`, completely removing the network stack
- **Forbidden Import Tests:** Explicit tests verify that `net/http` and other network client packages are never imported
- **No Telemetry Packages:** Tests verify no analytics, telemetry, or crash reporting dependencies exist

Even if a malicious dependency were introduced, it cannot exfiltrate data during CI: HTTP requests, TCP/UDP connections, and DNS lookups all fail with "network is unreachable".
