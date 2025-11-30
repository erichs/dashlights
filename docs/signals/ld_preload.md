# LD_PRELOAD / DYLD_INSERT_LIBRARIES Detected

## What this is

This signal detects when `LD_PRELOAD` (Linux) or `DYLD_INSERT_LIBRARIES` (macOS) environment variables are set. These variables allow loading custom shared libraries before any other libraries, effectively hijacking function calls in any program you run.

While legitimate for debugging and testing, these variables are a common persistence and privilege escalation technique used by attackers.

## Why this matters

**Security Risk**:
- **Rootkit technique**: Attackers use LD_PRELOAD to hide processes, files, and network connections
- **Credential theft**: Can intercept authentication functions to steal passwords
- **Code injection**: Malicious libraries can execute arbitrary code in every process
- **Privilege escalation**: Can hijack setuid binaries to gain root access
- **Persistence**: Survives reboots if set in shell configuration files

**Attack examples**:
```bash
# Attacker sets LD_PRELOAD to intercept SSH passwords
export LD_PRELOAD=/tmp/evil.so
# Now every program loads evil.so first

# Intercept getpwnam() to hide users
# Intercept readdir() to hide files
# Intercept socket() to hide network connections
```

**Legitimate uses** (rare):
- Debugging memory leaks with valgrind
- Performance profiling
- Testing library replacements
- Working around library bugs

## How to remediate

### Unset the variable immediately

**Check if set**:
```bash
# Linux
echo $LD_PRELOAD

# macOS
echo $DYLD_INSERT_LIBRARIES
```

**Unset it**:
```bash
# Linux
unset LD_PRELOAD

# macOS
unset DYLD_INSERT_LIBRARIES

# Verify
env | grep -E 'LD_PRELOAD|DYLD_INSERT_LIBRARIES'
```

### Remove from shell configuration

**Find where it's set**:
```bash
# Search all shell config files
grep -r "LD_PRELOAD\|DYLD_INSERT_LIBRARIES" \
  ~/.bashrc ~/.bash_profile ~/.profile ~/.zshrc ~/.zprofile \
  /etc/profile /etc/bash.bashrc /etc/zsh/zshrc 2>/dev/null
```

**Remove the export statement**:
```bash
# Edit the file and remove lines like:
# export LD_PRELOAD=/path/to/library.so
# export DYLD_INSERT_LIBRARIES=/path/to/library.dylib

nano ~/.bashrc  # or whichever file contains it
```

**Reload shell configuration**:
```bash
source ~/.bashrc  # or ~/.zshrc
# Or start a new shell
exec $SHELL
```

### Investigate the loaded library

**If LD_PRELOAD was set, investigate the library**:
```bash
# Check what library was being loaded
echo $LD_PRELOAD

# Examine the library file
ls -la /path/to/library.so
file /path/to/library.so

# Check file ownership and permissions
stat /path/to/library.so

# See what symbols it exports
nm -D /path/to/library.so

# Check for suspicious strings
strings /path/to/library.so | less
```

**Common malicious indicators**:
- Library in `/tmp` or `/var/tmp`
- Owned by a different user
- Recently created/modified
- Contains strings like "password", "keylog", "backdoor"
- Exports functions like `getpwnam`, `readdir`, `socket`

### Check for persistence mechanisms

**System-wide LD_PRELOAD**:
```bash
# Check /etc/ld.so.preload (system-wide preload)
cat /etc/ld.so.preload

# This file should normally not exist or be empty
# If it contains libraries, investigate them
```

**Check systemd services** (Linux):
```bash
# Search for LD_PRELOAD in systemd units
sudo grep -r "LD_PRELOAD" /etc/systemd/system/
sudo grep -r "LD_PRELOAD" /usr/lib/systemd/system/
```

**Check launchd** (macOS):
```bash
# Search for DYLD_INSERT_LIBRARIES in launchd plists
sudo grep -r "DYLD_INSERT_LIBRARIES" /Library/LaunchDaemons/
sudo grep -r "DYLD_INSERT_LIBRARIES" /Library/LaunchAgents/
grep -r "DYLD_INSERT_LIBRARIES" ~/Library/LaunchAgents/
```

### Platform-specific cleanup

**Linux**:
```bash
# Unset LD_PRELOAD
unset LD_PRELOAD

# Check and clear /etc/ld.so.preload if suspicious
sudo cat /etc/ld.so.preload
# If it contains suspicious libraries:
sudo rm /etc/ld.so.preload

# Rebuild ld cache
sudo ldconfig
```

**macOS**:
```bash
# Unset DYLD_INSERT_LIBRARIES
unset DYLD_INSERT_LIBRARIES

# macOS System Integrity Protection (SIP) prevents DYLD_* 
# from affecting system binaries, but check anyway
csrutil status
# Should show: System Integrity Protection status: enabled
```

### Security best practices

1. **Never set LD_PRELOAD in production** unless absolutely necessary and well-documented

2. **Use alternatives for debugging**:
   ```bash
   # Instead of LD_PRELOAD, use:
   
   # For memory debugging
   valgrind --leak-check=full ./program
   
   # For system call tracing
   strace ./program
   
   # For library call tracing
   ltrace ./program
   ```

3. **Monitor for LD_PRELOAD usage**:
   ```bash
   # Add to your monitoring/alerting
   if [ -n "$LD_PRELOAD" ]; then
     echo "WARNING: LD_PRELOAD is set to $LD_PRELOAD" | mail -s "Security Alert" admin@example.com
   fi
   ```

4. **Use file integrity monitoring**:
   ```bash
   # Monitor /etc/ld.so.preload with AIDE, Tripwire, or similar
   sudo aide --check
   ```

5. **Restrict library loading** (advanced):
   ```bash
   # Use SELinux or AppArmor to restrict which libraries can be loaded
   ```

### If you suspect compromise

1. **Isolate the system** from the network
2. **Collect forensic evidence**:
   ```bash
   # Save the library file
   cp $LD_PRELOAD /tmp/evidence/
   
   # Save process list
   ps auxf > /tmp/evidence/processes.txt
   
   # Save network connections
   netstat -tulpn > /tmp/evidence/network.txt
   ```
3. **Check for other indicators of compromise**
4. **Rotate all credentials** used on the system
5. **Consider full system rebuild** if rootkit is confirmed

