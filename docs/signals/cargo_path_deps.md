# Cargo Path Dependencies

## What this is

This signal detects path dependencies in Rust's `Cargo.toml` file. Path dependencies are local filesystem references to crates (e.g., `path = "../my-local-crate"`) instead of published crates from crates.io.

While path dependencies are useful during local development, they break builds on other machines or in CI/CD pipelines because the referenced paths don't exist elsewhere.

## Why this matters

**Build Failures**: When you commit code with path dependencies:
- **CI/CD pipelines fail** because the local paths don't exist in the build environment
- **Team members can't build** your code on their machines
- **Deployment breaks** because production environments don't have access to your local filesystem

**Security & Supply Chain**: Path dependencies can also indicate:
- **Uncommitted changes** to dependencies that aren't tracked in version control
- **Inconsistent builds** across environments, making it hard to reproduce security issues
- **Dependency confusion** attacks if the path dependency shadows a legitimate crate

**Operational Impact**: This is similar to Go's `replace` directives - they're debugging tools that should never make it to production code.

## How to remediate

### For published crates

1. **Replace path dependencies with crates.io versions**:
   ```toml
   # Before (breaks builds)
   [dependencies]
   my-lib = { path = "../my-lib" }
   
   # After (works everywhere)
   [dependencies]
   my-lib = "1.0.0"
   ```

2. **Update Cargo.lock**:
   ```bash
   cargo update
   ```

3. **Test the build**:
   ```bash
   cargo build
   cargo test
   ```

### For unpublished internal crates

If you're working with internal crates that aren't published to crates.io:

1. **Use a Git dependency** instead:
   ```toml
   [dependencies]
   my-lib = { git = "https://github.com/myorg/my-lib", tag = "v1.0.0" }
   ```

2. **Or publish to a private registry**:
   ```toml
   [dependencies]
   my-lib = { version = "1.0.0", registry = "my-private-registry" }
   ```

### During development

If you need path dependencies for active development:

1. **Use a local override** in `.cargo/config.toml` (not committed):
   ```toml
   [patch.crates-io]
   my-lib = { path = "../my-lib" }
   ```

2. **Or use cargo's `--path` flag** when testing:
   ```bash
   cargo build --manifest-path ../my-lib/Cargo.toml
   ```

### Before committing

1. **Check for path dependencies**:
   ```bash
   grep -n "path.*=" Cargo.toml
   ```

2. **Run a clean build** to verify:
   ```bash
   cargo clean
   cargo build
   ```

