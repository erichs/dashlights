# Untracked Crypto Keys

## What this is

This signal detects cryptographic key files (`.pem`, `.key`, `.crt`, `.p12`, `.pfx`) in your project directory that are not listed in `.gitignore`. These files often contain private keys, certificates, and other sensitive cryptographic material that should never be committed to version control.

## Why this matters

**Security Risk**:
- **Private key exposure**: Committing private keys exposes them to anyone with repository access
- **Certificate theft**: SSL/TLS certificates and private keys can be stolen
- **Credential compromise**: Many key files contain authentication credentials
- **Public exposure**: If repository is public or becomes public, keys are exposed to the internet

**Real-world impact**:
- **Man-in-the-middle attacks**: Stolen SSL keys allow traffic interception
- **Impersonation**: Attackers can impersonate your services
- **Data decryption**: Encrypted data can be decrypted with stolen keys
- **Code signing abuse**: Stolen code signing certificates allow malware distribution

**Compliance**: Many regulations (PCI-DSS, HIPAA, SOC 2) require cryptographic keys to be protected and not stored in version control.

## How to remediate

### Add crypto files to .gitignore

**Add common crypto file patterns**:
```bash
# Add to .gitignore
cat >> .gitignore <<'EOF'
# Cryptographic keys and certificates
*.pem
*.key
*.crt
*.cer
*.p12
*.pfx
*.jks
*.keystore
*.truststore

# SSH keys
id_rsa
id_dsa
id_ecdsa
id_ed25519
*.ppk

# GPG keys
*.gpg
*.asc
secring.*

# SSL/TLS
*.csr
*.der

# AWS/Cloud credentials
*.credentials
credentials
EOF
```

**Verify files are ignored**:
```bash
git status
# Should not show any .pem, .key, .crt files
```

### Remove from Git if already committed

**Check if crypto files are tracked**:
```bash
# Find tracked crypto files
git ls-files | grep -E '\.(pem|key|crt|p12|pfx)$'
```

**Remove from Git** (keep local files):
```bash
# Remove specific file
git rm --cached server.key

# Remove all crypto files
git ls-files | grep -E '\.(pem|key|crt|p12|pfx)$' | xargs git rm --cached

# Commit removal
git add .gitignore
git commit -m "Remove crypto keys from version control and add to .gitignore"
```

**Purge from Git history** (if keys were committed):
```bash
# WARNING: Rewrites history - coordinate with team!

# Using git-filter-repo (recommended)
git filter-repo --path-glob '*.pem' --invert-paths
git filter-repo --path-glob '*.key' --invert-paths
git filter-repo --path-glob '*.crt' --invert-paths
git filter-repo --path-glob '*.p12' --invert-paths

# Or using BFG Repo-Cleaner
bfg --delete-files '*.{pem,key,crt,p12,pfx}'
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Force push (WARNING: destructive)
git push origin --force --all
```

### Rotate exposed keys

**If crypto keys were committed, assume they are compromised**:

1. **Generate new keys**:
   ```bash
   # Generate new SSL certificate
   openssl req -x509 -newkey rsa:4096 -keyout new-server.key -out new-server.crt -days 365 -nodes

   # Or use Let's Encrypt
   certbot certonly --standalone -d example.com
   ```

2. **Update all services** with new keys

3. **Revoke old certificates**:
   ```bash
   # Revoke SSL certificate
   certbot revoke --cert-path /path/to/old-cert.pem

   # Or contact your CA to revoke
   ```

4. **Monitor for unauthorized use** of old keys

### Use example files instead

**Create example files without real keys**:
```bash
# Create example certificate file
cat > server.crt.example <<'EOF'
# SSL Certificate
# Copy this file to server.crt and add your actual certificate
# DO NOT commit server.crt to Git

-----BEGIN CERTIFICATE-----
(Your certificate here)
-----END CERTIFICATE-----
EOF

# Create example key file
cat > server.key.example <<'EOF'
# SSL Private Key
# Copy this file to server.key and add your actual private key
# DO NOT commit server.key to Git

-----BEGIN PRIVATE KEY-----
(Your private key here)
-----END PRIVATE KEY-----
EOF

# Commit example files
git add server.crt.example server.key.example
git commit -m "Add example certificate files"
```

### Store keys securely

**Option 1: Use environment variables**:
```bash
# Store key content in environment variable
export SSL_PRIVATE_KEY="$(cat server.key)"
export SSL_CERTIFICATE="$(cat server.crt)"

# Application reads from environment
```

**Option 2: Use secret management**:
```bash
# AWS Secrets Manager
aws secretsmanager create-secret \
  --name ssl-private-key \
  --secret-string file://server.key

# HashiCorp Vault
vault kv put secret/ssl private_key=@server.key

# 1Password
op create document server.key --vault Production --title "SSL Private Key"
```

**Option 3: Use encrypted files**:
```bash
# Encrypt with GPG
gpg -c server.key
# Creates server.key.gpg (can be committed)

# Decrypt when needed
gpg -d server.key.gpg > server.key

# Or use SOPS
sops -e server.key > server.key.enc
```

**Option 4: Use certificate management services**:
- Let's Encrypt (free SSL certificates)
- AWS Certificate Manager
- Google Cloud Certificate Manager
- Azure Key Vault

### Best practices

1. **Never commit private keys**:
   ```bash
   # Always add to .gitignore first
   echo "*.key" >> .gitignore
   echo "*.pem" >> .gitignore
   ```

2. **Use separate keys per environment**:
   ```
   dev-server.key (development)
   staging-server.key (staging)
   prod-server.key (production - in secret manager)
   ```

3. **Use short-lived certificates**:
   ```bash
   # Let's Encrypt certificates expire in 90 days
   # Forces regular rotation
   certbot renew
   ```

4. **Set proper permissions**:
   ```bash
   # Private keys should be readable only by owner
   chmod 600 *.key
   chmod 600 *.pem

   # Certificates can be more permissive
   chmod 644 *.crt
   ```

5. **Use pre-commit hooks**:
   ```bash
   #!/bin/bash
   # .git/hooks/pre-commit

   if git diff --cached --name-only | grep -qE '\.(pem|key|crt|p12|pfx)$'; then
     echo "Error: Attempting to commit crypto keys!"
     echo "These files should not be in version control."
     exit 1
   fi
   ```

6. **Scan for secrets**:
   ```bash
   # Use gitleaks or trufflehog
   gitleaks detect --source . --verbose

   # Or git-secrets
   git secrets --scan
   ```

7. **Document key management**:
   ```markdown
   # Key Management

   ## SSL Certificates
   - Stored in AWS Certificate Manager
   - Auto-renewed via Let's Encrypt
   - Access via IAM role

   ## API Keys
   - Stored in 1Password vault "Production"
   - Rotated every 90 days
   ```

### Platform-specific considerations

**Docker**:
```dockerfile
# Don't copy keys into image
# Use secrets or volumes instead

# docker-compose.yml
services:
  app:
    secrets:
      - ssl_key
      - ssl_cert

secrets:
  ssl_key:
    file: ./server.key
  ssl_cert:
    file: ./server.crt
```

**Kubernetes**:
```yaml
# Use Kubernetes secrets
apiVersion: v1
kind: Secret
metadata:
  name: ssl-secret
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
```

**CI/CD**:
```yaml
# GitHub Actions - use secrets
- name: Setup SSL
  env:
    SSL_KEY: ${{ secrets.SSL_PRIVATE_KEY }}
    SSL_CERT: ${{ secrets.SSL_CERTIFICATE }}
  run: |
    echo "$SSL_KEY" > server.key
    echo "$SSL_CERT" > server.crt
```

### Audit crypto files

**Find all crypto files in project**:
```bash
# Find all crypto files
find . -type f \( \
  -name "*.pem" -o \
  -name "*.key" -o \
  -name "*.crt" -o \
  -name "*.p12" -o \
  -name "*.pfx" \
\) ! -path "./.git/*"

# Check if they're in .gitignore
git check-ignore *.pem *.key *.crt
```

**Verify permissions**:
```bash
# Private keys should be 600
find . -name "*.key" -o -name "*.pem" | xargs ls -la

# Fix if needed
find . -name "*.key" -o -name "*.pem" | xargs chmod 600
```

### If keys are compromised

1. **Immediately revoke certificates**
2. **Generate new keys and certificates**
3. **Update all services** with new keys
4. **Rotate any other credentials** that may have been exposed
5. **Monitor for unauthorized use**
6. **Investigate** how keys were exposed
7. **Update procedures** to prevent recurrence


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_UNTRACKED_CRYPTO_KEYS=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
