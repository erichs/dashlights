# HTTP Proxy Active

## What this is

This signal detects when HTTP/HTTPS proxy environment variables (`HTTP_PROXY`, `HTTPS_PROXY`, `http_proxy`, `https_proxy`) are set. While proxies are legitimate for corporate networks, they can also be used by attackers to intercept traffic, steal credentials, and monitor your activity.

## Why this matters

**Security Risk**:
- **Man-in-the-middle attacks**: Malicious proxies can intercept and modify all HTTP/HTTPS traffic
- **Credential theft**: Proxies can capture authentication tokens, API keys, and passwords
- **Data exfiltration**: Attackers can monitor and steal sensitive data passing through the proxy
- **SSL/TLS interception**: Some proxies perform SSL inspection, decrypting your encrypted traffic

**Legitimate uses**:
- Corporate networks that require proxy for internet access
- Development tools like Charles Proxy or Burp Suite for debugging
- Privacy tools like Tor or VPNs

**Malicious uses**:
- Malware sets proxy to route traffic through attacker-controlled servers
- Attackers use proxies to bypass network security controls
- Proxies used for persistence and command-and-control

## How to remediate

### Verify the proxy is legitimate

**Check what proxy is set**:
```bash
echo $HTTP_PROXY
echo $HTTPS_PROXY
echo $http_proxy
echo $https_proxy
echo $ALL_PROXY
echo $NO_PROXY
```

**Verify it's your organization's proxy**:
```bash
# Check if it's your corporate proxy
# Example: http://proxy.company.com:8080

# If it's an unknown IP or domain, investigate immediately
```

### Remove proxy if not needed

**Unset proxy variables**:
```bash
unset HTTP_PROXY
unset HTTPS_PROXY
unset http_proxy
unset https_proxy
unset ALL_PROXY
unset NO_PROXY

# Verify
env | grep -i proxy
```

**Remove from shell configuration**:
```bash
# Find where proxy is set
grep -r "proxy" ~/.bashrc ~/.bash_profile ~/.zshrc ~/.zprofile /etc/profile

# Edit and remove proxy settings
nano ~/.bashrc  # or ~/.zshrc

# Remove lines like:
# export HTTP_PROXY=http://proxy.example.com:8080
# export HTTPS_PROXY=http://proxy.example.com:8080
```

**Reload shell configuration**:
```bash
source ~/.bashrc  # or ~/.zshrc
# Or start a new shell
exec $SHELL
```

### Check system-wide proxy settings

**macOS**:
```bash
# Check system proxy settings
networksetup -getwebproxy Wi-Fi
networksetup -getsecurewebproxy Wi-Fi

# Disable if not needed
sudo networksetup -setwebproxystate Wi-Fi off
sudo networksetup -setsecurewebproxystate Wi-Fi off
```

**Linux (GNOME)**:
```bash
# Check GNOME proxy settings
gsettings get org.gnome.system.proxy mode
gsettings get org.gnome.system.proxy.http host
gsettings get org.gnome.system.proxy.http port

# Disable if not needed
gsettings set org.gnome.system.proxy mode 'none'
```

**Windows**:
```powershell
# Check proxy settings
Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings' | Select-Object ProxyServer, ProxyEnable

# Disable proxy
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings' -Name ProxyEnable -Value 0
```

### Investigate suspicious proxies

**If proxy points to unknown server**:
```bash
# Get proxy address
PROXY_HOST=$(echo $HTTP_PROXY | sed 's|http://||' | sed 's|https://||' | cut -d: -f1)

# Look up the IP/domain
nslookup $PROXY_HOST
whois $PROXY_HOST

# Check if it's listening
nc -zv $PROXY_HOST 8080

# Check for suspicious processes
ps aux | grep -i proxy
lsof -i -P | grep -i proxy
```

**Signs of malicious proxy**:
- Unknown IP address or domain
- Proxy running on localhost (127.0.0.1) but you didn't start it
- Proxy process owned by another user
- Proxy in unusual location (/tmp, /var/tmp)

### Configure proxy correctly (if needed)

**For corporate proxy**:
```bash
# Add to ~/.bashrc or ~/.zshrc
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
export NO_PROXY=localhost,127.0.0.1,.company.com

# For authenticated proxy
export HTTP_PROXY=http://username:password@proxy.company.com:8080
# Better: use .netrc or credential manager instead of password in env var
```

**Exclude internal domains**:
```bash
# Don't proxy internal traffic
export NO_PROXY=localhost,127.0.0.1,.internal,.local,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
```

### Tool-specific proxy configuration

**Git**:
```bash
# Check Git proxy
git config --global --get http.proxy
git config --global --get https.proxy

# Set Git proxy
git config --global http.proxy http://proxy.company.com:8080
git config --global https.proxy http://proxy.company.com:8080

# Unset Git proxy
git config --global --unset http.proxy
git config --global --unset https.proxy
```

**npm**:
```bash
# Check npm proxy
npm config get proxy
npm config get https-proxy

# Set npm proxy
npm config set proxy http://proxy.company.com:8080
npm config set https-proxy http://proxy.company.com:8080

# Unset npm proxy
npm config delete proxy
npm config delete https-proxy
```

**Docker**:
```bash
# Configure Docker daemon proxy
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo nano /etc/systemd/system/docker.service.d/http-proxy.conf

# Add:
[Service]
Environment="HTTP_PROXY=http://proxy.company.com:8080"
Environment="HTTPS_PROXY=http://proxy.company.com:8080"
Environment="NO_PROXY=localhost,127.0.0.1"

# Reload
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### Best practices

1. **Only use trusted proxies**:
   - Corporate proxies managed by your IT department
   - Development tools you installed yourself
   - Well-known privacy tools (Tor, VPNs)

2. **Use proxy auto-config (PAC)** instead of manual settings:
   ```bash
   # Instead of HTTP_PROXY, use PAC file
   export auto_proxy=http://proxy.company.com/proxy.pac
   ```

3. **Monitor proxy usage**:
   ```bash
   # Add to your security monitoring
   if [ -n "$HTTP_PROXY" ]; then
     echo "Proxy active: $HTTP_PROXY" | logger -t security
   fi
   ```

4. **Use certificate pinning** to prevent SSL interception:
   ```bash
   # In your application, pin expected certificates
   # This prevents proxy from intercepting SSL
   ```

5. **Audit proxy settings** regularly:
   ```bash
   # Check all proxy-related settings
   env | grep -i proxy
   git config --global --get-regexp proxy
   npm config list | grep proxy
   ```

6. **Use VPN instead of proxy** when possible:
   - VPNs encrypt all traffic, not just HTTP
   - Harder for attackers to intercept
   - Better for security and privacy

### If you suspect malicious proxy

1. **Disconnect from network** immediately
2. **Document the proxy settings**:
   ```bash
   env | grep -i proxy > /tmp/proxy-evidence.txt
   ps aux | grep -i proxy >> /tmp/proxy-evidence.txt
   ```
3. **Remove the proxy** settings
4. **Scan for malware**:
   ```bash
   # Use antivirus or malware scanner
   clamscan -r /
   ```
5. **Check for persistence**:
   ```bash
   # Check startup scripts
   cat /etc/profile /etc/bash.bashrc ~/.bashrc ~/.zshrc

   # Check cron jobs
   crontab -l
   sudo crontab -l
   ```
6. **Rotate all credentials** that may have been intercepted
7. **Report to security team**


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_PROXY_ACTIVE=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
