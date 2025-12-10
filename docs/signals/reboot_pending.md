# Reboot Pending

## What this is

This signal detects when your system has a pending reboot, typically after installing system updates or kernel patches. On Linux, this is indicated by the presence of `/var/run/reboot-required` file. A pending reboot means security patches and critical updates haven't taken effect yet.

## Why this matters

**Security Risk**:
- **Unpatched vulnerabilities**: Security patches don't take effect until you reboot
- **Kernel exploits**: Kernel updates require a reboot to protect against known exploits
- **Privilege escalation**: Attackers can exploit unpatched kernel vulnerabilities
- **Compliance violations**: Security policies often require timely patching and reboots

**Operational Risk**:
- **System instability**: Running with pending updates can cause unexpected behavior
- **Forced reboots**: System may reboot automatically at an inconvenient time
- **Service disruptions**: Delayed reboots can lead to emergency maintenance windows

**Why reboots matter**:
- Kernel security patches only apply after reboot
- Some library updates require process restarts or reboots
- System stability improvements need reboot to take effect

## How to remediate

### Check what requires reboot

**Check if reboot is required**:
```bash
# Check for reboot-required file
ls -la /var/run/reboot-required

# See which packages require reboot
cat /var/run/reboot-required.pkgs
```

**Check for kernel updates**:
```bash
# Current running kernel
uname -r

# Installed kernels
dpkg -l | grep linux-image

# Compare to see if newer kernel is installed
```

### Reboot the system

**Save your work and reboot**:
```bash
# Save all work first!

# Reboot immediately
sudo reboot

# Or schedule a reboot
sudo shutdown -r +10  # Reboot in 10 minutes
sudo shutdown -r 20:00  # Reboot at 8 PM

# Cancel scheduled reboot
sudo shutdown -c
```

**Graceful reboot**:
```bash
# 1. Save all work
# 2. Close applications
# 3. Notify team members
# 4. Reboot
sudo reboot
```

### Check what's running after reboot

**Verify kernel version**:
```bash
# Check running kernel
uname -r

# Should match the latest installed kernel
dpkg -l | grep linux-image | grep ^ii
```

**Verify services restarted**:
```bash
# Check service status
sudo systemctl status

# Check specific services
sudo systemctl status docker
sudo systemctl status nginx
```

### Alternative: Restart services without reboot

**For library updates** (not kernel):
```bash
# Check which services need restart
sudo checkrestart  # From debian-goodies package

# Or use needrestart
sudo needrestart

# Restart specific services
sudo systemctl restart service-name
```

**Restart user session** (for user-space updates):
```bash
# Log out and log back in
# Or restart display manager
sudo systemctl restart gdm  # GNOME
sudo systemctl restart lightdm  # LightDM
```

### Platform-specific guidance

**Ubuntu/Debian**:
```bash
# Check if reboot required
cat /var/run/reboot-required

# See which packages
cat /var/run/reboot-required.pkgs

# Reboot
sudo reboot
```

**RHEL/CentOS/Fedora**:
```bash
# Check if reboot required
sudo needs-restarting -r

# List services that need restart
sudo needs-restarting -s

# Reboot
sudo reboot
```

**macOS**:
```bash
# Check for pending updates
softwareupdate -l

# Install updates (may require reboot)
sudo softwareupdate -i -a

# Reboot
sudo shutdown -r now
```

**Windows**:
```powershell
# Check for pending reboot
Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired'

# Reboot
shutdown /r /t 0
```

### Schedule reboots during maintenance windows

**Use cron for scheduled reboots**:
```bash
# Edit root crontab
sudo crontab -e

# Reboot every Sunday at 3 AM
0 3 * * 0 /sbin/shutdown -r now

# Or use at for one-time scheduled reboot
echo "sudo reboot" | at 3:00 AM tomorrow
```

**Use systemd timers**:
```bash
# Create reboot timer
sudo systemctl edit --force --full reboot.timer

[Unit]
Description=Weekly reboot

[Timer]
OnCalendar=Sun *-*-* 03:00:00
Persistent=true

[Install]
WantedBy=timers.target

# Enable timer
sudo systemctl enable reboot.timer
sudo systemctl start reboot.timer
```

### Best practices

1. **Reboot regularly** to apply security patches:
   ```bash
   # Schedule weekly reboots during maintenance windows
   # Sunday 3 AM is common for non-critical systems
   ```

2. **Test updates in staging** before production:
   ```bash
   # Apply updates to staging first
   # Verify everything works
   # Then apply to production
   ```

3. **Use unattended-upgrades** for automatic security updates:
   ```bash
   # Install unattended-upgrades
   sudo apt install unattended-upgrades

   # Configure automatic reboots
   sudo dpkg-reconfigure -plow unattended-upgrades

   # Edit /etc/apt/apt.conf.d/50unattended-upgrades
   Unattended-Upgrade::Automatic-Reboot "true";
   Unattended-Upgrade::Automatic-Reboot-Time "03:00";
   ```

4. **Monitor uptime**:
   ```bash
   # Check system uptime
   uptime

   # Alert if uptime is too long (>30 days)
   if [ $(awk '{print int($1)}' /proc/uptime) -gt 2592000 ]; then
     echo "System uptime exceeds 30 days - reboot needed"
   fi
   ```

5. **Document reboot procedures**:
   - Pre-reboot checklist
   - Services to verify after reboot
   - Rollback plan if issues occur

6. **Use live patching** for critical systems (if available):
   ```bash
   # Ubuntu Livepatch (requires Ubuntu Pro)
   sudo ua enable livepatch

   # Or use KernelCare, Ksplice, etc.
   ```

7. **Notify users** before rebooting:
   ```bash
   # Send wall message
   sudo wall "System will reboot in 10 minutes for security updates"

   # Wait 10 minutes
   sleep 600

   # Reboot
   sudo reboot
   ```

### For servers and production systems

1. **Use rolling reboots** for high availability:
   ```bash
   # Reboot servers one at a time
   # Ensure load balancer removes server from pool
   # Reboot
   # Verify health checks pass
   # Move to next server
   ```

2. **Use configuration management**:
   ```yaml
   # Ansible example
   - name: Reboot if required
     reboot:
       reboot_timeout: 600
     when: reboot_required_file.stat.exists
   ```

3. **Monitor for failed reboots**:
   ```bash
   # Check if system came back up
   # Alert if reboot takes too long
   # Have console access ready
   ```

4. **Keep kernel up to date**:
   ```bash
   # Check for kernel updates weekly
   sudo apt update
   sudo apt list --upgradable | grep linux-image
   ```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_REBOOT_PENDING=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
