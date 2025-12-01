# Contributing to Dashlights

Thank you for your interest in contributing to Dashlights! We welcome contributions from the community and are grateful for your help in making this project better.

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors. We expect all participants to:

- Be respectful and considerate in communication
- Accept constructive criticism gracefully
- Focus on what is best for the community and the project
- Show empathy towards other community members

### Contributor Eligibility

By submitting a contribution, you assert that:

- You are legally eligible to make the contribution
- You have the right to grant the license for your contribution
- Your contribution is your original work or you have permission to submit it
- You understand that your contribution will be publicly available under the project's license

If you have any questions about your eligibility to contribute, please reach out to the maintainers before submitting.

## Getting Started

### Prerequisites

- **Go 1.24 or later** - [Install Go](https://golang.org/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Unix systems
- **gosec** (optional) - Installed automatically by `make gosec`

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/dashlights.git
   cd dashlights
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/erichs/dashlights.git
   ```

4. **Install Git hooks** (recommended):
   ```bash
   make hooks
   ```
   This installs pre-commit (formatting check) and pre-push (documentation validation) hooks.

5. **Build and test**:
   ```bash
   # Run all checks (format, build, test, security scan)
   make

   # Or run individual steps
   make build
   make test
   make coverage
   ```

6. **Verify everything works**:
   ```bash
   ./dashlights --details
   ```

## Development Workflow

### Creating a Branch

Always create a new branch for your work:

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### Making Changes

1. **Write your code** following the style guide (see below)
2. **Add tests** for any new functionality
3. **Update documentation** if you're adding or changing features
4. **Run tests locally**:
   ```bash
   make test
   make coverage
   ```
5. **Format your code**:
   ```bash
   make fmt
   ```
6. **Run security scan**:
   ```bash
   make gosec
   ```

### Committing Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add Docker socket detection for macOS"
```

**Good commit messages**:
- Start with a verb in present tense ("Add", "Fix", "Update", "Remove")
- Be specific about what changed
- Reference issue numbers if applicable (e.g., "Fix #123: ...")

### Submitting a Pull Request

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub from your fork to `erichs/dashlights:main`

3. **Fill out the PR template** with:
   - Description of changes
   - Related issue numbers
   - Testing performed
   - Screenshots (if UI changes)

4. **Wait for CI checks** to pass (see Pull Request Process below)

5. **Respond to review feedback** promptly and professionally

## Style Guide

### Go Code Style

- **Follow standard Go conventions**: Use `gofmt` for formatting
- **Run `make fmt`** before committing to auto-format all files
- **CI enforces formatting**: `make fmt-check` must pass
- **Use meaningful names**: Variables, functions, and types should be self-documenting
- **Add comments**: Especially for exported functions and complex logic
- **Keep functions focused**: Each function should do one thing well

### Code Organization

- **Signal implementations**: Place in `src/signals/`
- **Tests**: Co-locate with code (e.g., `docker_socket.go` → `docker_socket_test.go`)
- **Documentation**: Place in `docs/signals/`

## Testing Requirements

### Coverage Standards

- **Project maintains ~90% test coverage** - please help us keep it there!
- **All new code must include tests** - PRs without tests will not be merged
- **Test both success and failure cases** - edge cases matter

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make coverage

# Run tests with coverage for signals package only
make coverage-signals

# Generate HTML coverage report
make coverage-html

# Run integration tests (including performance)
make test-integration

# Run race detector
make test-race
```

### Writing Tests

- **Use table-driven tests** where appropriate
- **Test edge cases**: empty inputs, nil values, boundary conditions
- **Use `t.TempDir()`** for temporary files/directories
- **Mock external dependencies**: file systems, environment variables, etc.
- **Test concurrency**: If your code runs concurrently, test it under contention

Example test structure:
```go
func TestMySignal(t *testing.T) {
    tests := []struct {
        name     string
        setup    func()
        expected bool
    }{
        {"detects issue", func() { /* setup */ }, true},
        {"no issue", func() { /* setup */ }, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            signal := NewMySignal()
            result := signal.Check(context.Background())
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

## Signal Development

### Adding a New Signal

When adding a new security signal, you must:

1. **Implement the Signal interface** in `src/signals/your_signal.go`:
   ```go
   type Signal interface {
       Check(ctx context.Context) bool  // Returns true if issue detected
       Name() string                     // Human-readable name (e.g., "Open Door")
       Emoji() string                    // Emoji for diagnostic output (e.g., "🔑")
       Diagnostic() string               // Brief explanation of the issue
       Remediation() string              // How to fix the issue
   }
   ```

2. **Create comprehensive unit tests** in `src/signals/your_signal_test.go`:
   - Test detection of the issue (true case)
   - Test when no issue exists (false case)
   - Test edge cases and error conditions
   - Aim for >85% coverage for your signal

3. **Create documentation** in `docs/signals/your_signal.md`:
   - Must follow the three-section format (see below)
   - Filename must match signal name (e.g., `DockerSocketSignal` → `docker_socket.md`)
   - Use snake_case for filename (e.g., `ssh_agent_bloat.md`)

4. **Register your signal** in `src/signals/registry.go`:
   ```go
   func GetAllSignals() []Signal {
       return []Signal{
           // ... existing signals ...
           NewYourSignal(),
       }
   }
   ```

### Signal Naming Convention

- **Type name**: `YourSignalSignal` (e.g., `DockerSocketSignal`)
- **Constructor**: `NewYourSignal()` (e.g., `NewDockerSocketSignal()`)
- **Documentation file**: `your_signal.md` (e.g., `docker_socket.md`)

### Documentation Format

All signal documentation must follow this three-section format:

#### 1. What this is

Explain what the signal detects and why it matters:

```markdown
## What this is

This signal detects [specific security issue].

[Explanation of the issue]

## Why this matters

**Security Risk**:
- [Risk 1]
- [Risk 2]

**Example attack**:
```bash
# Show concrete example of the vulnerability
```
```

#### 2. How to remediate

Provide step-by-step instructions to fix the issue:

```markdown
## How to remediate

### Fix [Issue Name]

**Check current state**:
```bash
# Commands to verify the issue
```

**Fix the issue**:
```bash
# Commands to remediate
```

**Verify**:
```bash
# Commands to confirm fix
```
```

#### 3. Security best practices

Provide broader security guidance:

```markdown
## Security best practices

### [Category 1]
- Best practice 1
- Best practice 2

### [Category 2]
- Best practice 3
- Best practice 4
```

### Example Signal Implementation

See `src/signals/docker_socket.go` and `docs/signals/docker_socket.md` for a complete example.

## Documentation Parity

**Critical requirement**: All signals must have matching documentation files.

- **Every signal** in `src/signals/` must have a corresponding file in `docs/signals/`
- **Documentation must be kept in sync** with code changes
- **Pre-push hook validates** documentation parity (installed via `make hooks`)
- **CI will fail** if documentation is missing or out of sync

### Validation

The pre-push hook checks:
1. Every signal has a documentation file
2. Documentation filename matches signal name
3. Documentation follows the required format

You can manually validate:
```bash
# Install hooks if not already done
make hooks

# Or manually run the validation
./scripts/hooks/pre-push
```

## Security Scanning

All code must pass security scanning with gosec:

```bash
# Run security scan
make gosec
```

**Requirements**:
- All PRs must pass gosec with **zero findings**
- Gosec runs in **audit mode** (stricter than default)
- **No `//nolint:gosec` or `// #nosec` comments** without maintainer approval
- Security findings must be fixed, not suppressed

**Common issues**:
- **G304 (CWE-22)**: File path injection - validate/sanitize file paths
- **G115**: Integer overflow - use explicit type conversions
- **G104 (CWE-703)**: Unhandled errors - always check error returns
- **G602 (CWE-118)**: Slice bounds - validate indices before access

## Git Hooks

Install Git hooks to catch issues before pushing:

```bash
make hooks
```

This installs:
- **pre-commit**: Checks Go formatting (`make fmt-check`)
- **pre-push**: Validates signal documentation parity

**Benefits**:
- Catch formatting issues before committing
- Ensure documentation is complete before pushing
- Faster feedback loop than waiting for CI

## Pull Request Process

### What to Expect

1. **Automated CI checks** will run:
   - ✅ Go formatting check (`make fmt-check`)
   - ✅ All tests pass (`make test`)
   - ✅ Security scan passes (`make gosec`)
   - ✅ Code coverage is maintained
   - ✅ Documentation validation

2. **Code review** by maintainers:
   - Usually within 1-3 business days
   - May request changes or ask questions
   - Be responsive to feedback

3. **Merge**:
   - Once approved and CI passes, maintainers will merge
   - Your contribution will be included in the next release!

### CI Checks That Must Pass

All of these must be green before merge:

- **format-check**: Code is properly formatted with `gofmt`
- **test**: All tests pass
- **build**: Project builds successfully
- **security**: Gosec scan passes with zero findings
- **codecov**: Code coverage is maintained (no significant drops)

### Tips for Faster Review

- **Keep PRs focused**: One feature or fix per PR
- **Write good descriptions**: Explain what and why, not just how
- **Add tests**: PRs with tests get merged faster
- **Respond promptly**: Address feedback quickly
- **Be patient**: Maintainers are volunteers with day jobs

## Questions or Need Help?

- **Open an issue**: For bugs, feature requests, or questions
- **Check existing issues**: Your question might already be answered

## License

By contributing to Dashlights, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

---

Thank you for contributing to Dashlights! Your efforts help make development environments more secure for everyone. 🚀

