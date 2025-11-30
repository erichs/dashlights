# Privileged Path Entry

## What this is

This signal detects when your `PATH` environment variable includes directories that are writable by other users (world-writable or group-writable). This is a security risk because attackers can place malicious executables in these directories, and when you run a command, the malicious version executes instead of the legitimate one.

Common problematic paths:
- `/tmp` - World-writable temporary directory
- `/var/tmp` - World-writable temporary directory
- Current directory (`.`) - Can be manipulated by attackers
- Any directory with permissions 777, 775, or 755 owned by another user

## Why this matters

**Security Risk**:
- **Command hijacking**: Attackers place malicious binaries in writable PATH directories with names of common commands
- **Privilege escalation**: When you run a command as root (via sudo), the malicious version executes with elevated privileges
- **Credential theft**: Fake commands can steal passwords and tokens
- **Backdoor installation**: Malicious commands can install persistent backdoors

**Attack example**:
```bash
# Attacker creates malicious 'ls' in /tmp
cat > /tmp/ls <<'EOF'
#!/bin/bash
# Steal credentials
env | grep -i 'password\|secret\|token' > /tmp/stolen.txt
# Then run real ls
/bin/ls "$@"
EOF
chmod +x /tmp/ls

# If /tmp is in your PATH before /bin, you run the malicious version
ls  # Runs /tmp/ls instead of /bin/ls
```

## How to remediate

### Remove writable directories from PATH

**Check current PATH**:
```bash
echo $PATH | tr ':' '\n'
```

**Identify writable directories in PATH**:
```bash
# Check each PATH entry for write permissions
echo $PATH | tr ':' '\n' | while read dir; do
  if [ -d "$dir" ] && [ -w "$dir" ]; then
    echo "WRITABLE: $dir"
    ls -ld "$dir"
  fi
done
```

**Remove problematic directories**:
```bash
# Remove /tmp from PATH
export PATH=$(echo $PATH | tr ':' '\n' | grep -v '^/tmp$' | tr '\n' ':' | sed 's/:$//')

# Remove /var/tmp from PATH
export PATH=$(echo $PATH | tr ':' '\n' | grep -v '^/var/tmp$' | tr '\n' ':' | sed 's/:$//')

# Remove current directory (.)
export PATH=$(echo $PATH | tr ':' '\n' | grep -v '^\.$' | tr '\n' ':' | sed 's/:$//')

# Verify
echo $PATH
```

### Update shell configuration

**Find where PATH is set**:
```bash
# Search shell config files
grep -n "PATH=" ~/.bashrc ~/.bash_profile ~/.profile ~/.zshrc ~/.zprofile
```

**Edit shell configuration**:
```bash
# Edit your shell RC file
nano ~/.bashrc  # or ~/.zshrc

# Remove lines that add writable directories to PATH
# Look for lines like:
# export PATH="/tmp:$PATH"
# export PATH=".:$PATH"
# export PATH="/var/tmp:$PATH"

# Save and reload
source ~/.bashrc  # or ~/.zshrc
```

### Set a secure PATH

**Recommended PATH** (in order of precedence):
```bash
# In ~/.bashrc or ~/.zshrc
export PATH="/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin"

# Add user-specific bin directory (if it exists and is secure)
if [ -d "$HOME/bin" ]; then
  chmod 700 "$HOME/bin"  # Ensure it's not writable by others
  export PATH="$HOME/bin:$PATH"
fi

# Add user-local bin directory
if [ -d "$HOME/.local/bin" ]; then
  chmod 700 "$HOME/.local/bin"
  export PATH="$HOME/.local/bin:$PATH"
fi
```

**Verify PATH is secure**:
```bash
# Check for writable directories
echo $PATH | tr ':' '\n' | while read dir; do
  if [ -d "$dir" ]; then
    ls -ld "$dir"
  fi
done
```

### Platform-specific fixes

**macOS**:
```bash
# Default secure PATH
export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

# Add Homebrew (if installed)
if [ -d "/opt/homebrew/bin" ]; then
  export PATH="/opt/homebrew/bin:$PATH"
fi

# Add user bin
if [ -d "$HOME/bin" ]; then
  chmod 700 "$HOME/bin"
  export PATH="$HOME/bin:$PATH"
fi

# Add to ~/.zshrc (default shell on macOS)
nano ~/.zshrc
```

**Linux**:
```bash
# Default secure PATH
export PATH="/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin"

# Add user bin
if [ -d "$HOME/.local/bin" ]; then
  chmod 700 "$HOME/.local/bin"
  export PATH="$HOME/.local/bin:$PATH"
fi

# Add to ~/.bashrc
nano ~/.bashrc
```

**Windows (WSL)**:
```bash
# WSL inherits Windows PATH, which can be problematic
# Override in ~/.bashrc
export PATH="/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

# Add Windows interop paths if needed (carefully)
export PATH="$PATH:/mnt/c/Windows/System32"
```

### Secure user bin directories

**If you use ~/bin or ~/.local/bin**:
```bash
# Create with secure permissions
mkdir -p ~/bin
chmod 700 ~/bin

# Or
mkdir -p ~/.local/bin
chmod 700 ~/.local/bin

# Verify
ls -ld ~/bin ~/.local/bin
# Should show: drwx------ (700)
```

### Best practices

1. **Never add current directory (.) to PATH**:
   ```bash
   # Bad
   export PATH=".:$PATH"
   
   # Good - use explicit paths
   ./mycommand
   ```

2. **Never add world-writable directories** to PATH:
   ```bash
   # Bad
   export PATH="/tmp:$PATH"
   export PATH="/var/tmp:$PATH"
   ```

3. **Order PATH from most to least trusted**:
   ```bash
   # Most trusted first
   export PATH="$HOME/bin:/usr/local/bin:/usr/bin:/bin"
   ```

4. **Use absolute paths** for critical commands in scripts:
   ```bash
   #!/bin/bash
   # Instead of: rm -rf /tmp/data
   # Use: /bin/rm -rf /tmp/data
   ```

5. **Audit PATH regularly**:
   ```bash
   # Add to your security checklist
   echo $PATH | tr ':' '\n' | while read dir; do
     if [ -d "$dir" ] && [ -w "$dir" ]; then
       echo "WARNING: Writable directory in PATH: $dir"
     fi
   done
   ```

6. **Use `which` to verify command locations**:
   ```bash
   # Check which version of a command will run
   which ls
   # Should show: /bin/ls (not /tmp/ls)
   
   # Check all versions in PATH
   which -a ls
   ```

7. **Set secure PATH in sudo**:
   ```bash
   # Edit /etc/sudoers (use visudo)
   sudo visudo
   
   # Add or modify secure_path
   Defaults secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
   ```

### If you suspect compromise

1. **Check for suspicious executables** in PATH directories:
   ```bash
   # Find recently modified executables
   find $(echo $PATH | tr ':' ' ') -type f -executable -mtime -7 -ls 2>/dev/null
   ```

2. **Verify system commands** haven't been replaced:
   ```bash
   # Compare checksums with known good values
   sha256sum /bin/ls /bin/cat /bin/bash
   ```

3. **Restore from backups** if commands have been tampered with

4. **Investigate** how the attacker gained write access

