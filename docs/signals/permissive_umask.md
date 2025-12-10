# Permissive Umask

## What this is

This signal detects when your umask is set to an overly permissive value. The umask controls the default permissions for newly created files and directories. A permissive umask (like `0000` or `0002`) means new files are created with permissions that allow other users to read or write them.

- **umask 0000**: New files are world-writable (permissions 666/777)
- **umask 0002**: New files are group-writable (permissions 664/775)
- **umask 0022**: New files are readable by all but writable only by owner (permissions 644/755) - recommended
- **umask 0077**: New files are accessible only by owner (permissions 600/700) - most secure

## Why this matters

**Security Risk**:
- **Information disclosure**: Other users can read your files, including config files with secrets
- **Unauthorized modification**: With umask 0000, other users can modify your files
- **Privilege escalation**: Attackers with limited access can modify your scripts and configs
- **Credential theft**: Shell scripts, config files, and SSH keys become readable by others

**Examples of exposure**:
```bash
# With umask 0000:
echo "API_KEY=secret" > config.sh
ls -la config.sh
# -rw-rw-rw- (world-writable!)

# With umask 0022:
echo "API_KEY=secret" > config.sh
ls -la config.sh
# -rw-r--r-- (world-readable, but not writable)

# With umask 0077:
echo "API_KEY=secret" > config.sh
ls -la config.sh
# -rw------- (only owner can read/write)
```

**Multi-user systems**: On shared servers, development machines, or jump boxes, permissive umask allows other users to access your files.

## How to remediate

### Set a secure umask

**Check current umask**:
```bash
umask
# Shows: 0002 (bad) or 0022 (ok) or 0077 (best)
```

**Set umask for current session**:
```bash
# Recommended for most users
umask 0022

# Most secure (only owner can access files)
umask 0077

# Verify
umask
```

**Test the new umask**:
```bash
# Create a test file
touch test-file.txt
ls -la test-file.txt

# With umask 0022: -rw-r--r--
# With umask 0077: -rw-------

# Clean up
rm test-file.txt
```

### Make umask permanent

**For Bash** (add to ~/.bashrc):
```bash
# Edit ~/.bashrc
nano ~/.bashrc

# Add this line:
umask 0022  # or 0077 for maximum security

# Reload
source ~/.bashrc
```

**For Zsh** (add to ~/.zshrc):
```bash
# Edit ~/.zshrc
nano ~/.zshrc

# Add this line:
umask 0022  # or 0077 for maximum security

# Reload
source ~/.zshrc
```

**For all users** (system-wide):
```bash
# Edit /etc/profile (requires root)
sudo nano /etc/profile

# Add this line:
umask 0022

# Or for maximum security
umask 0077
```

### Fix existing files with wrong permissions

**Find world-writable files**:
```bash
# In home directory
find ~ -type f -perm -002 -ls

# In current directory
find . -type f -perm -002 -ls
```

**Fix permissions on existing files**:
```bash
# Fix all files in home directory
find ~ -type f -perm -002 -exec chmod o-w {} \;

# Fix specific directory
find ~/myproject -type f -perm -002 -exec chmod o-w {} \;

# Or use a safer default
find ~/myproject -type f -exec chmod 644 {} \;
find ~/myproject -type d -exec chmod 755 {} \;
```

**Fix sensitive files** (scripts, configs, keys):
```bash
# SSH keys
chmod 600 ~/.ssh/id_*
chmod 644 ~/.ssh/id_*.pub
chmod 700 ~/.ssh

# Shell scripts
find ~ -name "*.sh" -exec chmod 700 {} \;

# Config files
chmod 600 ~/.bashrc ~/.zshrc ~/.profile
chmod 600 ~/.aws/credentials
chmod 600 ~/.npmrc
```

### Platform-specific configuration

**macOS**:
```bash
# Set umask in ~/.zshrc (default shell)
echo "umask 0022" >> ~/.zshrc
source ~/.zshrc

# Or in ~/.bash_profile if using bash
echo "umask 0022" >> ~/.bash_profile
source ~/.bash_profile
```

**Linux**:
```bash
# Set umask in ~/.bashrc
echo "umask 0022" >> ~/.bashrc
source ~/.bashrc

# System-wide (requires root)
echo "umask 0022" | sudo tee -a /etc/profile
```

**Windows (WSL)**:
```bash
# Set umask in ~/.bashrc
echo "umask 0022" >> ~/.bashrc
source ~/.bashrc

# Note: Windows filesystem permissions work differently
# umask mainly affects files in WSL filesystem, not Windows drives
```

### Choosing the right umask

**umask 0022** (Recommended for most users):
- New files: 644 (rw-r--r--)
- New directories: 755 (rwxr-xr-x)
- **Pros**: Allows others to read files, good for shared projects
- **Cons**: Others can read your files
- **Use when**: Working on shared systems, collaborative projects

**umask 0027** (Balanced):
- New files: 640 (rw-r-----)
- New directories: 750 (rwxr-x---)
- **Pros**: Group can read, others cannot
- **Cons**: Requires proper group management
- **Use when**: Working in teams with group-based access control

**umask 0077** (Most secure):
- New files: 600 (rw-------)
- New directories: 700 (rwx------)
- **Pros**: Maximum privacy, only owner can access
- **Cons**: Others can't read your files (may break collaboration)
- **Use when**: Handling sensitive data, personal machines, security-critical work

### Best practices

1. **Set umask in shell config**, not in individual scripts:
   ```bash
   # In ~/.bashrc or ~/.zshrc
   umask 0022
   ```

2. **Use explicit permissions** for sensitive files:
   ```bash
   # Don't rely on umask alone
   touch secret.txt
   chmod 600 secret.txt
   ```

3. **Audit file permissions** regularly:
   ```bash
   # Find world-writable files
   find ~ -type f -perm -002 -ls

   # Find world-readable files with secrets
   find ~ -name "*secret*" -o -name "*password*" -o -name "*.key" | \
     xargs ls -la
   ```

4. **Use ACLs** for fine-grained control (Linux):
   ```bash
   # Set ACL for specific user
   setfacl -m u:username:r-- file.txt

   # View ACLs
   getfacl file.txt
   ```

5. **Document your umask choice** in team documentation

6. **Test umask** after changing it:
   ```bash
   # Create test files and check permissions
   touch test.txt
   mkdir testdir
   ls -la test.txt testdir
   rm test.txt
   rmdir testdir
   ```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_PERMISSIVE_UMASK=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
