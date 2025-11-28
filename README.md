## Dashlights
[![Go Report
Card](https://goreportcard.com/badge/github.com/erichs/dashlights)](https://goreportcard.com/report/github.com/erichs/dashlights)

> A fast security hygiene checker that signals impending security issues

Dashlights is a lightning-fast security hygiene checker designed to run in your shell prompt. It performs concurrent security checks and alerts you to potential issues before they become problems.

**Performance:** Completes in ≤10ms (typically ~2ms) - fast enough for shell prompts

## Quick Start

```shell
# Default output: shows count of security issues
$ dashlights
🚨 0

# Diagnostic mode: shows detailed information
$ dashlights -d
✅ No security issues detected
```

## What It Checks

Dashlights performs 18 concurrent security checks (2 disabled for performance):

### Identity & Access Management (IAM)
1. **SSH Keys** 🔑 - Detects SSH private keys with incorrect permissions (should be 0600)
2. **SSH Agent Forwarding** 👻 - Detects forwarded SSH agents (security risk on untrusted servers)
3. **Naked Credentials** 🩲 - Finds raw secrets in environment variables
4. **Privileged Path** 💣 - Detects current directory (`.`) in PATH

### Operational Security (OpSec)
5. **Trojan Horse** 🐴 - Checks for LD_PRELOAD/DYLD_INSERT_LIBRARIES (rootkit vector)
6. **Blind Spot** 🕶️ - Detects disabled shell history
7. **Production Panic** 🚨 - Alerts when kubectl/AWS context points to production
8. **Man in the Middle** 🕵️ - Alerts on active proxy settings
9. **Loose Cannon** 😷 - Checks for permissive umask (0000 or 0002)

### Repository Hygiene
10. **.env Not Ignored** 🔓 - Checks if .env files are tracked in git
11. **Git Email Mismatch** 🎭 - Detects personal email in work repos (or vice versa)
12. **Root-Owned Home Files** 👑 - Finds files in $HOME owned by root
13. **World-Writable Configs** 🌍 - Detects config files with dangerous permissions
14. **Untracked Crypto Keys** 🗝️ - Finds private keys not in .gitignore

### System Health
15. **Disk Space** 💾 - Alerts when disk usage exceeds 90%
16. **Reboot Pending** 🔄 - Detects pending system reboot (Linux)
17. **Zombie Processes** 🧟 - Alerts on excessive zombie processes
18. **Docker Socket** 🐳 - Checks Docker socket permissions

*Disabled for performance: Sudo Cached (12ms), Time Drift (21ms)*

## Advanced: Custom Dashboard Lights

Dashlights also supports custom environment variable indicators (legacy feature):

```shell
$ export DASHLIGHT_VPN_1F517="VPN is up"
$ dashlights
🚨 0 🔗
```

Any environment variable of the form `DASHLIGHT_{name}_{utf8hex}` will be displayed as a custom indicator.

## Installation

```shell
go install github.com/erichs/dashlights@latest
```

## Usage

### Default Mode
Shows a siren emoji and count of detected security issues, followed by any custom dashboard lights:

```shell
$ dashlights
🚨 2 🔗
```

### Diagnostic Mode (`-d` or `--obd`)
Shows detailed information about each detected security issue:

```shell
$ dashlights -d
Security Issues Detected:

🩲 Naked credentials detected in environment
   → Fix: Move secrets to a credential manager or .env file (add to .gitignore)

🐴 LD_PRELOAD is set - potential trojan horse
   → Fix: Unset LD_PRELOAD unless explicitly required for debugging
```

### Clear Mode (`-c`)
Clears all custom DASHLIGHT_ environment variables:

```shell
$ dashlights -c
```

### List Mode (`-l`)
Lists all custom dashboard lights:

```shell
$ dashlights -l
🔗 VPN is up
```

## Command Line Options

```
Usage: dashlights [--obd] [--list] [--clear]

Options:
  --obd, -d              On-Board Diagnostics: display detailed security diagnostics
  --list, -l             List custom dashboard lights
  --clear, -c            Shell code to clear set dashlights
  --help, -h             Display this help and exit
```

## Performance

Dashlights is designed to be fast enough for shell prompts:
- **Target:** ≤10ms execution time
- **Actual:** ~7ms on modern hardware (18 concurrent checks)
- **Verified:** Integration test enforces performance threshold

The tool uses concurrent goroutines to run all 18 security checks in parallel. Each check was benchmarked individually:
- **Fast checks (<5ms):** Environment variable scans, single syscalls, file reads
- **Borderline checks (5-9ms):** File stat operations, git commands, directory scans
- **Disabled checks (≥10ms):** External command execution (sudo), network calls (NTP)
