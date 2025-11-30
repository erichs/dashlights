# Root-Owned Home Directory

## What this is

This signal detects when files or directories in your home directory are owned by root instead of your user account. This typically happens when you run commands with `sudo` that create or modify files in your home directory, leaving them owned by root.

## Why this matters

**Permission Issues**:
- **Access denied**: You can't modify or delete root-owned files in your own home directory
- **Application failures**: Programs running as your user can't write to root-owned config files
- **Build failures**: Development tools can't create artifacts in root-owned directories
- **Git issues**: Can't commit changes if repository files are owned by root

**Security Concerns**:
- **Privilege escalation**: Root-owned files in your home can be exploited for privilege escalation
- **Unexpected behavior**: Applications may behave differently when config files are owned by root
- **Audit trail**: Unclear who actually modified the files

**Common causes**:
```bash
# Running editor with sudo
sudo vim ~/.bashrc  # Now .bashrc is owned by root!

# Creating files with sudo
sudo touch ~/file.txt  # file.txt is owned by root

# Running commands in home directory as root
sudo npm install  # node_modules owned by root
sudo docker run -v ~/data:/data  # Files created as root
```

## How to remediate

### Find root-owned files

**Find all root-owned files in home directory**:
```bash
# Find files owned by root
find ~ -user root -ls

# Count how many
find ~ -user root | wc -l

# Find only in specific directories
find ~/projects -user root -ls
```

**Find root-owned directories**:
```bash
# Find directories owned by root
find ~ -type d -user root -ls
```

### Fix ownership

**Change ownership back to your user**:
```bash
# Fix a specific file
sudo chown $USER:$USER ~/file.txt

# Fix a directory and all contents
sudo chown -R $USER:$USER ~/projects/myproject

# Fix all root-owned files in home directory (careful!)
sudo chown -R $USER:$USER ~

# Or be more selective
find ~ -user root -exec sudo chown $USER:$USER {} \;
```

**Verify ownership**:
```bash
# Check specific file
ls -la ~/file.txt
# Should show: -rw-r--r-- 1 youruser yourgroup ...

# Check directory
ls -ld ~/projects/myproject
# Should show: drwxr-xr-x 1 youruser yourgroup ...
```

### Fix permissions too

**After fixing ownership, fix permissions**:
```bash
# Fix file permissions (644 for files, 755 for directories)
find ~ -user $USER -type f -exec chmod 644 {} \;
find ~ -user $USER -type d -exec chmod 755 {} \;

# Or use a safer approach for specific directory
chmod -R u+rwX,go+rX,go-w ~/projects/myproject
```

**Fix executable scripts**:
```bash
# Make scripts executable
find ~ -name "*.sh" -exec chmod 755 {} \;
```

### Prevent future issues

**Avoid using sudo in home directory**:
```bash
# Bad
sudo vim ~/.bashrc
sudo npm install
sudo pip install package

# Good
vim ~/.bashrc  # No sudo needed
npm install  # Run as your user
pip install --user package  # Install to user directory
```

**Use proper permissions for tools**:
```bash
# For npm - fix npm permissions
mkdir -p ~/.npm-global
npm config set prefix '~/.npm-global'
export PATH=~/.npm-global/bin:$PATH

# For pip - use user installs
pip install --user package

# For Docker - add user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

### Platform-specific fixes

**macOS**:
```bash
# Find root-owned files
sudo find ~ -user root -ls

# Fix ownership
sudo chown -R $USER:staff ~/projects

# Fix common locations
sudo chown -R $USER:staff ~/Library/Caches
sudo chown -R $USER:staff ~/.npm
```

**Linux**:
```bash
# Find root-owned files
find ~ -user root -ls

# Fix ownership (use your primary group)
sudo chown -R $USER:$USER ~/projects

# Fix common locations
sudo chown -R $USER:$USER ~/.cache
sudo chown -R $USER:$USER ~/.local
sudo chown -R $USER:$USER ~/.config
```

**Windows (WSL)**:
```bash
# WSL has different ownership model
# Fix ownership in WSL filesystem
sudo chown -R $USER:$USER ~/projects

# Note: Windows filesystem (/mnt/c) has different permissions
```

### Fix common scenarios

**Node.js/npm**:
```bash
# If node_modules is owned by root
sudo chown -R $USER:$USER node_modules
sudo chown $USER:$USER package-lock.json

# Prevent future issues
npm config set prefix ~/.npm-global
```

**Python/pip**:
```bash
# If .local is owned by root
sudo chown -R $USER:$USER ~/.local

# Use --user flag for pip
pip install --user package
```

**Docker volumes**:
```bash
# If Docker created root-owned files
sudo chown -R $USER:$USER ~/docker-data

# Use user mapping in docker-compose.yml
services:
  app:
    user: "${UID}:${GID}"
    volumes:
      - ./data:/data
```

**Git repositories**:
```bash
# If .git is owned by root
sudo chown -R $USER:$USER ~/projects/myrepo

# Fix Git safe directory warning
git config --global --add safe.directory ~/projects/myrepo
```

### Best practices

1. **Never use sudo for files in home directory**:
   ```bash
   # If you need to edit a file, don't use sudo
   # If the file is owned by root, fix ownership first
   sudo chown $USER:$USER file.txt
   vim file.txt
   ```

2. **Use user-specific installation paths**:
   ```bash
   # npm
   npm config set prefix ~/.npm-global
   
   # pip
   pip install --user package
   
   # cargo
   # Already installs to ~/.cargo by default
   ```

3. **Add user to necessary groups** instead of using sudo:
   ```bash
   # Docker
   sudo usermod -aG docker $USER
   
   # VirtualBox
   sudo usermod -aG vboxusers $USER
   
   # Log out and back in for changes to take effect
   ```

4. **Use proper Docker user mapping**:
   ```bash
   # Run container as your user
   docker run --user $(id -u):$(id -g) -v ~/data:/data image
   
   # Or in docker-compose.yml
   user: "${UID}:${GID}"
   ```

5. **Audit home directory regularly**:
   ```bash
   # Check for root-owned files weekly
   find ~ -user root -ls
   
   # Add to cron
   0 0 * * 0 find ~ -user root -ls | mail -s "Root-owned files in home" you@example.com
   ```

6. **Use configuration management**:
   ```bash
   # Use dotfiles repository
   # Manage configs with version control
   # Never edit system files directly
   ```

7. **Document when sudo is necessary**:
   ```bash
   # If you must use sudo, document why
   # And fix ownership immediately after
   sudo some-command
   sudo chown -R $USER:$USER ~/affected-directory
   ```

### Quick fix script

**Create a script to fix common issues**:
```bash
#!/bin/bash
# fix-home-ownership.sh

echo "Finding root-owned files in home directory..."
ROOT_FILES=$(find ~ -user root 2>/dev/null)

if [ -z "$ROOT_FILES" ]; then
  echo "No root-owned files found."
  exit 0
fi

echo "Found root-owned files:"
echo "$ROOT_FILES"
echo ""
read -p "Fix ownership? (yes/no): " confirm

if [ "$confirm" = "yes" ]; then
  echo "Fixing ownership..."
  find ~ -user root -exec sudo chown $USER:$USER {} \;
  echo "Done!"
else
  echo "Aborted."
fi
```

**Make it executable and run**:
```bash
chmod +x fix-home-ownership.sh
./fix-home-ownership.sh
```

