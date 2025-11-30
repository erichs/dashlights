# Dashlights
[![Go Report Card](https://goreportcard.com/badge/github.com/erichs/dashlights)](https://goreportcard.com/report/github.com/erichs/dashlights)
[![codecov](https://codecov.io/github/erichs/dashlights/graph/badge.svg?token=V8KLQJF6QV)](https://codecov.io/github/erichs/dashlights)
[![openssf](https://www.bestpractices.dev/projects/11518/badge)](https://www.bestpractices.dev/projects/11518)

> A fast, security-focused "check engine light" for your prompt!

## What does this do?

Dashlights continuously scans for routine security and developer hygiene trouble signals, just like a 'check engine light' for your development environment.

- **Fast enough to put in your prompt!** Guaranteed to return in less than 10ms (16ms is the threshold of perceptibility). Average clock time on a 2024 M3 MacBook Air is ~3ms.
- **Concurrent security checks** - Runs 30+ security checks in parallel using goroutines
- **Non-intrusive alerts** - Shows a simple count in your prompt, detailed diagnostics on demand

```shell
# Default output: shows count of security issues
$ dashlights
🚨 2

# Diagnostic mode: shows detailed information
$ dashlights -d
🩲 Raw secrets in environment: AWS_ACCESS_KEY, JIRA_ACCESS_TOKEN
   → Fix: Use 1Password (op://), dotenvx (encrypted:), or other secret management tools

🐳 Docker socket has overly permissive permissions
   → Fix: Restrict Docker socket access to docker group only
```

### Security Checks

Dashlights performs over 30 concurrent security checks:

#### Identity & Access Management (IAM)
1. **Naked Credential** 🩲 - Finds raw secrets in environment variables
2. **Privileged Path** 💣 - Detects current directory (`.`) in PATH
3. **AWS CLI Alias Hijacking** 🪝 - Detects malicious AWS CLI aliases that override core commands

#### Operational Security (OpSec)
4. **Trojan Horse** 🐴 - Checks for LD_PRELOAD/DYLD_INSERT_LIBRARIES (rootkit vector)
5. **Blind Spot** 🕶️ - Detects disabled shell history
6. **Prod Panic** 🚨 - Alerts when kubectl/AWS context points to production
7. **Man in the Middle** 🕵️ - Alerts on active proxy settings
8. **Loose Cannon** 😷 - Checks for permissive umask (0000 or 0002)
9. **Exposed Socket** 🐳 - Checks Docker socket permissions and orphaned DOCKER_HOST
10. **Debug Mode Enabled** 🐛 - Detects debug/trace/verbose environment variables
11. **History Permissions** 🔐 - Checks shell history files for world-readable permissions
12. **SSH Agent Key Bloat** 🔑 - Detects too many keys in SSH agent (causes MaxAuthTries lockouts)
13. **Open Door** 🔑 - Detects SSH private keys with incorrect permissions

#### Repository Hygiene
14. **Unignored Secret** 📝 - Checks if .env files exist but aren't in .gitignore
15. **Root-Owned Home Files** 👑 - Finds files in $HOME owned by root
16. **World-Writable Configs** 🌍 - Detects config files with dangerous permissions
17. **Dead Letter** 🗝️ - Finds cryptographic keys not in .gitignore
18. **Go Replace Directive** 🔄 - Detects replace directives in go.mod (breaks builds)
19. **PyCache Pollution** 🐍 - Checks for __pycache__ directories not properly ignored
20. **NPM RC Tokens** 📦 - Detects auth tokens in project .npmrc (should be in ~/.npmrc)
21. **Cargo Path Dependencies** 🦀 - Checks for path dependencies in Cargo.toml
22. **Missing __init__.py** 📁 - Detects Python packages missing __init__.py files
23. **Snapshot Dependency** ☕ - Checks for SNAPSHOT dependencies on release branches (Java/Maven)

#### System Health
24. **Full Tank** 💾 - Alerts when disk usage exceeds 90%
25. **Reboot Pending** ♻️ - Detects pending system reboot (Linux)
26. **Zombie Processes** 🧟 - Alerts on excessive zombie processes
27. **Dangling Symlinks** 💔 - Detects symlinks pointing to non-existent targets
28. **Time Drift Detected** ⏰ - Detects drift between system time and filesystem time

#### Infrastructure Security (InfraSec)
29. **Local Terraform State** 🏗️ - Checks for local terraform.tfstate files (should use remote state)
30. **Root Kube Context** ☸️ - Alerts when Kubernetes context uses kube-system namespace
31. **Dangerous TF_VAR** 🔐 - Checks for dangerous Terraform variables in environment (secrets in shell history)

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
# Install eget first (if you don't have it)
curl https://zyedidia.github.io/eget.sh | sh

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

After installing dashlights, add it to your shell prompt to get continuous security monitoring.

### Bash

Add to your `~/.bashrc`:

```bash
# Add dashlights to your prompt
PS1='$(dashlights) '"$PS1"
```

### Zsh

Add to your `~/.zshrc`:

```bash
# For left prompt (PROMPT)
PROMPT='$(dashlights) '"$PROMPT"

# Or for right prompt (RPROMPT)
RPROMPT='$(dashlights)'
```

### oh-my-zsh

Add to your `~/.zshrc` after the oh-my-zsh initialization:

```bash
# Source oh-my-zsh first
source $ZSH/oh-my-zsh.sh

# Then add dashlights to your prompt
PROMPT='$(dashlights) '"$PROMPT"
```

### Powerlevel10k

If you use Powerlevel10k, you can add dashlights as a custom segment. Add to your `~/.zshrc`:

```bash
# Add before sourcing powerlevel10k
POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(dashlights_custom $POWERLEVEL9K_LEFT_PROMPT_ELEMENTS)
POWERLEVEL9K_CUSTOM_DASHLIGHTS="dashlights"
POWERLEVEL9K_CUSTOM_DASHLIGHTS_BACKGROUND="none"
POWERLEVEL9K_CUSTOM_DASHLIGHTS_FOREGROUND="red"
```

Or add to your right prompt:

```bash
POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=($POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS dashlights_custom)
```

### Fish

Add to your `~/.config/fish/config.fish`:

```fish
# Add dashlights to your prompt
function fish_prompt
    echo -n (dashlights)" "
    # ... rest of your prompt configuration
end
```

## Usage

### Default Mode
Shows a siren emoji and count of detected security issues, followed by any custom dashboard lights:

```shell
$ dashlights
🚨 2 🔗

# or with no issues or custom lights:
$ dashlights

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

(see Custom Dashboard Lights below)

```shell
$ dashlights -l
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
Usage: dashlights [--obd] [--list] [--clear]

Options:
  --obd, -d              On-Board Diagnostics: display detailed security diagnostics
  --list, -l             List custom dashboard lights
  --clear, -c            Shell code to clear set dashlights
  --help, -h             Display this help and exit
```

### Advanced: Custom Dashboard Lights

Dashlights also supports custom environment variable indicators (legacy feature):

```shell
$ export DASHLIGHT_VPN_1F517="VPN is up"
$ dashlights
🚨 0 🔗
```

Any environment variable of the form `DASHLIGHT_{name}_{utf8hex}` will be displayed as a custom indicator.

## Performance

Dashlights is designed to be fast enough for shell prompts:
- **Target:** ≤10ms execution time
- **Actual:** ~3ms on modern hardware (30+ concurrent checks)
- **Verified:** Integration test enforces performance threshold

## Concurrency & Thread-Safety

Dashlights is designed to be safe when multiple instances run concurrently (e.g., multiple terminal prompts rendering simultaneously):

- **Fresh State:** Each execution creates fresh signal instances, preventing shared mutable state
- **Process-Wide Operations:** Operations that modify process-wide state (e.g., umask checks) are serialized with mutexes
- **Unique Resources:** Temporary files use unique names to prevent collisions between concurrent instances
- **Tested:** Comprehensive concurrency tests verify thread-safety under high contention

This design ensures that running dashlights in multiple terminal windows or tmux panes simultaneously will not cause race conditions or incorrect results.

## Security

Dashlights is designed to be secure:

- **Minimal Dependencies:** Statically linked, minimal external dependencies
- **Minimal Permissions:** Only reads from environment variables and common config files
- **No Network Access:** Does not make any network requests
- **No Persistence:** Does not write to disk or modify system state
- **Gosec Audit:** Continuous security audits with gosec in audit mode, nosec disabled
