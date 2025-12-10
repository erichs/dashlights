# Rotting Secrets

## What this is

This signal detects **long-lived sensitive files** (older than 7 days) in common directories. It identifies the same file types as the "Dumpster Fire" signal but specifically flags files that have been sitting around for a while - likely forgotten.

**Scanned locations:**
- `~/Downloads`
- `~/Desktop`
- Current working directory (`$PWD`)
- `/tmp`

**File types detected:**
- Database dumps: `*.sql`, `*.sqlite`, `*.db`, `*.bak`, `dump-*`, `backup-*`, `*prod*`
- Network artifacts: `*.har`, `*.pcap`
- Key files: `*.keychain`, `*.pem`, `*.pfx`, `*.jks`

**Age threshold:** Files with modification time older than 7 days

## Why this matters

**Forgotten Sensitive Data is a Liability:**
- Old database dumps may contain stale but still-sensitive data
- Forgotten HAR files can contain valid API tokens that haven't been rotated
- PCAP files may contain credentials for systems still in use
- The longer sensitive files sit around, the higher the chance of accidental exposure

**The 7-Day Threshold:**
- If you needed a file for active debugging, you'd use it within a week
- Files older than 7 days are likely forgotten or no longer actively needed
- This threshold balances catching "rotting" files vs. false positives on active work

**Common Scenarios:**
- "Oh right, that prod pg dump I grabbed three weeks ago..."
- "I forgot I downloaded that certificate for testing last month..."
- "That HAR file has been sitting there since I debugged that auth issue..."

## How to remediate

### Find old sensitive files

```bash
# Find SQL files older than 7 days
find ~/Downloads ~/Desktop /tmp -name "*.sql" -mtime +7 2>/dev/null

# Find all old sensitive files
find ~/Downloads ~/Desktop /tmp \
  \( -name "*.sql" -o -name "*.sqlite" -o -name "*.db" -o -name "*.bak" \
     -o -name "*.har" -o -name "*.pcap" -o -name "*.pem" -o -name "*.pfx" \
     -o -name "dump-*" -o -name "backup-*" \) \
  -mtime +7 2>/dev/null
```

### Review and clean up

**Check file contents before deleting:**
```bash
# For SQL files - check what's in them
head -50 ~/Downloads/old-dump.sql

# For HAR files - check endpoints
grep -o '"url":"[^"]*"' ~/Downloads/request.har | head -10
```

**Delete files you no longer need:**
```bash
# Remove individual file
rm ~/Downloads/prod-backup-2024-01-01.sql

# Remove all old SQL files from Downloads (careful!)
find ~/Downloads -name "*.sql" -mtime +7 -exec rm {} \;
```

### Set up automatic cleanup

**macOS - Periodic cleanup script:**
```bash
# Create cleanup script
cat > ~/bin/cleanup-sensitive-files.sh << 'EOF'
#!/bin/bash
# Clean up old sensitive files from common directories

find ~/Downloads -type f \( \
  -name "*.sql" -o -name "*.sqlite" -o -name "*.db" -o -name "*.bak" \
  -o -name "*.har" -o -name "*.pcap" \
  -o -name "dump-*" -o -name "backup-*" \
\) -mtime +7 -delete

echo "Cleaned up old sensitive files from Downloads"
EOF
chmod +x ~/bin/cleanup-sensitive-files.sh
```

**Schedule with launchd (macOS):**
```bash
# Create a LaunchAgent to run weekly
cat > ~/Library/LaunchAgents/com.user.cleanup-sensitive.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.cleanup-sensitive</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>~/bin/cleanup-sensitive-files.sh</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Weekday</key>
        <integer>1</integer>
        <key>Hour</key>
        <integer>9</integer>
    </dict>
</dict>
</plist>
EOF
launchctl load ~/Library/LaunchAgents/com.user.cleanup-sensitive.plist
```

**Linux - cron job:**
```bash
# Add weekly cleanup cron job
(crontab -l 2>/dev/null; echo "0 9 * * 1 ~/bin/cleanup-sensitive-files.sh") | crontab -
```

### Best practices

1. **Don't download to Downloads** - Use project-specific temp directories
2. **Set calendar reminders** - Weekly cleanup of sensitive file locations
3. **Use time-limited sharing** - Services like `wormhole` expire automatically
4. **Enable cloud sync exclusions** - Exclude Downloads from iCloud/Dropbox sync

## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_ROTTING_SECRETS=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).

