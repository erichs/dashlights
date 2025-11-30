# Debug Mode Enabled

## What this is

This signal detects when debug, trace, or verbose environment variables (`DEBUG`, `TRACE`, `VERBOSE`) are set in your shell environment. These variables are commonly used by applications and frameworks to enable detailed logging and diagnostic output.

## Why this matters

**Security Risks**:
- **Information disclosure**: Debug logs often contain sensitive data like API keys, tokens, database queries with parameters, user data, and internal system paths
- **Performance degradation**: Verbose logging can significantly slow down applications and fill up disk space
- **Log injection attacks**: Detailed logs may be more vulnerable to log injection if they include unsanitized user input

**Production Impact**:
- **Log spam**: Debug mode can generate gigabytes of logs, filling up disk space and making it hard to find actual errors
- **Compliance violations**: Logging sensitive data (PII, PHI, PCI data) in debug logs can violate regulatory requirements
- **Incident response**: Excessive logging can obscure security events and make forensic analysis harder

**Examples of leaked data**:
- Database connection strings with passwords
- API request/response bodies containing tokens
- User session data and cookies
- Internal IP addresses and network topology
- Stack traces revealing code structure

## How to remediate

### Unset debug variables

**For current shell session**:
```bash
unset DEBUG
unset TRACE
unset VERBOSE
```

**Verify they're unset**:
```bash
env | grep -E '^(DEBUG|TRACE|VERBOSE)='
# Should return nothing
```

### Remove from shell configuration

**Check your shell RC files**:
```bash
# For bash
grep -n 'DEBUG\|TRACE\|VERBOSE' ~/.bashrc ~/.bash_profile ~/.profile

# For zsh
grep -n 'DEBUG\|TRACE\|VERBOSE' ~/.zshrc ~/.zprofile

# For fish
grep -n 'DEBUG\|TRACE\|VERBOSE' ~/.config/fish/config.fish
```

**Remove the export statements**:
```bash
# Edit the file and remove lines like:
# export DEBUG=1
# export TRACE=true
# export VERBOSE=1

nano ~/.bashrc  # or ~/.zshrc
```

**Reload your shell**:
```bash
# For bash
source ~/.bashrc

# For zsh
source ~/.zshrc

# Or just start a new shell
exec $SHELL
```

### Platform-specific cleanup

**macOS**:
```bash
# Check launchd environment
launchctl getenv DEBUG
launchctl getenv TRACE
launchctl getenv VERBOSE

# Unset if present
launchctl unsetenv DEBUG
launchctl unsetenv TRACE
launchctl unsetenv VERBOSE
```

**Linux (systemd)**:
```bash
# Check systemd user environment
systemctl --user show-environment | grep -E 'DEBUG|TRACE|VERBOSE'

# Unset if present
systemctl --user unset-environment DEBUG TRACE VERBOSE
```

**Windows (PowerShell)**:
```powershell
# Check environment variables
Get-ChildItem Env: | Where-Object { $_.Name -match 'DEBUG|TRACE|VERBOSE' }

# Remove them
Remove-Item Env:DEBUG -ErrorAction SilentlyContinue
Remove-Item Env:TRACE -ErrorAction SilentlyContinue
Remove-Item Env:VERBOSE -ErrorAction SilentlyContinue
```

### Best practices

1. **Use application-specific debug flags** instead of global environment variables:
   ```bash
   # Instead of: export DEBUG=1
   # Use: myapp --debug
   ```

2. **Use log levels** in your applications:
   ```bash
   # Set log level explicitly
   export LOG_LEVEL=info  # or warn, error
   ```

3. **Use separate environments** for development and production:
   ```bash
   # In development
   export NODE_ENV=development
   
   # In production (never set DEBUG here)
   export NODE_ENV=production
   ```

4. **Check before deploying**:
   ```bash
   # Add to your deployment checklist
   env | grep -E '^(DEBUG|TRACE|VERBOSE)=' && echo "WARNING: Debug mode enabled!"
   ```

5. **Use structured logging** that can be configured per environment without environment variables

