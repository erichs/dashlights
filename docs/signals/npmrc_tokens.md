# NPM RC Tokens in Project Root

## What this is

This signal detects `.npmrc` files in the project root directory that contain authentication tokens. The `.npmrc` file is used by npm to configure package registry settings, and when it contains auth tokens, it should be in your home directory (`~/.npmrc`), not in the project root where it might be committed to version control.

## Why this matters

**Security Risk**:
- **Accidental commits**: `.npmrc` files in project roots often get committed to Git, exposing tokens to anyone with repository access
- **Token theft**: NPM tokens provide full access to publish packages and access private registries
- **Supply chain attacks**: Stolen tokens can be used to publish malicious versions of your packages
- **Public exposure**: If the repository is public or becomes public, tokens are exposed to the internet

**Real-world impact**:
- **Package hijacking**: Attackers can publish malicious versions of your packages
- **Private package access**: Stolen tokens grant access to your organization's private packages
- **Account takeover**: Tokens may provide access to npm account settings
- **Billing fraud**: Attackers can use your npm organization for their own packages

## How to remediate

### Move .npmrc to home directory

**Check what's in project .npmrc**:
```bash
cat .npmrc
```

**Move auth tokens to ~/.npmrc**:
```bash
# Copy auth-related lines to ~/.npmrc
grep -E '(_auth|_authToken|registry.*:_authToken)' .npmrc >> ~/.npmrc

# Remove duplicates
sort -u ~/.npmrc -o ~/.npmrc

# Set proper permissions
chmod 600 ~/.npmrc
```

**Remove auth tokens from project .npmrc**:
```bash
# Remove auth lines from project .npmrc
sed -i.bak '/_auth/d' .npmrc
sed -i.bak '/_authToken/d' .npmrc

# Or delete the file entirely if it only contained auth
rm .npmrc
```

**Add .npmrc to .gitignore**:
```bash
echo ".npmrc" >> .gitignore
git add .gitignore
git commit -m "Add .npmrc to .gitignore"
```

### Remove from Git if already committed

**Check if .npmrc is tracked**:
```bash
git ls-files | grep .npmrc
```

**Remove from Git** (keep local file):
```bash
# Remove from Git but keep local file
git rm --cached .npmrc

# Commit the removal
git add .gitignore
git commit -m "Remove .npmrc from version control"
```

**Purge from Git history** (if it contains tokens):
```bash
# WARNING: This rewrites history - coordinate with team!

# Using git-filter-repo (recommended)
git filter-repo --path .npmrc --invert-paths

# Or using BFG Repo-Cleaner
bfg --delete-files .npmrc
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Force push (WARNING: destructive)
git push origin --force --all
```

### Rotate exposed tokens

**If .npmrc with tokens was committed**:

1. **Revoke the exposed token**:
   ```bash
   # Via npm CLI
   npm token revoke <token-id>

   # Or via npm website
   # Go to https://www.npmjs.com/settings/tokens
   ```

2. **Create a new token**:
   ```bash
   # Create a new token
   npm token create --read-only  # or --publish for publishing

   # Or via npm website
   # Go to https://www.npmjs.com/settings/tokens
   ```

3. **Update ~/.npmrc** with new token:
   ```bash
   # Edit ~/.npmrc
   nano ~/.npmrc

   # Replace old token with new one
   //registry.npmjs.org/:_authToken=npm_newTokenHere
   ```

### Configure .npmrc correctly

**User-level ~/.npmrc** (for auth tokens):
```ini
# ~/.npmrc
//registry.npmjs.org/:_authToken=npm_yourTokenHere

# For scoped packages
@myorg:registry=https://registry.npmjs.org/
//registry.npmjs.org/:_authToken=npm_yourTokenHere

# Set proper permissions
chmod 600 ~/.npmrc
```

**Project-level .npmrc** (for non-sensitive config only):
```ini
# .npmrc (in project root - safe to commit)
# NO AUTH TOKENS HERE!

# Package configuration
save-exact=true
package-lock=true
engine-strict=true

# Registry configuration (without auth)
@myorg:registry=https://registry.npmjs.org/

# This file is safe to commit
```

### Alternative: Use environment variables

**Instead of .npmrc tokens**:
```bash
# Set token as environment variable
export NPM_TOKEN="npm_yourTokenHere"

# Configure .npmrc to use environment variable
echo '//registry.npmjs.org/:_authToken=${NPM_TOKEN}' > .npmrc

# Now .npmrc can be committed (it references env var, not the token)
```

**In CI/CD**:
```yaml
# GitHub Actions
- name: Setup Node
  uses: actions/setup-node@v3
  with:
    node-version: '18'
    registry-url: 'https://registry.npmjs.org'
- name: Install dependencies
  run: npm ci
  env:
    NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Platform-specific considerations

**macOS/Linux**:
```bash
# Check permissions on ~/.npmrc
ls -la ~/.npmrc
# Should be: -rw------- (600)

# Fix if needed
chmod 600 ~/.npmrc
```

**Windows**:
```powershell
# Check if .npmrc exists in project
Test-Path .npmrc

# Move auth to user .npmrc
$userNpmrc = "$env:USERPROFILE\.npmrc"
Get-Content .npmrc | Select-String "_auth" | Add-Content $userNpmrc

# Remove from project
Remove-Item .npmrc
```

### Best practices

1. **Never commit .npmrc with tokens**:
   ```bash
   # Always add to .gitignore first
   echo ".npmrc" >> .gitignore
   ```

2. **Use scoped tokens** with minimal permissions:
   ```bash
   # Create read-only token for CI
   npm token create --read-only --cidr=0.0.0.0/0

   # Create publish token with IP restrictions
   npm token create --publish --cidr=your.ci.ip.address/32
   ```

3. **Audit npm tokens** regularly:
   ```bash
   # List all tokens
   npm token list

   # Revoke unused tokens
   npm token revoke <token-id>
   ```

4. **Use .npmrc.example** for documentation:
   ```bash
   # Create a template without tokens
   cat > .npmrc.example <<EOF
   # NPM Configuration
   # Copy to .npmrc and add your auth token

   save-exact=true
   package-lock=true

   # Add your token to ~/.npmrc instead:
   # //registry.npmjs.org/:_authToken=npm_yourTokenHere
   EOF

   git add .npmrc.example
   ```

5. **Use npm organizations** for better access control:
   - Create organization-specific tokens
   - Use team-based permissions
   - Enable 2FA for all organization members

6. **Monitor for token exposure**:
   ```bash
   # Use tools like gitleaks or trufflehog
   gitleaks detect --source . --verbose
   ```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_NPMRC_TOKENS=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
