# Root Login

## What this is

This signal detects when you are logged in as the root user (UID 0). Running as root gives you unrestricted access to the entire system, which is dangerous for day-to-day development work.

Root access means:
- No permission checks on file operations
- Ability to modify any system file or configuration
- Full access to all users' data
- Ability to install system-wide software and services

## Why this matters

**Security Risk**:
- **Increased attack surface**: Malicious code executed as root has full system control
- **Supply chain attacks**: Package managers and build tools run with your privileges - as root, compromised packages can do anything
- **Credential exposure**: Root access can read any file, including other users' credentials
- **No audit trail separation**: Actions taken as root are harder to attribute to specific tasks

**Operational Risk**:
- **Accidental damage**: Typos and mistakes have no safety net (e.g., `rm -rf /` instead of `rm -rf ./`)
- **No permission errors**: You won't notice when you're doing something that should require elevated privileges
- **Configuration drift**: System files can be accidentally modified without realizing it
- **Harder debugging**: Issues caused by running as root may not reproduce for normal users

**Examples of risk**:
```bash
# Accidental system destruction
rm -rf /tmp /important  # Missing space before /important

# Installing malicious packages with full access
npm install compromised-package  # Can modify /etc, /usr, anywhere

# Overwriting system files
echo "debug" > /etc/passwd  # Oops, meant /tmp/passwd

# Running unknown scripts with full privileges
curl https://example.com/install.sh | bash  # Now has root access
```

## How to remediate

### Switch to a non-root user

**Check current user**:
```bash
# See current user
whoami

# Check effective UID
id -u
# 0 means root

# Full user info
id
```

**Switch to normal user**:
```bash
# Switch to your normal user account
su - username

# Or exit root shell
exit
```

### Use sudo for privileged commands

**Run individual commands with sudo**:
```bash
# Instead of running as root
apt install package  # BAD: running as root

# Use sudo for specific commands
sudo apt install package  # GOOD: elevate just for this command

# Edit system files
sudo nano /etc/hosts

# Restart services
sudo systemctl restart nginx
```

**Configure sudo access**:
```bash
# Add user to sudo group (Debian/Ubuntu)
sudo usermod -aG sudo username

# Add user to wheel group (RHEL/CentOS)
sudo usermod -aG wheel username

# Verify sudo access
sudo -l
```

### Best practices

1. **Never log in as root for development**:
   ```bash
   # Create a non-root user if you don't have one
   sudo useradd -m -s /bin/bash devuser
   sudo passwd devuser
   sudo usermod -aG sudo devuser
   ```

2. **Use sudo with specific commands**:
   ```bash
   # GOOD: Specific elevated command
   sudo apt update

   # BAD: Opening a root shell
   sudo -i  # Avoid this
   sudo su  # Avoid this
   ```

3. **Set a sudo timeout**:
   ```bash
   # Edit sudoers to require password more frequently
   sudo visudo
   # Add: Defaults timestamp_timeout=5
   ```

4. **Audit sudo usage**:
   ```bash
   # View sudo logs
   sudo grep sudo /var/log/auth.log

   # Or on systems with journald
   sudo journalctl -e _COMM=sudo
   ```

5. **Use sudo for Docker instead of running as root**:
   ```bash
   # Add user to docker group
   sudo usermod -aG docker $USER

   # Log out and back in, then:
   docker ps  # Works without sudo
   ```

### If you must use root access

Sometimes root access is necessary. If so:

1. **Time-box your root session**: Switch back to normal user as soon as possible
2. **Document what you're doing**: Keep notes of root-level changes
3. **Use sudo instead**: Even if you could use root, prefer `sudo` for audit trail
4. **Double-check destructive commands**: Especially `rm`, `mv`, `chmod`
5. **Never run untrusted code as root**: Avoid `curl | bash` patterns


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_ROOT_LOGIN=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
