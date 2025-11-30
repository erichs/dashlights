# .env File Not Ignored

## What this is

This signal detects when a `.env` file exists in your project directory but is not listed in `.gitignore`. The `.env` file typically contains environment variables including API keys, database passwords, and other secrets that should never be committed to version control.

## Why this matters

**Security Risk**:
- **Credential exposure**: Committing `.env` files exposes all your secrets to anyone with repository access
- **Public leaks**: If the repository is public or becomes public later, your secrets are exposed to the internet
- **Git history**: Even if you remove the file later, it remains in Git history forever
- **Forks and clones**: Secrets propagate to all forks and clones of the repository

**Real-world impact**:
- **API key theft**: Attackers scan GitHub for exposed API keys and use them within minutes
- **Database breaches**: Exposed database credentials lead to data theft
- **AWS bill shock**: Stolen AWS keys are used for cryptocurrency mining
- **Account takeover**: Exposed credentials allow attackers to impersonate your services

**Compliance**: Many regulations (GDPR, PCI-DSS, SOC 2) require secrets to be stored securely, not in version control.

## How to remediate

### Immediate action - Add to .gitignore

**If .gitignore exists**:
```bash
# Add .env to .gitignore
echo ".env" >> .gitignore
echo "*.env" >> .gitignore
echo ".env.*" >> .gitignore
```

**If .gitignore doesn't exist**:
```bash
# Create .gitignore with common patterns
cat > .gitignore <<EOF
# Environment variables
.env
.env.*
*.env
.env.local
.env.*.local

# OS files
.DS_Store
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo
EOF
```

### Remove from Git if already committed

**Check if .env is tracked**:
```bash
git ls-files | grep .env
```

**If it's tracked, remove it from Git** (but keep local copy):
```bash
# Remove from Git but keep local file
git rm --cached .env

# Commit the removal
git add .gitignore
git commit -m "Remove .env from version control and add to .gitignore"
```

**If it's in Git history, you must purge it**:
```bash
# WARNING: This rewrites history - coordinate with team first!

# Using git-filter-repo (recommended)
git filter-repo --path .env --invert-paths

# Or using BFG Repo-Cleaner
bfg --delete-files .env
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Force push (WARNING: destructive)
git push origin --force --all
```

### Rotate exposed secrets

**If .env was committed, assume all secrets are compromised**:

1. **Rotate all API keys and tokens**
2. **Change all passwords**
3. **Revoke and regenerate database credentials**
4. **Update all services with new credentials**
5. **Monitor for unauthorized access**

### Create a .env.example template

**Create a template without secrets**:
```bash
# Create .env.example with placeholder values
cat > .env.example <<EOF
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=your_username_here
DB_PASSWORD=your_password_here

# API Keys
API_KEY=your_api_key_here
SECRET_KEY=your_secret_key_here

# AWS
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
EOF

# Commit the example file
git add .env.example
git commit -m "Add .env.example template"
```

### Use encrypted secrets instead

**Option 1: Use dotenvx (encrypted .env)**:
```bash
# Install dotenvx
npm install -g @dotenvx/dotenvx

# Encrypt your .env file
dotenvx encrypt

# Now you can commit .env.encrypted
git add .env.encrypted
```

**Option 2: Use SOPS**:
```bash
# Install SOPS
brew install sops  # macOS
# or download from https://github.com/mozilla/sops

# Encrypt .env
sops -e .env > .env.encrypted

# Decrypt when needed
sops -d .env.encrypted > .env
```

**Option 3: Use 1Password**:
```bash
# Reference secrets from 1Password instead of .env
# In your code, use op:// references
export API_KEY="op://vault/item/field"
```

### Best practices

1. **Always add .env to .gitignore** before creating it:
   ```bash
   echo ".env" >> .gitignore
   git add .gitignore
   git commit -m "Add .gitignore"
   # Now create .env
   ```

2. **Use different .env files** for different environments:
   ```bash
   .env.development
   .env.staging
   .env.production
   # All should be in .gitignore
   ```

3. **Document required variables** in .env.example

4. **Use a pre-commit hook** to prevent committing .env:
   ```bash
   # .git/hooks/pre-commit
   if git diff --cached --name-only | grep -q "^\.env$"; then
     echo "Error: Attempting to commit .env file!"
     exit 1
   fi
   ```

5. **Scan for secrets** before pushing:
   ```bash
   # Use gitleaks or trufflehog
   gitleaks detect --source . --verbose
   ```

