# Production Panic Mode

## What this is

This signal detects when environment variables indicate you're working in a production environment. It looks for variables like `ENV=production`, `ENVIRONMENT=prod`, `NODE_ENV=production`, or similar patterns that suggest you're connected to or working with production systems.

This is a warning to be extra careful, as mistakes in production can have serious consequences.

## Why this matters

**Operational Risk**:
- **Data loss**: Accidental deletions or modifications affect real user data
- **Service outages**: Mistakes can take down production services
- **Financial impact**: Downtime costs money and damages reputation
- **Customer impact**: Errors affect real users and customers

**Security Risk**:
- **Credential exposure**: Production credentials are more valuable to attackers
- **Compliance violations**: Mistakes with production data can violate regulations (GDPR, HIPAA, PCI-DSS)
- **Audit trail**: Production changes should be tracked and approved

**Common mistakes in production**:
```bash
# Accidentally dropping production database
DROP DATABASE users;  # Oops, thought I was in dev!

# Deleting production data
rm -rf /var/lib/production-data/*

# Deploying untested code
git push production main  # Without testing first

# Exposing secrets
echo $DATABASE_PASSWORD  # In a screen share or log
```

## How to remediate

### Verify you need production access

**Ask yourself**:
1. Do I really need to be in production right now?
2. Can this be done in staging/development instead?
3. Is there a safer way to accomplish this task?
4. Have I tested this in a non-production environment first?

**If you don't need production access**:
```bash
# Switch to development environment
unset ENV
unset ENVIRONMENT
unset NODE_ENV
export NODE_ENV=development

# Or use a different terminal/session for production
```

### Use separate terminals for production

**Color-code your terminals**:

**iTerm2 (macOS)**:
```bash
# Add to ~/.zshrc or ~/.bashrc
if [ "$NODE_ENV" = "production" ] || [ "$ENV" = "production" ]; then
  echo -e "\033]6;1;bg;red;brightness;255\a"
  echo -e "\033]6;1;bg;green;brightness;0\a"
  echo -e "\033]6;1;bg;blue;brightness;0\a"
  export PS1="%F{red}[PRODUCTION]%f $PS1"
fi
```

**Gnome Terminal (Linux)**:
```bash
# Add to ~/.bashrc
if [ "$NODE_ENV" = "production" ] || [ "$ENV" = "production" ]; then
  export PS1="\[\e[41m\][PRODUCTION]\[\e[0m\] $PS1"
fi
```

**Windows Terminal**:
```json
// In settings.json
{
  "profiles": {
    "list": [
      {
        "name": "Production",
        "colorScheme": "Red Alert",
        "background": "#330000"
      }
    ]
  }
}
```

### Add safety checks to scripts

**Require confirmation for production**:
```bash
#!/bin/bash

if [ "$ENV" = "production" ]; then
  echo "WARNING: You are in PRODUCTION environment!"
  echo "This script will modify production data."
  read -p "Are you sure you want to continue? (type 'yes' to confirm): " confirm
  if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
  fi
fi

# Rest of script...
```

**Add dry-run mode**:
```bash
#!/bin/bash

DRY_RUN=${DRY_RUN:-false}

if [ "$ENV" = "production" ] && [ "$DRY_RUN" != "true" ]; then
  echo "ERROR: Must use DRY_RUN=true in production first"
  exit 1
fi

if [ "$DRY_RUN" = "true" ]; then
  echo "[DRY RUN] Would delete files..."
else
  rm -rf /data/*
fi
```

### Use production access controls

**Require MFA for production**:
```bash
# AWS example
aws configure set mfa_serial arn:aws:iam::123456789:mfa/user

# Require MFA token for production access
aws sts get-session-token --serial-number arn:aws:iam::123456789:mfa/user --token-code 123456
```

**Use bastion hosts**:
```bash
# Access production only through bastion
ssh -J bastion.example.com production.example.com

# Not directly
# ssh production.example.com  # Blocked by firewall
```

**Use time-limited credentials**:
```bash
# Request temporary production access (expires in 1 hour)
request-prod-access --duration 1h --reason "Investigating incident #1234"
```

### Best practices for production work

1. **Always test in staging first**:
   ```bash
   # Test in staging
   export ENV=staging
   ./deploy.sh

   # Verify it works
   # Then deploy to production
   export ENV=production
   ./deploy.sh
   ```

2. **Use read-only access by default**:
   ```bash
   # Request read-only production access
   export PROD_MODE=readonly

   # Request write access only when needed
   export PROD_MODE=readwrite
   ```

3. **Pair program for production changes**:
   - Have someone review your commands before executing
   - Use screen sharing for visibility
   - Document what you're doing

4. **Use change management**:
   - Create a change ticket
   - Get approval before making changes
   - Document rollback plan
   - Schedule maintenance windows

5. **Enable audit logging**:
   ```bash
   # Log all commands in production
   if [ "$ENV" = "production" ]; then
     export PROMPT_COMMAND='history -a; logger -t production "$(history 1)"'
   fi
   ```

6. **Use infrastructure as code**:
   ```bash
   # Instead of manual changes
   # Use Terraform, Ansible, etc.
   terraform plan
   terraform apply
   ```

7. **Set up alerts**:
   ```bash
   # Alert when someone accesses production
   if [ "$ENV" = "production" ]; then
     curl -X POST https://slack.com/api/chat.postMessage \
       -d "text=User $USER accessed production environment"
   fi
   ```

### Create a production safety checklist

**Before making production changes**:
- [ ] Is this change necessary?
- [ ] Have I tested this in staging?
- [ ] Do I have a rollback plan?
- [ ] Have I notified the team?
- [ ] Is this during a maintenance window?
- [ ] Do I have a backup?
- [ ] Have I reviewed the commands?
- [ ] Is someone else available to help if needed?

### Alternative: Use separate accounts

**Instead of environment variables**:
```bash
# Use separate AWS accounts
aws configure --profile dev
aws configure --profile prod

# Use separate Kubernetes contexts
kubectl config use-context dev
kubectl config use-context prod

# Use separate database connections
psql -h dev-db.example.com
psql -h prod-db.example.com
```

### If you make a mistake in production

1. **Don't panic** - panicking makes it worse
2. **Assess the impact** - what broke?
3. **Notify the team** immediately
4. **Execute rollback plan** if you have one
5. **Document what happened** for post-mortem
6. **Learn from it** - update procedures to prevent recurrence


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_PROD_PANIC=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
