# Time Drift

## What this is

This signal detects when your system clock is more than 5 minutes off from actual time. Time drift can cause authentication failures, certificate validation errors, and security issues.

## Why this matters

**Authentication Failures**:
- **Kerberos fails**: Requires time sync within 5 minutes
- **OAuth/SAML fails**: Time-based tokens become invalid
- **API authentication fails**: Many APIs reject requests with incorrect timestamps
- **2FA/TOTP fails**: Time-based one-time passwords don't work

**Security Issues**:
- **Certificate validation fails**: SSL/TLS certificates appear expired or not yet valid
- **Log correlation impossible**: Can't correlate events across systems
- **Replay attack vulnerability**: Attackers can replay old requests
- **Audit trail corruption**: Timestamps in logs are incorrect

**Operational Impact**:
- **Build failures**: Time-sensitive build processes fail
- **Scheduled tasks run at wrong time**: Cron jobs, backups, etc.
- **Database replication issues**: Time-based replication breaks
- **Monitoring alerts**: False positives/negatives due to time mismatch

## How to remediate

### Check current time

**Check system time**:
```bash
# Current system time
date

# Compare with actual time
# Should match: https://time.is
```

**Check time drift**:
```bash
# Linux - check NTP status
timedatectl status

# macOS - check time sync
sudo sntp -d time.apple.com

# Show time difference
ntpdate -q pool.ntp.org
```

### Sync time immediately

**Linux (systemd-timesyncd)**:
```bash
# Force immediate sync
sudo timedatectl set-ntp true

# Restart time sync service
sudo systemctl restart systemd-timesyncd

# Check status
timedatectl status

# Verify time is synced
timedatectl timesync-status
```

**Linux (ntpd)**:
```bash
# Stop ntpd
sudo systemctl stop ntpd

# Force sync
sudo ntpdate pool.ntp.org

# Start ntpd
sudo systemctl start ntpd

# Check status
ntpq -p
```

**Linux (chrony)**:
```bash
# Force sync
sudo chronyc makestep

# Check status
chronyc tracking
chronyc sources
```

**macOS**:
```bash
# Enable automatic time sync
sudo systemsetup -setusingnetworktime on

# Set time server
sudo systemsetup -setnetworktimeserver time.apple.com

# Force sync
sudo sntp -sS time.apple.com

# Verify
date
```

**Windows**:
```powershell
# Sync time
w32tm /resync

# Check status
w32tm /query /status

# Configure time server
w32tm /config /manualpeerlist:"time.windows.com" /syncfromflags:manual /update

# Restart time service
Restart-Service w32time
```

### Enable automatic time sync

**Linux (systemd-timesyncd)**:
```bash
# Enable NTP
sudo timedatectl set-ntp true

# Configure NTP servers
sudo nano /etc/systemd/timesyncd.conf

# Add:
[Time]
NTP=pool.ntp.org time.google.com time.cloudflare.com
FallbackNTP=time.nist.gov

# Restart service
sudo systemctl restart systemd-timesyncd

# Enable on boot
sudo systemctl enable systemd-timesyncd
```

**Linux (chrony)**:
```bash
# Install chrony
sudo apt install chrony  # Debian/Ubuntu
sudo yum install chrony  # RHEL/CentOS

# Configure
sudo nano /etc/chrony/chrony.conf

# Add NTP servers:
server pool.ntp.org iburst
server time.google.com iburst
server time.cloudflare.com iburst

# Restart
sudo systemctl restart chronyd
sudo systemctl enable chronyd

# Verify
chronyc tracking
```

**macOS**:
```bash
# Enable automatic time
sudo systemsetup -setusingnetworktime on

# Set time server
sudo systemsetup -setnetworktimeserver time.apple.com

# Verify
sudo systemsetup -getusingnetworktime
sudo systemsetup -getnetworktimeserver
```

**Windows**:
```powershell
# Set time server
w32tm /config /manualpeerlist:"time.windows.com,0x8 time.google.com,0x8" /syncfromflags:manual /reliable:YES /update

# Start service
Set-Service w32time -StartupType Automatic
Start-Service w32time

# Sync
w32tm /resync

# Verify
w32tm /query /status
```

### Set correct timezone

**Linux**:
```bash
# List available timezones
timedatectl list-timezones

# Set timezone
sudo timedatectl set-timezone America/New_York

# Verify
timedatectl
```

**macOS**:
```bash
# List timezones
sudo systemsetup -listtimezones

# Set timezone
sudo systemsetup -settimezone America/New_York

# Verify
sudo systemsetup -gettimezone
```

**Windows**:
```powershell
# List timezones
Get-TimeZone -ListAvailable

# Set timezone
Set-TimeZone -Name "Eastern Standard Time"

# Verify
Get-TimeZone
```

### Troubleshooting

**If time keeps drifting**:
```bash
# Check if NTP service is running
sudo systemctl status systemd-timesyncd  # or chronyd or ntpd

# Check NTP servers are reachable
ntpdate -q pool.ntp.org

# Check for hardware clock issues
sudo hwclock --show
sudo hwclock --systohc  # Sync hardware clock to system clock
```

**If running in VM**:
```bash
# VirtualBox - enable time sync
VBoxManage guestproperty set "VM Name" "/VirtualBox/GuestAdd/VBoxService/--timesync-set-threshold" 1000

# VMware - enable time sync in VM settings
# Or install VMware Tools

# Docker - use host time
docker run --rm -v /etc/localtime:/etc/localtime:ro image

# WSL - sync with Windows
sudo hwclock -s
```

**If firewall blocks NTP**:
```bash
# Allow NTP (UDP port 123)
sudo ufw allow 123/udp

# Or use HTTPS-based time sync
# Install htpdate
sudo apt install htpdate
sudo htpdate -s www.google.com
```

### Best practices

1. **Use multiple NTP servers**:
   ```
   server 0.pool.ntp.org iburst
   server 1.pool.ntp.org iburst
   server 2.pool.ntp.org iburst
   server 3.pool.ntp.org iburst
   ```

2. **Use geographically close servers**:
   ```
   # North America
   server north-america.pool.ntp.org

   # Europe
   server europe.pool.ntp.org

   # Asia
   server asia.pool.ntp.org
   ```

3. **Monitor time drift**:
   ```bash
   # Check drift regularly
   chronyc tracking | grep "System time"

   # Alert if drift > 1 second
   ```

4. **Use hardware clock**:
   ```bash
   # Sync hardware clock
   sudo hwclock --systohc

   # Check hardware clock
   sudo hwclock --show
   ```

5. **For servers, use local NTP server**:
   ```
   # Point to internal NTP server
   server ntp.company.local iburst
   ```

6. **For cloud instances**:
   ```
   # AWS
   server 169.254.169.123 prefer iburst

   # GCP
   server metadata.google.internal prefer iburst

   # Azure
   server time.windows.com prefer iburst
   ```

### Recommended NTP servers

**Public NTP pools**:
- `pool.ntp.org` - Global pool
- `time.google.com` - Google Public NTP
- `time.cloudflare.com` - Cloudflare NTP
- `time.nist.gov` - NIST time servers

**Cloud provider NTP**:
- AWS: `169.254.169.123`
- GCP: `metadata.google.internal`
- Azure: `time.windows.com`

### Security considerations

1. **Use NTP authentication** (for critical systems)
2. **Firewall NTP traffic** to trusted servers only
3. **Monitor for NTP amplification attacks**
4. **Use NTS (Network Time Security)** if available
5. **Validate time sources** are legitimate


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_TIME_DRIFT=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
