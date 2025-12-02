# Pwn Request Risk

## What this is

This signal detects dangerous GitHub Actions workflow patterns that could lead to "pwn request" vulnerabilities. A pwn request occurs when a workflow uses `pull_request_target` trigger combined with an explicit checkout of untrusted PR code, allowing malicious PR authors to execute arbitrary code with elevated privileges.

## Why this matters

**Repository Compromise**:
- **Write access exposure**: `pull_request_target` workflows have write permissions to the target repository
- **Secret theft**: Attackers can exfiltrate repository secrets (API keys, tokens, credentials)
- **Supply chain attacks**: Malicious code can be injected into releases or dependencies
- **Persistent backdoors**: Attackers may add malicious commits or modify CI/CD pipelines

**Attack Vectors**:
- Modified build scripts (`Makefile`, `package.json` scripts)
- Malicious test code that runs during CI
- `npm install` preinstall/postinstall hooks
- Custom GitHub Actions with embedded malicious code
- Any code path that executes during the workflow

**Historical Context**:
The PwnRequest vulnerability pattern was responsible for the first documented case of the November 2025 NPM Shai-Hulud 2.0 supply-chain worm.

## How to remediate

### Option 1: Use `pull_request` trigger instead (recommended)

The safest approach is to use `pull_request` instead of `pull_request_target`:

```yaml
# SAFE: pull_request trigger has limited permissions
on: pull_request

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm install && npm test
```

### Option 2: Use the workflow_run pattern

For scenarios requiring both untrusted code execution AND repository write access:

```yaml
# Workflow 1: ReceivePR.yml (unprivileged)
name: Receive PR
on: pull_request

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm install && npm test
      - name: Save PR number
        run: echo ${{ github.event.number }} > ./pr-number.txt
      - uses: actions/upload-artifact@v4
        with:
          name: pr-info
          path: pr-number.txt
```

```yaml
# Workflow 2: CommentPR.yml (privileged, runs after ReceivePR)
name: Comment on PR
on:
  workflow_run:
    workflows: ["Receive PR"]
    types: [completed]

jobs:
  comment:
    runs-on: ubuntu-latest
    if: github.event.workflow_run.conclusion == 'success'
    steps:
      - name: Download artifact
        uses: actions/download-artifact@v4
      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            // Safe to use secrets here
```

### Option 3: Add persist-credentials: false (partial mitigation)

If you must use `pull_request_target` with PR checkout:

```yaml
# PARTIALLY MITIGATED: Still risky, but limits credential exposure
on: pull_request_target

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: false  # Required!
      # WARNING: Secrets are still accessible to running code
```

**Note**: This only prevents the repository token from being stored on disk. Running code can still access secrets passed via environment variables or other means.

### What NOT to do

```yaml
# VULNERABLE: DO NOT USE THIS PATTERN
on: pull_request_target

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      # Untrusted PR code now runs with write permissions!
      - run: npm install
      - run: npm test
```

## Safe uses of pull_request_target

The `pull_request_target` trigger IS safe when:

1. **No checkout of PR code**: Only labeling, commenting, or other metadata operations
2. **Default checkout only**: Without `ref:` parameter (checks out target branch, not PR)
3. **Treating PR content as data**: Reading files without executing them

```yaml
# SAFE: No checkout, just labeling
on: pull_request_target

jobs:
  label:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/labeler@v4
```

```yaml
# SAFE: Default checkout (target branch, not PR)
on: pull_request_target

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4  # No ref = safe
      - run: echo "Checking target branch"
```

## Detection details

This signal uses a hybrid detection approach for performance:

1. **Fast path**: Check if `.github/workflows/` directory exists
2. **Quick scan**: Line-by-line search for `pull_request_target` string
3. **YAML parse**: Full parsing only for files containing the trigger

The signal detects:
- `pull_request_target` trigger (string, array, or map format)
- Combined with `actions/checkout` step
- That specifies `ref:` containing PR head reference
- Without `persist-credentials: false`

## References

- **Source**: [Keeping your GitHub Actions and workflows secure Part 1: Preventing pwn requests](https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/) - GitHub Security Lab
- **Author**: Jaroslav Lobačevski (jarlob)
- **Published**: August 3, 2021

## Related signals

- Consider also checking for [expression injection vulnerabilities](https://securitylab.github.com/resources/github-actions-untrusted-input/) in `run:` blocks

