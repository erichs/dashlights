# Unprotected SSH Keys

## What this is

This signal detects SSH private keys in your `~/.ssh` directory that are not protected with a passphrase. Unprotected SSH keys are a significant security risk because anyone who gains access to the key file can use it to authenticate as you.

## Why this matters

**Security Risk**:
- **Credential theft**: If your laptop is stolen or compromised, attackers can use your SSH keys
- **Lateral movement**: Stolen keys allow attackers to access all servers where your key is authorized
- **No second factor**: Unprotected keys provide single-factor authentication
- **Persistent access**: Keys often remain valid for years, providing long-term access

**Attack scenarios**:
```bash
# Attacker gains access to your laptop
# Copies your SSH key
cp ~/.ssh/id_rsa /tmp/stolen_key

# Uses it to access your servers
ssh -i /tmp/stolen_key user@production-server

# Now has access to all your servers
```

**Compliance**: Many security frameworks require multi-factor authentication and key protection.

## How to remediate

### Add passphrase to existing keys

**Check if key has passphrase**:
```bash
# Try to view the key
ssh-keygen -y -f ~/.ssh/id_rsa

# If it prompts for passphrase: ✓ Protected
# If it shows the public key immediately: ✗ Unprotected
```

**Add passphrase to existing key**:
```bash
# Add passphrase to unprotected key
ssh-keygen -p -f ~/.ssh/id_rsa

# You'll be prompted:
# Enter old passphrase: (press Enter if none)
# Enter new passphrase: (type a strong passphrase)
# Enter same passphrase again: (confirm)
```

**Verify passphrase was added**:
```bash
# Try to view the key again
ssh-keygen -y -f ~/.ssh/id_rsa
# Should now prompt for passphrase
```

### Generate new keys with passphrase

**Generate new ed25519 key** (recommended):
```bash
# Generate new key with passphrase
ssh-keygen -t ed25519 -C "your-email@example.com"

# You'll be prompted:
# Enter file: (press Enter for default)
# Enter passphrase: (type a strong passphrase)
# Enter same passphrase again: (confirm)
```

**Generate RSA key** (if ed25519 not supported):
```bash
# Generate 4096-bit RSA key
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
```

**Copy public key to servers**:
```bash
# Copy new key to server
ssh-copy-id -i ~/.ssh/id_ed25519.pub user@server

# Or manually
cat ~/.ssh/id_ed25519.pub | ssh user@server 'cat >> ~/.ssh/authorized_keys'
```

**Remove old unprotected key** (after verifying new key works):
```bash
# Test new key first
ssh -i ~/.ssh/id_ed25519 user@server

# If it works, remove old key from servers
ssh user@server
nano ~/.ssh/authorized_keys
# Remove the old key

# Then delete local old key
rm ~/.ssh/id_rsa ~/.ssh/id_rsa.pub
```

### Use SSH agent to avoid typing passphrase

**Add key to SSH agent**:
```bash
# Start SSH agent (if not running)
eval "$(ssh-agent -s)"

# Add key to agent (will prompt for passphrase once)
ssh-add ~/.ssh/id_ed25519

# Verify key is loaded
ssh-add -l
```

**macOS: Use Keychain**:
```bash
# Add key to macOS Keychain (remembers passphrase)
ssh-add --apple-use-keychain ~/.ssh/id_ed25519

# Configure SSH to use Keychain
cat >> ~/.ssh/config <<EOF
Host *
    UseKeychain yes
    AddKeysToAgent yes
EOF
```

**Linux: Use keyring**:
```bash
# Install keyring (GNOME)
sudo apt install gnome-keyring

# Or use KDE Wallet
sudo apt install kwalletmanager

# SSH agent will integrate with keyring automatically
```

**Windows: Use Pageant or Windows SSH Agent**:
```powershell
# Start SSH agent
Start-Service ssh-agent

# Add key
ssh-add $env:USERPROFILE\.ssh\id_ed25519
```

### Use hardware security keys

**YubiKey** (most secure):
```bash
# Generate key on YubiKey (requires YubiKey 5.2.3+)
ssh-keygen -t ed25519-sk -C "your-email@example.com"

# Or FIDO2 resident key
ssh-keygen -t ed25519-sk -O resident -C "your-email@example.com"

# Key is stored on hardware, can't be stolen
```

**Benefits**:
- Private key never leaves the hardware device
- Requires physical presence to use
- Resistant to malware and theft

### Best practices for passphrases

**Strong passphrase characteristics**:
- At least 20 characters
- Mix of words, numbers, symbols
- Not based on personal information
- Unique (not reused)

**Good passphrase examples**:
```
correct-horse-battery-staple-2024!
MyDog@Loves2Swim&PlayFetch
Coffee+Laptop=Productivity2024
```

**Use a password manager**:
- Store SSH key passphrases in 1Password, Bitwarden, etc.
- Generate strong random passphrases
- Never forget passphrases

### Rotate keys regularly

**Key rotation schedule**:
```bash
# Every 6-12 months:
# 1. Generate new key with passphrase
ssh-keygen -t ed25519 -C "your-email@example.com"

# 2. Copy to all servers
for server in server1 server2 server3; do
    ssh-copy-id -i ~/.ssh/id_ed25519.pub user@$server
done

# 3. Test new key
ssh -i ~/.ssh/id_ed25519 user@server1

# 4. Remove old key from servers
# 5. Delete old local key
```

### Audit your SSH keys

**Find all SSH keys**:
```bash
# List all private keys
find ~/.ssh -type f -name "id_*" ! -name "*.pub"

# Check each for passphrase
for key in ~/.ssh/id_*; do
    if [ -f "$key" ] && [[ "$key" != *.pub ]]; then
        echo "Checking $key..."
        ssh-keygen -y -f "$key" > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo "  ✗ UNPROTECTED"
        else
            echo "  ✓ Protected"
        fi
    fi
done
```

**Check key permissions**:
```bash
# Private keys should be 600
ls -la ~/.ssh/id_*

# Fix permissions if needed
chmod 600 ~/.ssh/id_*
chmod 644 ~/.ssh/id_*.pub
```

### Platform-specific recommendations

**macOS**:
```bash
# Use Keychain for passphrase storage
ssh-add --apple-use-keychain ~/.ssh/id_ed25519

# Configure SSH
cat >> ~/.ssh/config <<EOF
Host *
    UseKeychain yes
    AddKeysToAgent yes
    IdentityFile ~/.ssh/id_ed25519
EOF
```

**Linux**:
```bash
# Use systemd user service for SSH agent
# (see ssh_agent_bloat.md for setup)

# Or add to shell startup
echo 'eval "$(ssh-agent -s)"' >> ~/.bashrc
echo 'ssh-add ~/.ssh/id_ed25519' >> ~/.bashrc
```

**Windows**:
```powershell
# Use Windows SSH Agent
Set-Service -Name ssh-agent -StartupType Automatic
Start-Service ssh-agent

# Add key
ssh-add $env:USERPROFILE\.ssh\id_ed25519
```

### If key is compromised

1. **Immediately remove from all servers**:
   ```bash
   ssh user@server
   nano ~/.ssh/authorized_keys
   # Delete the compromised key
   ```

2. **Generate new key** with passphrase

3. **Deploy new key** to all servers

4. **Rotate any other credentials** accessed from compromised servers

5. **Investigate** how the key was compromised

6. **Monitor** for unauthorized access

### Additional security measures

1. **Use certificate-based authentication** for large deployments
2. **Implement key expiration** policies
3. **Use bastion hosts** for production access
4. **Enable MFA** on servers when possible
5. **Monitor SSH access** logs
6. **Use short-lived credentials** when available
7. **Implement just-in-time access** for sensitive systems

### Key management tools

**For teams**:
- **Teleport**: Certificate-based SSH access
- **Boundary**: HashiCorp's access management
- **AWS Systems Manager Session Manager**: No SSH keys needed
- **Google Cloud IAP**: Identity-aware proxy

**For individuals**:
- **1Password SSH Agent**: Manage keys in 1Password
- **Secretive**: macOS app for Secure Enclave keys
- **ssh-vault**: Encrypt SSH keys at rest

