# Terraform State Local

## What this is

This signal detects when Terraform state files (`terraform.tfstate`) are stored locally instead of in a remote backend. Local state files contain sensitive information and should not be stored on local machines or committed to version control.

## Why this matters

**Security Risk**:
- **Credential exposure**: State files contain plaintext secrets, API keys, passwords, and connection strings
- **Infrastructure details**: State reveals your entire infrastructure topology and configuration
- **Accidental commits**: Local state files often get committed to Git, exposing secrets
- **No encryption**: Local state is stored unencrypted on disk

**Collaboration Issues**:
- **State conflicts**: Multiple team members can't work on the same infrastructure
- **No locking**: Concurrent applies can corrupt state
- **No versioning**: Can't roll back to previous state
- **No audit trail**: Can't track who made changes

**Example of exposed data in state**:
```json
{
  "resources": [{
    "instances": [{
      "attributes": {
        "password": "super-secret-password",
        "connection_string": "postgresql://user:pass@host/db",
        "private_key": "-----BEGIN RSA PRIVATE KEY-----..."
      }
    }]
  }]
}
```

## How to remediate

### Migrate to remote backend

**Option 1: AWS S3 + DynamoDB** (recommended for AWS):

**Create S3 bucket and DynamoDB table**:
```bash
# Create S3 bucket for state
aws s3 mb s3://my-terraform-state --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket my-terraform-state \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket my-terraform-state \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Create DynamoDB table for locking
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

**Configure backend in Terraform**:
```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "project/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}
```

**Migrate existing state**:
```bash
# Initialize backend
terraform init

# Terraform will detect local state and ask to migrate
# Answer "yes" to copy state to remote backend

# Verify migration
terraform state list

# Remove local state files
rm terraform.tfstate terraform.tfstate.backup
```

**Option 2: Terraform Cloud** (easiest):

**Create Terraform Cloud account**:
1. Go to https://app.terraform.io
2. Create organization
3. Create workspace

**Configure backend**:
```hcl
# backend.tf
terraform {
  cloud {
    organization = "my-org"
    
    workspaces {
      name = "my-project"
    }
  }
}
```

**Login and migrate**:
```bash
# Login to Terraform Cloud
terraform login

# Initialize and migrate
terraform init

# Remove local state
rm terraform.tfstate terraform.tfstate.backup
```

**Option 3: Azure Storage**:

**Create storage account**:
```bash
# Create resource group
az group create --name terraform-state-rg --location eastus

# Create storage account
az storage account create \
  --name mytfstate \
  --resource-group terraform-state-rg \
  --location eastus \
  --sku Standard_LRS \
  --encryption-services blob

# Create container
az storage container create \
  --name tfstate \
  --account-name mytfstate
```

**Configure backend**:
```hcl
# backend.tf
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "mytfstate"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}
```

**Option 4: Google Cloud Storage**:

**Create GCS bucket**:
```bash
# Create bucket
gsutil mb -l us-central1 gs://my-terraform-state

# Enable versioning
gsutil versioning set on gs://my-terraform-state
```

**Configure backend**:
```hcl
# backend.tf
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "terraform/state"
  }
}
```

### Add state files to .gitignore

**Ensure state files are ignored**:
```bash
# Add to .gitignore
cat >> .gitignore <<EOF
# Terraform
*.tfstate
*.tfstate.*
.terraform/
.terraform.lock.hcl
terraform.tfvars
*.auto.tfvars
EOF

# Verify
git status
# Should not show any .tfstate files
```

### Remove state from Git if committed

**Check if state is tracked**:
```bash
git ls-files | grep tfstate
```

**Remove from Git** (keep local file):
```bash
# Remove from Git
git rm --cached terraform.tfstate terraform.tfstate.backup

# Commit removal
git add .gitignore
git commit -m "Remove Terraform state from version control"
```

**Purge from Git history** (if contains secrets):
```bash
# WARNING: Rewrites history - coordinate with team!

# Using git-filter-repo
git filter-repo --path terraform.tfstate --invert-paths
git filter-repo --path 'terraform.tfstate.backup' --invert-paths

# Force push
git push origin --force --all
```

**Rotate exposed secrets**:
If state was committed, assume all secrets are compromised:
1. Rotate all credentials in the state file
2. Update Terraform configuration with new credentials
3. Apply changes
4. Monitor for unauthorized access

### Best practices

1. **Always use remote backend**:
   ```hcl
   # Never use local backend in production
   terraform {
     backend "s3" { ... }  # ✓ Good
     # backend "local" { ... }  # ✗ Bad
   }
   ```

2. **Enable state encryption**:
   ```hcl
   # S3 backend with encryption
   terraform {
     backend "s3" {
       encrypt = true  # Always enable
       # ...
     }
   }
   ```

3. **Enable state locking**:
   ```hcl
   # Prevents concurrent modifications
   terraform {
     backend "s3" {
       dynamodb_table = "terraform-state-lock"
       # ...
     }
   }
   ```

4. **Enable versioning**:
   ```bash
   # S3 versioning for state recovery
   aws s3api put-bucket-versioning \
     --bucket my-terraform-state \
     --versioning-configuration Status=Enabled
   ```

5. **Restrict access to state**:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": {
         "AWS": "arn:aws:iam::123456789:role/TerraformRole"
       },
       "Action": "s3:*",
       "Resource": [
         "arn:aws:s3:::my-terraform-state",
         "arn:aws:s3:::my-terraform-state/*"
       ]
     }]
   }
   ```

6. **Use separate state per environment**:
   ```hcl
   # Development
   terraform {
     backend "s3" {
       key = "dev/terraform.tfstate"
     }
   }
   
   # Production
   terraform {
     backend "s3" {
       key = "prod/terraform.tfstate"
     }
   }
   ```

7. **Never commit state files**:
   ```bash
   # Add pre-commit hook
   #!/bin/bash
   if git diff --cached --name-only | grep -q "\.tfstate"; then
     echo "Error: Attempting to commit Terraform state!"
     exit 1
   fi
   ```

### State file security

**Encrypt state at rest**:
- S3: Enable default encryption
- Azure: Enable storage encryption
- GCS: Enable encryption
- Terraform Cloud: Encrypted by default

**Encrypt state in transit**:
- Always use HTTPS/TLS
- Use VPN for additional security

**Audit state access**:
```bash
# S3 access logging
aws s3api put-bucket-logging \
  --bucket my-terraform-state \
  --bucket-logging-status '{
    "LoggingEnabled": {
      "TargetBucket": "my-logs-bucket",
      "TargetPrefix": "terraform-state-access/"
    }
  }'
```

### Backup and recovery

**Backup state regularly**:
```bash
# Download current state
terraform state pull > backup-$(date +%Y%m%d).tfstate

# Store securely (encrypted)
gpg -c backup-$(date +%Y%m%d).tfstate
```

**Restore from backup**:
```bash
# Push state to backend
terraform state push backup.tfstate
```

### Migration checklist

- [ ] Choose remote backend (S3, Terraform Cloud, etc.)
- [ ] Create backend resources (bucket, table, etc.)
- [ ] Configure backend in Terraform
- [ ] Run `terraform init` to migrate
- [ ] Verify state migrated successfully
- [ ] Remove local state files
- [ ] Add state files to .gitignore
- [ ] Remove state from Git history if committed
- [ ] Rotate any exposed secrets
- [ ] Test Terraform operations
- [ ] Document backend configuration for team

