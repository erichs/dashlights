# Shell History Disabled

## What this is

This signal detects when shell history is intentionally disabled through environment variables. This includes:
- `HISTFILE` set to `/dev/null` (discards all history)
- `HISTCONTROL` set to `ignorespace` or `ignoreboth` (ignores commands starting with a space)

While some users disable history for privacy, it creates security blind spots that hinder incident response and forensic analysis.

## Why this matters

**Security & Incident Response**:
- **No audit trail**: When investigating security incidents, shell history is crucial for understanding what commands were executed
- **Forensic analysis impossible**: Can't determine what an attacker did or what data they accessed
- **Compliance violations**: Many security frameworks require command auditing
- **Insider threat detection**: Can't detect malicious insider activity without command history

**Operational Impact**:
- **Troubleshooting harder**: Can't review what commands were run before a problem occurred
- **Knowledge loss**: Can't reference previous commands for complex operations
- **Training issues**: New team members can't learn from command history

**Why attackers disable it**:
- **Cover tracks**: Attackers disable history to hide malicious commands
- **Avoid detection**: Prevents security tools from analyzing command patterns
- **Persistence**: Some malware disables history as part of installation

## How to remediate

### Re-enable shell history

**For Bash**:
```bash
# Remove HISTFILE=/dev/null if set
unset HISTFILE

# Remove or modify HISTCONTROL
unset HISTCONTROL
# Or set to a safer value
export HISTCONTROL=ignoredups  # Only ignore duplicate commands

# Set history size
export HISTSIZE=10000
export HISTFILESIZE=20000

# Enable timestamp in history
export HISTTIMEFORMAT="%F %T "

# Append to history file instead of overwriting
shopt -s histappend
```

**For Zsh**:
```bash
# Remove HISTFILE=/dev/null if set
unset HISTFILE
# Or set to default
export HISTFILE=~/.zsh_history

# Set history size
export HISTSIZE=10000
export SAVEHIST=10000

# Enable timestamp
setopt EXTENDED_HISTORY

# Append to history
setopt APPEND_HISTORY
setopt INC_APPEND_HISTORY
```

### Update shell configuration files

**Check your RC files**:
```bash
# For Bash
grep -n "HISTFILE\|HISTCONTROL" ~/.bashrc ~/.bash_profile ~/.profile

# For Zsh
grep -n "HISTFILE\|HISTCONTROL" ~/.zshrc ~/.zprofile
```

**Remove or modify problematic settings**:
```bash
# Edit your shell RC file
nano ~/.bashrc  # or ~/.zshrc

# Remove these lines:
# export HISTFILE=/dev/null
# export HISTCONTROL=ignorespace
# export HISTCONTROL=ignoreboth

# Add these instead:
export HISTSIZE=10000
export HISTFILESIZE=20000
export HISTTIMEFORMAT="%F %T "
export HISTCONTROL=ignoredups
```

**Reload configuration**:
```bash
# For Bash
source ~/.bashrc

# For Zsh
source ~/.zshrc

# Or start a new shell
exec $SHELL
```

### Verify history is working

**Check settings**:
```bash
# Verify HISTFILE is not /dev/null
echo $HISTFILE

# Verify HISTCONTROL
echo $HISTCONTROL

# Check history size
echo $HISTSIZE
```

**Test history**:
```bash
# Run a test command
echo "test command"

# Check it appears in history
history | tail -5

# Verify history file exists and is being written
ls -lh ~/.bash_history  # or ~/.zsh_history
tail ~/.bash_history
```

### Secure history files

**Set proper permissions**:
```bash
# History files should be readable only by you
chmod 600 ~/.bash_history
chmod 600 ~/.zsh_history

# Verify
ls -la ~/.bash_history ~/.zsh_history
```

**Prevent accidental deletion**:
```bash
# Make history file append-only (Linux)
sudo chattr +a ~/.bash_history

# Remove append-only attribute if needed
sudo chattr -a ~/.bash_history
```

### Alternative: Selective history control

If you need to occasionally run sensitive commands without logging them:

**Use a space prefix** (if HISTCONTROL=ignorespace):
```bash
# This command won't be logged (note the leading space)
 export SECRET_KEY="sensitive-value"
```

**Or temporarily disable history**:
```bash
# Disable for current session
set +o history

# Run sensitive commands
export SECRET_KEY="sensitive-value"

# Re-enable history
set -o history
```

**Or use a separate shell**:
```bash
# Start a new shell with history disabled
HISTFILE=/dev/null bash

# Run sensitive commands
# ...

# Exit when done
exit
```

### Best practices

1. **Always keep history enabled** for security and operational reasons
2. **Use proper secret management** instead of disabling history
3. **Set appropriate history size** (10000+ commands)
4. **Enable timestamps** for better forensics
5. **Protect history files** with proper permissions (600)
6. **Back up history files** periodically
7. **Use centralized logging** (syslog, auditd) for critical systems
8. **Educate team members** on why history is important

### For compliance environments

If you need command auditing for compliance:

```bash
# Enable bash auditing with auditd (Linux)
sudo auditctl -a always,exit -F arch=b64 -S execve -k commands

# Or use a centralized logging solution
# - Splunk
# - ELK Stack
# - Datadog
```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_HISTORY_DISABLED=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
