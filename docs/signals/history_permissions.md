# Shell History World-Readable

## What this is

This signal detects when shell history files (`~/.bash_history`, `~/.zsh_history`) have permissions that allow other users on the system to read them. Shell history often contains sensitive information like passwords, API keys, and confidential commands that should only be accessible to the file owner.

## Why this matters

**Security Risk**:
- **Credential exposure**: Shell history often contains commands with passwords, API keys, and tokens typed directly on the command line
- **Lateral movement**: Attackers who gain limited access can read history files to find credentials for privilege escalation
- **Information disclosure**: History reveals system architecture, internal hostnames, database names, and operational procedures
- **Compliance violations**: Exposing command history may violate data protection regulations

**Examples of sensitive data in history**:
```bash
mysql -u root -p'MySecretPassword123'
export AWS_SECRET_ACCESS_KEY=abc123...
curl -H "Authorization: Bearer secret-token" https://api.example.com
ssh user@internal-server.company.local
```

**Multi-user systems**: On shared systems (servers, jump boxes, development machines), world-readable history files allow any user to see what others have typed.

## How to remediate

### Fix permissions immediately

**Check current permissions**:
```bash
ls -la ~/.bash_history ~/.zsh_history
# Should show: -rw------- (600)
# Bad: -rw-r--r-- (644) or -rw-rw-r-- (664)
```

**Fix permissions**:
```bash
# Set to owner read/write only
chmod 600 ~/.bash_history
chmod 600 ~/.zsh_history

# Verify
ls -la ~/.bash_history ~/.zsh_history
# Should show: -rw------- 1 youruser youruser ...
```

### Fix for all history files

**Find and fix all history files**:
```bash
# Find all history files in your home directory
find ~ -maxdepth 1 -name ".*history" -type f

# Fix permissions on all of them
find ~ -maxdepth 1 -name ".*history" -type f -exec chmod 600 {} \;

# Verify
find ~ -maxdepth 1 -name ".*history" -type f -exec ls -la {} \;
```

### Prevent future permission issues

**Set umask to prevent world-readable files**:
```bash
# Add to ~/.bashrc or ~/.zshrc
umask 077  # New files will be created with 600 permissions

# Or use a less restrictive but still safe umask
umask 027  # New files will be 640 (owner rw, group r, others none)
```

**Verify umask**:
```bash
umask
# Should show: 0077 or 0027
```

### Audit and clean sensitive data from history

**Search for potential secrets**:
```bash
# Search for common secret patterns
grep -i "password\|secret\|token\|key" ~/.bash_history

# Search for export statements (often contain secrets)
grep "^export.*=" ~/.bash_history

# Search for curl with Authorization headers
grep -i "authorization" ~/.bash_history
```

**Remove sensitive entries**:
```bash
# Edit history file and remove sensitive lines
nano ~/.bash_history

# Or use sed to remove specific patterns
sed -i '/password/d' ~/.bash_history
sed -i '/AWS_SECRET/d' ~/.bash_history

# Clear entire history if needed (nuclear option)
cat /dev/null > ~/.bash_history
history -c
```

**Rotate exposed secrets**:
If you found secrets in your history file and it was world-readable:
1. **Assume the secrets are compromised**
2. **Rotate all exposed credentials immediately**
3. **Monitor for unauthorized access**

### Platform-specific fixes

**macOS**:
```bash
# Fix permissions
chmod 600 ~/.bash_history ~/.zsh_history

# Set umask in shell config
echo "umask 077" >> ~/.zshrc
source ~/.zshrc
```

**Linux**:
```bash
# Fix permissions
chmod 600 ~/.bash_history ~/.zsh_history

# Set umask in shell config
echo "umask 077" >> ~/.bashrc
source ~/.bashrc

# For system-wide setting
echo "umask 077" | sudo tee -a /etc/profile
```

**Multi-user servers**:
```bash
# Audit all users' history files (requires root)
sudo find /home -name ".*history" -type f ! -perm 600 -ls

# Fix all users' history files (requires root)
sudo find /home -name ".*history" -type f ! -perm 600 -exec chmod 600 {} \;
```

### Best practices

1. **Never type secrets on the command line**:
   ```bash
   # Bad
   mysql -p'MyPassword123'
   
   # Good - prompt for password
   mysql -p
   # (then type password when prompted)
   ```

2. **Use secret management tools**:
   ```bash
   # Use 1Password CLI
   op read "op://vault/item/password" | mysql -p
   
   # Use environment files
   source .env  # (with proper permissions)
   mysql -p"$DB_PASSWORD"
   ```

3. **Use history control** to avoid logging sensitive commands:
   ```bash
   # Add to ~/.bashrc
   export HISTCONTROL=ignorespace
   
   # Then prefix sensitive commands with a space
    mysql -p'secret'  # (note the leading space)
   ```

4. **Set history file permissions in shell config**:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   chmod 600 ~/.bash_history 2>/dev/null
   chmod 600 ~/.zsh_history 2>/dev/null
   ```

5. **Regular audits**:
   ```bash
   # Add to cron for weekly checks
   0 0 * * 0 find ~ -name ".*history" ! -perm 600 -exec chmod 600 {} \;
   ```

6. **Use separate accounts** on shared systems instead of sharing user accounts

