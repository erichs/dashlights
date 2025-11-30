# Dangerous Terraform Variables

## What this is

This signal detects Terraform variables containing secrets that are passed via `TF_VAR_` environment variables. When you set sensitive values like passwords, API keys, or access keys using `TF_VAR_*` environment variables, these secrets often end up in your shell history, log files, and process listings.

The signal looks for `TF_VAR_` variables with names containing patterns like `password`, `secret_key`, `api_key`, `token`, `private_key`, or `credential`.

## Why this matters

**Security Exposure**: Secrets in `TF_VAR_` environment variables create multiple exposure points:
- **Shell history**: Commands like `export TF_VAR_db_password="secret123"` are saved in `~/.bash_history` or `~/.zsh_history`
- **Process listings**: Anyone with access to the system can see environment variables via `ps auxe` or `/proc/*/environ`
- **Log files**: CI/CD systems and shell sessions often log commands, capturing your secrets
- **Accident commits**: Shell history files can accidentally get committed to repositories

**Compliance**: Many security frameworks (SOC 2, PCI-DSS, HIPAA) require secrets to be stored securely, not in plaintext environment variables or shell history.

**Incident Response**: If your shell history contains secrets, you must treat it as compromised data during security incidents, complicating remediation.

## How to remediate

### Option 1: Use .tfvars files (Recommended)

1. **Create a `.tfvars` file** for sensitive values:
   ```bash
   # Create terraform.tfvars (add to .gitignore!)
   cat > terraform.tfvars <<EOF
   db_password = "your-secret-password"
   api_key = "your-api-key"
   EOF
   ```

2. **Add to .gitignore**:
   ```bash
   echo "terraform.tfvars" >> .gitignore
   echo "*.tfvars" >> .gitignore
   ```

3. **Use in Terraform**:
   ```bash
   terraform apply
   # Terraform automatically loads terraform.tfvars
   ```

### Option 2: Use a secret management tool

**For 1Password**:
```bash
# Reference secrets from 1Password
terraform apply -var="db_password=$(op read op://vault/item/password)"
```

**For AWS Secrets Manager**:
```hcl
# In your Terraform code
data "aws_secretsmanager_secret_version" "db_password" {
  secret_id = "prod/db/password"
}

resource "aws_db_instance" "main" {
  password = data.aws_secretsmanager_secret_version.db_password.secret_string
}
```

**For HashiCorp Vault**:
```hcl
data "vault_generic_secret" "db_password" {
  path = "secret/database"
}

resource "aws_db_instance" "main" {
  password = data.vault_generic_secret.db_password.data["password"]
}
```

### Option 3: Use encrypted .tfvars files

**With SOPS (Secrets OPerationS)**:
```bash
# Encrypt your tfvars file
sops -e terraform.tfvars > terraform.tfvars.enc

# Decrypt and apply
sops -d terraform.tfvars.enc | terraform apply -var-file=/dev/stdin
```

### Clean up existing exposure

1. **Clear shell history** of the secret:
   ```bash
   # Edit history file and remove lines with secrets
   nano ~/.bash_history  # or ~/.zsh_history
   
   # Or clear entire history (nuclear option)
   history -c
   rm ~/.bash_history
   ```

2. **Rotate the exposed secrets** immediately

3. **Unset the environment variables**:
   ```bash
   unset TF_VAR_db_password
   unset TF_VAR_api_key
   # etc.
   ```

### Best practices

- **Never use `TF_VAR_` for secrets** - use `.tfvars` files or secret management tools
- **Mark variables as sensitive** in Terraform:
  ```hcl
  variable "db_password" {
    type      = string
    sensitive = true
  }
  ```
- **Use different `.tfvars` files** for different environments (dev.tfvars, prod.tfvars)
- **Store `.tfvars` files securely** - never commit them to version control

