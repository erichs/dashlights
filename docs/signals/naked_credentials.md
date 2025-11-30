# Naked Credentials

## What this is

This signal detects raw secrets and credentials stored in environment variables without encryption or secret management. It looks for environment variables with names containing patterns like `SECRET`, `TOKEN`, `KEY`, `PASSWORD`, or `APIKEY` that contain plaintext values instead of references to secret management tools.

The signal specifically excludes:
- 1Password references (`op://...`)
- dotenvx encrypted values (`encrypted:...`)
- DASHLIGHT_ variables (used by this tool)

## Why this matters

**Security Exposure**:
- **Process listings**: Anyone with access to the system can see environment variables via `ps auxe` or `/proc/*/environ`
- **Log files**: Environment variables often get logged by applications, CI/CD systems, and monitoring tools
- **Error messages**: Stack traces and error reports may include environment variables
- **Child processes**: All child processes inherit environment variables, expanding the attack surface
- **Memory dumps**: Secrets in environment variables appear in core dumps and memory snapshots

**Compliance**: Many security frameworks (SOC 2, PCI-DSS, HIPAA) require secrets to be encrypted at rest and in transit, not stored in plaintext.

**Attack scenarios**:
```bash
# Attacker with limited access can see your secrets
ps auxe | grep AWS_SECRET_ACCESS_KEY

# Or read from /proc
cat /proc/$(pgrep myapp)/environ | tr '\0' '\n' | grep SECRET
```

## How to remediate

### Option 1: Use 1Password CLI (Recommended)

**Install 1Password CLI**:
```bash
# macOS
brew install --cask 1password-cli

# Linux
curl -sS https://downloads.1password.com/linux/keys/1password.asc | \
  sudo gpg --dearmor --output /usr/share/keyrings/1password-archive-keyring.gpg
```

**Store secrets in 1Password**:
```bash
# Create a secret in 1Password
op item create --category=password --title="API Key" \
  --vault="Development" password="your-secret-key"
```

**Reference secrets in your code**:
```bash
# Instead of:
export API_KEY="plaintext-secret"

# Use:
export API_KEY="op://Development/API Key/password"

# Then run your app with op:
op run -- ./myapp
```

### Option 2: Use dotenvx (Encrypted .env)

**Install dotenvx**:
```bash
npm install -g @dotenvx/dotenvx
```

**Encrypt your environment variables**:
```bash
# Create .env file
cat > .env <<EOF
API_KEY=your-secret-key
DB_PASSWORD=your-db-password
EOF

# Encrypt it
dotenvx encrypt

# Now your .env contains encrypted values like:
# API_KEY=encrypted:...
```

**Use encrypted values**:
```bash
# Run your app with dotenvx
dotenvx run -- ./myapp
```

### Option 3: Use cloud secret managers

**AWS Secrets Manager**:
```bash
# Store secret
aws secretsmanager create-secret \
  --name MyAppSecret \
  --secret-string "my-secret-value"

# Retrieve in your application code (not as env var)
# Use AWS SDK to fetch secrets at runtime
```

**Google Cloud Secret Manager**:
```bash
# Store secret
echo -n "my-secret-value" | \
  gcloud secrets create my-secret --data-file=-

# Retrieve in your application code
# Use Google Cloud SDK to fetch secrets at runtime
```

**HashiCorp Vault**:
```bash
# Store secret
vault kv put secret/myapp/config api_key="my-secret-value"

# Retrieve in your application code
# Use Vault SDK to fetch secrets at runtime
```

### Option 4: Use environment-specific secret files

**Create encrypted secret files**:
```bash
# Use SOPS (Secrets OPerationS)
brew install sops

# Create a secrets file
cat > secrets.yaml <<EOF
api_key: my-secret-value
db_password: my-db-password
EOF

# Encrypt it
sops -e secrets.yaml > secrets.enc.yaml

# Decrypt and load at runtime
sops -d secrets.enc.yaml | yq eval '.api_key' -
```

### Clean up existing naked credentials

**Find all secret environment variables**:
```bash
# List all environment variables with secret-like names
env | grep -iE '(SECRET|TOKEN|KEY|PASSWORD|APIKEY)' | cut -d= -f1
```

**Unset them**:
```bash
# Unset specific variables
unset AWS_SECRET_ACCESS_KEY
unset API_KEY
unset DB_PASSWORD

# Or unset all matching a pattern
env | grep -iE '(SECRET|TOKEN|KEY|PASSWORD)' | cut -d= -f1 | \
  while read var; do unset $var; done
```

**Remove from shell configuration**:
```bash
# Find where they're set
grep -r "export.*SECRET\|export.*TOKEN\|export.*KEY\|export.*PASSWORD" \
  ~/.bashrc ~/.bash_profile ~/.zshrc ~/.zprofile

# Edit and remove those lines
nano ~/.bashrc  # or ~/.zshrc
```

**Rotate the exposed secrets**:
1. **Assume all secrets are compromised**
2. **Generate new secrets** for all exposed credentials
3. **Update all services** with new credentials
4. **Revoke old credentials**
5. **Monitor for unauthorized access**

### Best practices

1. **Never set secrets as environment variables directly**:
   ```bash
   # Bad
   export API_KEY="sk-1234567890abcdef"
   
   # Good
   export API_KEY="op://vault/item/field"
   ```

2. **Use secret references** instead of values:
   ```bash
   # Reference to secret manager
   export DB_PASSWORD="op://Production/Database/password"
   
   # Or use a secret file
   export DB_PASSWORD_FILE="/run/secrets/db_password"
   ```

3. **Load secrets at runtime** in your application:
   ```python
   # Instead of os.getenv("API_KEY")
   # Use a secret manager SDK
   import boto3
   client = boto3.client('secretsmanager')
   secret = client.get_secret_value(SecretId='MyAppSecret')
   ```

4. **Use different secrets** for different environments:
   ```bash
   # Development
   export API_KEY="op://Development/API/key"
   
   # Production
   export API_KEY="op://Production/API/key"
   ```

5. **Audit environment variables** regularly:
   ```bash
   # Add to your security checklist
   env | grep -iE '(SECRET|TOKEN|KEY|PASSWORD)' | \
     grep -v "op://" | grep -v "encrypted:"
   ```

6. **Use short-lived credentials** when possible:
   ```bash
   # AWS temporary credentials
   aws sts get-session-token
   
   # Or use IAM roles instead of access keys
   ```

7. **Implement secret rotation**:
   - Rotate secrets regularly (every 90 days minimum)
   - Use automated rotation when available
   - Monitor for failed authentication attempts

