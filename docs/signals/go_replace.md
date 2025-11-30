# Go Replace Directive

## What this is

This signal detects `replace` directives in your `go.mod` file. Replace directives are used to override module dependencies with local filesystem paths or alternative versions, typically for debugging or development purposes.

While useful during development, replace directives break builds on other machines because the local paths don't exist elsewhere.

## Why this matters

**Build Failures**:
- **CI/CD pipelines fail** because the local paths referenced in replace directives don't exist in build environments
- **Team members can't build** your code on their machines
- **Deployment breaks** because production environments don't have access to your local filesystem
- **Go modules break** because replace directives override the module resolution system

**Security & Supply Chain**:
- **Dependency confusion**: Replace directives can mask which actual version of a dependency is being used
- **Inconsistent builds**: Different developers might have different versions at the replaced path
- **Audit trail loss**: It's unclear which version of code is actually running
- **Supply chain attacks**: Malicious replace directives could point to compromised code

## How to remediate

### Remove replace directives before committing

**Check for replace directives**:
```bash
grep "^replace " go.mod
```

**Remove them**:
```bash
# Edit go.mod and delete lines starting with "replace"
nano go.mod

# Or use sed
sed -i '/^replace /d' go.mod
```

**Update dependencies**:
```bash
go mod tidy
go mod verify
```

**Test the build**:
```bash
go build ./...
go test ./...
```

### Alternative approaches for development

**Option 1: Use go.work for local development**:
```bash
# Create a workspace file (not committed)
go work init
go work use . ../other-module

# Add go.work to .gitignore
echo "go.work" >> .gitignore
echo "go.work.sum" >> .gitignore
```

**Option 2: Use replace in a local go.mod.local**:
```bash
# Create a local override file
cat > go.mod.local <<EOF
replace github.com/myorg/mymodule => ../mymodule
EOF

# Add to .gitignore
echo "go.mod.local" >> .gitignore

# Use it during development
go build -modfile=go.mod.local
```

**Option 3: Use a Git dependency instead**:
```bash
# Instead of a local path, use a Git reference
# In go.mod:
require github.com/myorg/mymodule v0.0.0-20240101000000-abcdef123456

# Or use a branch
go get github.com/myorg/mymodule@feature-branch
```

### Valid use cases for replace directives

Replace directives are acceptable in these scenarios:

**1. Forked dependencies** (should be committed):
```go
// go.mod
replace github.com/original/module => github.com/yourorg/module v1.2.3
```

**2. Vendored dependencies**:
```go
// go.mod
replace github.com/some/module => ./vendor/github.com/some/module
```

**3. Security patches** for unmaintained modules:
```go
// go.mod
replace github.com/vulnerable/module => github.com/patched/module v1.0.1
```

### Clean up after development

**Before committing**:
```bash
# 1. Remove replace directives
sed -i '/^replace /d' go.mod

# 2. Update dependencies
go mod tidy

# 3. Verify everything still works
go build ./...
go test ./...

# 4. Check what changed
git diff go.mod go.sum

# 5. Commit
git add go.mod go.sum
git commit -m "Remove local replace directives"
```

### Prevent future issues

**Add a pre-commit hook**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

if git diff --cached go.mod | grep "^+replace.*=>.*\.\."; then
  echo "Error: go.mod contains local replace directive"
  echo "Remove replace directives before committing"
  exit 1
fi
```

**Add to CI/CD**:
```yaml
# .github/workflows/ci.yml
- name: Check for replace directives
  run: |
    if grep -q "^replace.*=>.*\.\." go.mod; then
      echo "Error: go.mod contains local replace directive"
      exit 1
    fi
```

### Best practices

1. **Use go.work** for local multi-module development instead of replace directives
2. **Document** any necessary replace directives in README.md
3. **Use version tags** instead of local paths when possible
4. **Test in a clean environment** before committing
5. **Review go.mod changes** carefully in pull requests

