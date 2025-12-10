# Disk Space Critical

## What this is

This signal detects when your root filesystem (`/`) is more than 90% full. Running out of disk space can cause cascading failures across your system, particularly affecting logging, security auditing, and application functionality.

## Why this matters

**Security Impact**:
- **Audit trail loss**: When disk is full, system logs (`/var/log`) stop being written, creating blind spots for security monitoring and incident response
- **Authentication failures**: Some authentication systems fail when they can't write session data or logs
- **Backup failures**: Automated backups may fail silently, leaving you without disaster recovery options
- **Update failures**: Security patches can't be installed if there's no disk space

**Operational Impact**:
- **Application crashes**: Apps that need to write temporary files, caches, or databases will fail
- **Database corruption**: Databases can become corrupted if they can't write transaction logs
- **System instability**: The OS itself may become unstable or unresponsive
- **Build failures**: CI/CD pipelines fail when they can't write build artifacts

**Why 90% threshold**: Most filesystems experience performance degradation above 90% usage, and many applications start failing before reaching 100%.

## How to remediate

### Immediate actions

**Check current disk usage**:
```bash
# Overview of all filesystems
df -h

# Detailed view of root filesystem
df -h /

# Human-readable with inodes
df -hi /
```

**Find large directories**:
```bash
# Top 10 largest directories in root
sudo du -h / 2>/dev/null | sort -rh | head -20

# Specific common culprits
sudo du -sh /var/log/* | sort -rh
sudo du -sh /tmp/* | sort -rh
sudo du -sh /var/cache/* | sort -rh
sudo du -sh ~/.cache/* | sort -rh
```

### Clean up common space hogs

**1. Clean package manager caches**:

**macOS (Homebrew)**:
```bash
# Clean old versions
brew cleanup

# See what will be removed
brew cleanup -n

# Remove all caches
rm -rf ~/Library/Caches/Homebrew/*
```

**Linux (apt)**:
```bash
# Clean package cache
sudo apt-get clean
sudo apt-get autoclean
sudo apt-get autoremove
```

**Linux (yum/dnf)**:
```bash
# Clean package cache
sudo yum clean all
# or
sudo dnf clean all
```

**2. Clean log files**:
```bash
# Truncate large log files (keeps file, removes content)
sudo truncate -s 0 /var/log/large-file.log

# Remove old rotated logs
sudo find /var/log -name "*.gz" -mtime +30 -delete
sudo find /var/log -name "*.old" -mtime +30 -delete

# Clean journal logs (systemd)
sudo journalctl --vacuum-time=7d
sudo journalctl --vacuum-size=500M
```

**3. Clean temporary files**:
```bash
# Clean /tmp (be careful!)
sudo find /tmp -type f -atime +7 -delete

# Clean user cache
rm -rf ~/.cache/*

# Clean thumbnail cache
rm -rf ~/.thumbnails/*
```

**4. Clean Docker (if installed)**:
```bash
# Remove unused containers, images, networks
docker system prune -a

# Remove volumes too (careful!)
docker system prune -a --volumes

# See what will be removed
docker system df
```

**5. Find and remove large files**:
```bash
# Find files larger than 100MB
sudo find / -type f -size +100M -exec ls -lh {} \; 2>/dev/null | sort -k5 -rh | head -20

# Find files larger than 1GB
sudo find / -type f -size +1G -exec ls -lh {} \; 2>/dev/null
```

### Platform-specific cleanup

**macOS**:
```bash
# Clean Xcode caches
rm -rf ~/Library/Developer/Xcode/DerivedData/*

# Clean iOS simulators
xcrun simctl delete unavailable

# Clean Time Machine local snapshots
tmutil listlocalsnapshots /
sudo tmutil deletelocalsnapshots <snapshot-date>
```

**Linux**:
```bash
# Clean old kernels (Ubuntu/Debian)
sudo apt-get autoremove --purge

# List installed kernels
dpkg --list | grep linux-image

# Remove specific old kernel
sudo apt-get remove linux-image-X.X.X-XX-generic
```

### Long-term solutions

1. **Set up log rotation**:
   ```bash
   # Check logrotate configuration
   cat /etc/logrotate.conf

   # Add custom rotation for your app
   sudo nano /etc/logrotate.d/myapp
   ```

2. **Monitor disk usage**:
   ```bash
   # Add to cron for daily checks
   echo "0 9 * * * df -h / | mail -s 'Disk Usage Report' you@example.com" | crontab -
   ```

3. **Expand disk space** (if on VM/cloud):
   ```bash
   # AWS: Resize EBS volume via console, then:
   sudo growpart /dev/xvda 1
   sudo resize2fs /dev/xvda1
   ```

4. **Move data to separate partition**:
   ```bash
   # Move /var/log to separate partition
   # (requires planning and downtime)
   ```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_DISK_SPACE=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
