# Missing Git Hooks

## What this is

This signal detects when a repository has git hook configuration (from tools like Husky, pre-commit, Lefthook, or simple `.githooks/` directories) but the hooks are not actually installed in the effective hooks directory. This typically happens when someone clones a repository but forgets to run the hook installation command.

## Why this matters

**Development Hygiene**:
- **Bypassed quality gates**: Pre-commit hooks often run linters, formatters, and tests that catch issues before they reach CI
- **Inconsistent commits**: Without hooks, developers may commit code that fails CI checks, wasting time
- **Security checks skipped**: Many teams use hooks to scan for secrets or validate commit messages
- **Team friction**: "Works on my machine" issues when hooks run for some developers but not others

**Real-world impact**:
- **CI failures**: Commits that would have been caught locally fail in CI, slowing down the team
- **Secret leaks**: Pre-commit secret scanning hooks won't catch leaked credentials
- **Style drift**: Code formatting hooks won't enforce consistent style
- **Bad commit messages**: Commit-msg hooks won't validate conventional commits

## How to remediate

### Identify your hook manager

Check which hook manager your project uses:

```bash
# Check for common hook manager configurations
ls -la .pre-commit-config.yaml .husky .lefthook.yml lefthook.yml .githooks .git-hooks 2>/dev/null
```

### Install hooks based on your tool

**pre-commit** (Python-based):
```bash
# Install pre-commit if needed
pip install pre-commit

# Install the hooks
pre-commit install
```

**Husky** (npm-based):
```bash
# Hooks are typically installed automatically with npm install
npm install

# Or manually install hooks
npx husky install
```

**Lefthook** (Go-based):
```bash
# Install lefthook if needed
brew install lefthook  # macOS
# or: go install github.com/evilmartians/lefthook@latest

# Install the hooks
lefthook install
```

**Simple `.githooks/` directory**:
```bash
# Configure git to use the .githooks directory
git config core.hooksPath .githooks

# Or copy hooks manually
cp .githooks/* .git/hooks/
chmod +x .git/hooks/*
```

### Verify hooks are installed

```bash
# Check what hooks are installed
ls -la .git/hooks/ | grep -v '.sample'

# Or check custom hooks path
git config core.hooksPath
ls -la $(git config core.hooksPath)

# Test a hook manually
.git/hooks/pre-commit
```

### Automate hook installation

**Add to package.json** (for npm projects):
```json
{
  "scripts": {
    "prepare": "husky install"
  }
}
```

**Add to Makefile**:
```makefile
.PHONY: setup
setup:
	pre-commit install
	# or: lefthook install
	# or: git config core.hooksPath .githooks
```

**Document in README**:
```markdown
## Development Setup

After cloning, install git hooks:

\`\`\`bash
pre-commit install
\`\`\`
```

### Best practices

1. **Use `prepare` script** in package.json to auto-install hooks on `npm install`
2. **Document hook installation** in your README or CONTRIBUTING.md
3. **Consider using `core.hooksPath`** in `.gitconfig` for consistency
4. **Run hooks in CI too** to catch any bypass: `pre-commit run --all-files`
5. **Keep hooks fast** - slow hooks lead to developers bypassing them with `--no-verify`

## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_MISSING_GIT_HOOKS=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).

