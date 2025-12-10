# Dumpster Fire

## What this is

This signal detects sensitive-looking files that have accumulated in common "junk drawer" locations:

- `~/Downloads` - Files downloaded from browsers, email, etc.
- `~/Desktop` - Files temporarily saved for quick access
- Current working directory (`$PWD`) - Project directories
- `/tmp` - System temporary directory

The signal uses **name-only pattern matching** (no content scanning) to quickly identify potentially sensitive files:

**Database files and backups:**
- `*.sql`, `*.sqlite`, `*.db`, `*.bak`
- `dump-*`, `backup-*`
- Files containing `prod` in the name

**Network/security artifacts:**
- `*.har` (HTTP Archive files - contain full requests/responses)
- `*.pcap` (Network packet captures)
- `*.keychain`, `*.pem`, `*.pfx`, `*.jks` (Key and certificate files)

## Why this matters

**Data Sprawl is a Security Risk:**
- Database dumps often contain production data with PII, credentials, or business secrets
- HAR files capture authentication tokens, session cookies, and API keys
- PCAP files may contain unencrypted passwords and sensitive traffic
- Key files left in common directories are at higher risk of accidental exposure

**Common Scenarios:**
- "I grabbed a prod database dump to debug an issue three weeks ago..."
- "I exported this HAR file to debug an API call and forgot about it..."
- "I downloaded this PEM file from the secrets manager to test locally..."

**Why These Directories:**
- `~/Downloads` and `~/Desktop` are often synced to cloud services (Dropbox, iCloud, etc.)
- `/tmp` may be world-readable and persists across reboots on some systems
- Project directories may accidentally commit sensitive files

## How to remediate

### Review detected files

Check what sensitive files have accumulated:

```bash
# List SQL files in Downloads
ls -la ~/Downloads/*.sql 2>/dev/null

# List all database files in common locations
find ~/Downloads ~/Desktop /tmp -name "*.sql" -o -name "*.sqlite" -o -name "*.db" 2>/dev/null

# Check for dump files
find ~/Downloads ~/Desktop /tmp -name "dump-*" -o -name "backup-*" 2>/dev/null
```

### Clean up sensitive files

**Securely delete files you no longer need:**
```bash
# macOS: Use srm for secure removal (if available) or regular rm
rm ~/Downloads/prod-database-dump.sql

# Or use shred on Linux
shred -u ~/Downloads/prod-database-dump.sql
```

**Move files to secure storage:**
```bash
# Move to an encrypted volume or secure location
mv ~/Downloads/important-backup.sql ~/secure-storage/
```

### Prevent future sprawl

**Use project-specific locations:**
```bash
# Create a .local or .tmp directory in your project (add to .gitignore)
mkdir -p .local
echo ".local/" >> .gitignore
```

**Clean up regularly:**
```bash
# Add to your weekly routine
find ~/Downloads -name "*.sql" -mtime +7 -delete
find ~/Downloads -name "*.har" -mtime +7 -delete
```

**Use time-limited downloads:**
```bash
# Set up automatic cleanup of Downloads folder (macOS)
# System Preferences > General > Automatically delete items in Trash after 30 days
```

### For HAR files specifically

HAR files are particularly dangerous as they contain full HTTP request/response data:

```bash
# Never share HAR files without sanitization
# Use a HAR sanitizer before sharing:
# https://nicedoc.io/nickreese/har-sanitizer

# Delete HAR files immediately after use
rm ~/Downloads/*.har
```

### For key files

Key and certificate files should be stored securely:

```bash
# Move keys to proper location with restricted permissions
mv ~/Downloads/server.pem ~/.ssh/
chmod 600 ~/.ssh/server.pem

# Or use a secrets manager
# 1Password, HashiCorp Vault, AWS Secrets Manager, etc.
```

## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_DUMPSTER_FIRE=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).

