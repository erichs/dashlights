# SSH Agent Bloat

## What this is

This signal detects when your SSH agent has more than 10 keys loaded. While SSH agents are useful for managing SSH keys, loading too many keys can cause authentication failures and security issues.

## Why this matters

**Authentication Failures**:
- **Too many authentication failures**: SSH servers limit authentication attempts (typically 6). If your agent tries all keys, you'll be locked out before the right key is tried
- **Connection refused**: Servers may temporarily ban your IP after too many failed attempts
- **Slow connections**: SSH tries each key sequentially, slowing down connections

**Security Concerns**:
- **Key exposure**: More keys in memory means more attack surface
- **Unclear key usage**: Hard to track which keys are used for what
- **Stale keys**: Old keys may still be loaded even after being revoked
- **Memory leaks**: Some SSH agent implementations leak memory with many keys

**Example error**:
```
Received disconnect from server: 2: Too many authentication failures
```

## How to remediate

### List loaded keys

**Check how many keys are loaded**:
```bash
# List all keys in SSH agent
ssh-add -l

# Count keys
ssh-add -l | wc -l
```

### Remove unnecessary keys

**Remove all keys**:
```bash
# Remove all keys from agent
ssh-add -D

# Verify
ssh-add -l
# Should show: The agent has no identities.
```

**Remove specific key**:
```bash
# Remove a specific key
ssh-add -d ~/.ssh/id_rsa

# Verify
ssh-add -l
```

### Add only needed keys

**Add specific keys**:
```bash
# Add only the keys you need
ssh-add ~/.ssh/id_ed25519
ssh-add ~/.ssh/work_key

# Verify
ssh-add -l
```

**Add keys with lifetime**:
```bash
# Add key that expires after 1 hour
ssh-add -t 3600 ~/.ssh/id_ed25519

# Add key that expires after 8 hours (work day)
ssh-add -t 28800 ~/.ssh/work_key
```

### Use SSH config for key selection

**Configure SSH to use specific keys per host**:
```bash
# Edit ~/.ssh/config
nano ~/.ssh/config
```

**Add host-specific key configuration**:
```
# Personal GitHub
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

# Work GitHub
Host github-work
    HostName github.com
    User git
    IdentityFile ~/.ssh/work_key
    IdentitiesOnly yes

# Work servers
Host *.company.com
    User youruser
    IdentityFile ~/.ssh/work_key
    IdentitiesOnly yes

# Personal servers
Host personal-server
    HostName server.example.com
    User youruser
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

**Important**: `IdentitiesOnly yes` prevents SSH from trying all keys in the agent.

### Organize your keys

**Use descriptive key names**:
```bash
# Instead of id_rsa, id_rsa_2, id_rsa_3
# Use descriptive names:
~/.ssh/github_personal
~/.ssh/github_work
~/.ssh/aws_production
~/.ssh/server_staging
```

**Add comments to keys**:
```bash
# When generating keys, add a comment
ssh-keygen -t ed25519 -C "github-personal" -f ~/.ssh/github_personal
ssh-keygen -t ed25519 -C "work-servers" -f ~/.ssh/work_key

# Comments show up in ssh-add -l
ssh-add -l
# 256 SHA256:... github-personal (ED25519)
# 256 SHA256:... work-servers (ED25519)
```

### Automate key management

**Create a script to load only needed keys**:
```bash
#!/bin/bash
# load-ssh-keys.sh

# Remove all keys
ssh-add -D

# Add only current project keys
ssh-add -t 28800 ~/.ssh/github_personal
ssh-add -t 28800 ~/.ssh/work_key

echo "Loaded SSH keys (expire in 8 hours):"
ssh-add -l
```

**Make it executable**:
```bash
chmod +x load-ssh-keys.sh
./load-ssh-keys.sh
```

### Platform-specific configuration

**macOS (using Keychain)**:
```bash
# Add key to macOS Keychain (persists across reboots)
ssh-add --apple-use-keychain ~/.ssh/id_ed25519

# Configure SSH to use Keychain
cat >> ~/.ssh/config <<EOF
Host *
    UseKeychain yes
    AddKeysToAgent yes
EOF
```

**Linux (systemd user service)**:
```bash
# Create systemd user service for SSH agent
mkdir -p ~/.config/systemd/user
cat > ~/.config/systemd/user/ssh-agent.service <<EOF
[Unit]
Description=SSH Agent

[Service]
Type=simple
Environment=SSH_AUTH_SOCK=%t/ssh-agent.socket
ExecStart=/usr/bin/ssh-agent -D -a $SSH_AUTH_SOCK

[Install]
WantedBy=default.target
EOF

# Enable and start
systemctl --user enable ssh-agent
systemctl --user start ssh-agent

# Add to shell config
echo 'export SSH_AUTH_SOCK="$XDG_RUNTIME_DIR/ssh-agent.socket"' >> ~/.bashrc
```

**Windows (Pageant)**:
```powershell
# Use Pageant (part of PuTTY)
# Or use Windows OpenSSH agent

# Start SSH agent service
Start-Service ssh-agent

# Set to start automatically
Set-Service -Name ssh-agent -StartupType Automatic

# Add key
ssh-add $env:USERPROFILE\.ssh\id_ed25519
```

### Best practices

1. **Use SSH config** instead of loading all keys:
   ```
   # ~/.ssh/config
   Host *
       IdentitiesOnly yes
   ```

2. **Load keys with expiration**:
   ```bash
   # Keys expire after 8 hours
   ssh-add -t 28800 ~/.ssh/id_ed25519
   ```

3. **Use one key per purpose**:
   - One for personal GitHub
   - One for work GitHub
   - One for production servers
   - One for staging servers

4. **Regularly audit loaded keys**:
   ```bash
   # Check weekly
   ssh-add -l
   
   # Remove unused keys
   ssh-add -D
   ```

5. **Use ed25519 keys** (smaller, faster, more secure):
   ```bash
   ssh-keygen -t ed25519 -C "description"
   ```

6. **Limit key lifetime** in SSH agent:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   alias ssh-add='ssh-add -t 28800'  # 8 hours
   ```

7. **Use separate agents** for different contexts:
   ```bash
   # Work agent
   SSH_AUTH_SOCK=~/.ssh/work-agent.sock ssh-agent

   # Personal agent
   SSH_AUTH_SOCK=~/.ssh/personal-agent.sock ssh-agent
   ```

### Troubleshooting

**If you get "Too many authentication failures"**:
```bash
# Option 1: Remove all keys and add only needed one
ssh-add -D
ssh-add ~/.ssh/specific_key
ssh user@server

# Option 2: Use -i flag to specify key
ssh -i ~/.ssh/specific_key user@server

# Option 3: Use SSH config with IdentitiesOnly
# (see SSH config section above)
```

**If keys keep getting added automatically**:
```bash
# Check for AddKeysToAgent in SSH config
grep -r "AddKeysToAgent" ~/.ssh/config

# Disable automatic key adding
Host *
    AddKeysToAgent no
```

**If agent has no keys after reboot**:
```bash
# macOS: Use Keychain
ssh-add --apple-use-keychain ~/.ssh/id_ed25519

# Linux: Use systemd service or add to shell startup
echo 'ssh-add ~/.ssh/id_ed25519' >> ~/.bashrc
```

### Security recommendations

1. **Use passphrase-protected keys**
2. **Rotate keys regularly** (every 6-12 months)
3. **Remove old keys** from servers when no longer needed
4. **Use hardware keys** (YubiKey) for critical access
5. **Monitor SSH agent** for unauthorized key additions
6. **Use certificate-based authentication** for large deployments

