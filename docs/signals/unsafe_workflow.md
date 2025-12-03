# Unsafe Workflow

## What this is

This signal detects dangerous GitHub Actions workflow patterns that could lead to security vulnerabilities:

1. **Pwn Request**: `pull_request_target` trigger combined with explicit checkout of untrusted PR code, allowing malicious PR authors to execute arbitrary code with elevated privileges.

2. **Expression Injection**: Untrusted GitHub context data (like issue titles, PR bodies, commit messages) used directly in `run:` blocks, allowing attackers to inject shell commands.

## Why this matters

**Repository Compromise**:
- **Write access exposure**: `pull_request_target` workflows have write permissions to the target repository
- **Secret theft**: Attackers can exfiltrate repository secrets (API keys, tokens, credentials)
- **Supply chain attacks**: Malicious code can be injected into releases or dependencies
- **Persistent backdoors**: Attackers may add malicious commits or modify CI/CD pipelines

**Expression Injection Attack Vectors**:
- Malicious issue/PR titles containing shell commands
- Crafted commit messages with embedded payloads
- Branch names designed to exploit shell interpolation
- Comment bodies with injection payloads

## How to remediate

### Pwn Request Remediation

#### Option 1: Use `pull_request` trigger instead (recommended)

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

#### Option 2: Use the workflow_run pattern

For scenarios requiring both untrusted code execution AND repository write access, split into two workflows.

#### Option 3: Add persist-credentials: false (partial mitigation)

```yaml
on: pull_request_target

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          persist-credentials: false  # Required!
```

### Expression Injection Remediation

#### Use environment variables as indirection

Instead of interpolating untrusted data directly in `run:` blocks, set it as an environment variable first:

```yaml
# VULNERABLE: Direct interpolation
- run: echo "${{ github.event.issue.title }}"

# SAFE: Use env variable
- name: Echo title safely
  env:
    TITLE: ${{ github.event.issue.title }}
  run: echo "$TITLE"
```

#### Safe patterns

Passing untrusted data to action inputs is safe (not shell-interpolated):

```yaml
# SAFE: Action inputs are not shell-interpolated
- uses: some-action@v1
  with:
    title: ${{ github.event.issue.title }}
```

### Untrusted GitHub Context Variables

The following context variables can be controlled by attackers and should never be used directly in `run:` blocks:

| Context Variable | Attack Vector |
|-----------------|---------------|
| `github.event.issue.title` | Malicious issue title |
| `github.event.issue.body` | Malicious issue body |
| `github.event.pull_request.title` | Malicious PR title |
| `github.event.pull_request.body` | Malicious PR body |
| `github.event.comment.body` | Malicious comment |
| `github.event.review.body` | Malicious review |
| `github.event.head_commit.message` | Malicious commit message |
| `github.event.head_commit.author.email` | Crafted author email |
| `github.event.head_commit.author.name` | Crafted author name |
| `github.event.pull_request.head.ref` | Malicious branch name |
| `github.event.pull_request.head.label` | Malicious label |
| `github.head_ref` | Malicious branch name |
| `github.event.commits` | Malicious commit data |
| `github.event.pages` | Malicious page data |
| `github.event.pull_request.head.repo.default_branch` | Malicious default branch name |

## Detection details

This signal uses a hybrid detection approach for performance:

1. **Fast path**: Check if `.github/workflows/` directory exists
2. **Quick scan**: Line-by-line search for trigger patterns and untrusted expressions
3. **YAML parse**: Full parsing only for files containing potential vulnerabilities

The signal detects:
- `pull_request_target` trigger with `actions/checkout` specifying PR head ref without `persist-credentials: false`
- Untrusted GitHub context expressions used directly in `run:` blocks

## References

- **Pwn Requests**: [Keeping your GitHub Actions and workflows secure Part 1: Preventing pwn requests](https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/) - GitHub Security Lab
- **Expression Injection**: [Keeping your GitHub Actions and workflows secure Part 2: Untrusted input](https://securitylab.github.com/resources/github-actions-untrusted-input/) - GitHub Security Lab
- **Author**: Jaroslav Lobačevski (jarlob)

