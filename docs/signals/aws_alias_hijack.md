# AWS CLI Alias Hijacking

## What this is

This signal detects potentially malicious AWS CLI aliases that override core AWS commands. The AWS CLI allows users to create custom command aliases in `~/.aws/cli/alias`, but if these aliases override critical AWS commands (like `s3`, `iam`, `sts`, etc.), they could be used for command injection attacks or to intercept sensitive operations.

The signal also checks for insecure file permissions on the alias file itself, as world-readable or world-writable alias files could allow attackers to inject malicious aliases.

## Why this matters

**Security Risk**: If an attacker gains access to modify your AWS CLI alias file, they could:
- Intercept AWS credentials by overriding commands like `configure` or `sts`
- Exfiltrate data by hijacking commands like `s3` or `secretsmanager`
- Execute arbitrary code when you run what you think are legitimate AWS commands
- Bypass audit trails by redirecting commands to malicious scripts

**Attack Vector**: This is a persistence mechanism that attackers use after initial compromise. By hijacking core AWS commands, they can maintain access and steal credentials even after you think you've secured your system.

## How to remediate

### Review and remove suspicious aliases

1. **Check your AWS CLI alias file**:
   ```bash
   cat ~/.aws/cli/alias
   ```

2. **Look for aliases that override core commands** such as:
   - `s3`, `ec2`, `iam`, `sts`, `lambda`, `kms`, `secretsmanager`
   - `configure`, `login`, `sso`
   - Any other core AWS service names

3. **Remove suspicious aliases**:
   ```bash
   # Edit the file and remove any suspicious entries
   nano ~/.aws/cli/alias
   ```

### Fix file permissions

1. **Set correct permissions** (owner read/write only):
   ```bash
   chmod 600 ~/.aws/cli/alias
   ```

2. **Verify permissions**:
   ```bash
   ls -la ~/.aws/cli/alias
   # Should show: -rw------- (600)
   ```

### Best practices

- **Never alias core AWS commands** - use descriptive names that don't conflict with built-in commands
- **Use prefixes** for custom aliases (e.g., `my-deploy` instead of `deploy`)
- **Regularly audit** your alias file for unexpected changes
- **Monitor** the file with integrity checking tools if you're in a high-security environment



## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_AWS_ALIAS_HIJACK=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).