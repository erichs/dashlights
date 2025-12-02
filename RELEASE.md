# Release Process

This document describes how to create a new release of dashlights.

## Overview

Releases are created from **feature branches** and merged to main via Pull Request. This is required because the main branch is protected and does not allow direct pushes.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Release Workflow                             │
├─────────────────────────────────────────────────────────────────────┤
│  1. Create feature branch                                           │
│  2. Make changes, commit, push → CI builds/scans                    │
│  3. Run `make release` → Creates changelog commit + tag             │
│  4. Push branch + tag → CI builds/scans again                       │
│  5. Create PR and merge to main                                     │
│  6. Release workflow triggers automatically → GoReleaser runs       │
└─────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### Fabric CLI

This project uses [Fabric](https://github.com/danielmiessler/fabric) by Daniel Miessler to generate changelog entries from git commit messages using AI.

#### Installation

Install Fabric following the [official installation instructions](https://github.com/danielmiessler/fabric?tab=readme-ov-file#installation).

**Quick install options:**

```bash
# macOS/Linux (recommended)
curl -s https://raw.githubusercontent.com/danielmiessler/fabric/main/installer/client/client.sh | bash

# Or with Go
go install github.com/danielmiessler/fabric@latest

# Or with Homebrew (macOS)
brew install fabric
```

After installation, configure Fabric with your AI provider:

```bash
fabric --setup
```

### Install the Changelog Pattern

Once Fabric is installed, install the custom changelog generation pattern:

```bash
make install-fabric-pattern
```

This copies the pattern from `scripts/fabric-patterns/create_git_changelog/system.md` to `~/.config/fabric/patterns/create_git_changelog/`.

## Make Targets

### `make install-fabric-pattern`

Installs the custom Fabric pattern for changelog generation.

- **Prerequisites**: Fabric CLI must be installed
- **What it does**: Copies the pattern to `~/.config/fabric/patterns/create_git_changelog/`
- **When to run**: Once after cloning the repo, or after updating the pattern

### `make release`

Creates a new release with an AI-generated changelog.

- **Prerequisites**: 
  - Fabric CLI installed
  - Fabric pattern installed (`make install-fabric-pattern`)
  - Clean working directory (no uncommitted changes)
- **What it does**: Runs `scripts/release.sh` (see workflow below)
- **When to run**: When you're ready to create a new release

## Release Workflow

### Step 1: Create a Feature Branch

Releases must be created from a feature branch, not from main:

```bash
git checkout main
git pull origin main
git checkout -b release-v0.4.0   # or any branch name
```

### Step 2: Ensure Prerequisites

```bash
# Check if fabric is installed
which fabric

# Install the changelog pattern
make install-fabric-pattern
```

### Step 3: Prepare for Release

Ensure your working directory is clean and all changes are committed:

```bash
git status
# Should show: "nothing to commit, working tree clean"
```

### Step 4: Run the Release Command

```bash
make release
```

> **Note**: This command will fail if run from the main branch. You must be on a feature branch.

### Step 5: Interactive Prompts

The release script will guide you through the process:

1. **Shows the latest tag** (e.g., `v0.3.0`)
2. **Prompts for new version** - Enter semver format (e.g., `0.4.0`)
3. **Generates changelog** - Uses Fabric to analyze commits since the last tag
4. **Shows the generated changelog** - Review the AI-generated content
5. **Prompts for confirmation** - Type `y` to proceed or `n` to cancel

### Step 6: Push Branch and Tag

If you confirmed, the script will:
- Commit `CHANGELOG.md` with message `"Update CHANGELOG for v0.4.0"`
- Create an annotated git tag `v0.4.0`

Then push the branch and tag together:

```bash
# Review the changes
git show v0.4.0

# Push branch and tag
git push origin release-v0.4.0 --tags
```

### Step 7: Create and Merge Pull Request

1. Go to GitHub and create a Pull Request from your branch to main
2. Wait for CI checks to pass
3. Merge the PR

### Step 8: Automatic Release

After the PR is merged, the release workflow automatically:
1. Detects the release tag on the merged commit
2. Checks out the tagged commit
3. Runs GoReleaser to build and publish the release

## What the Script Does

The `scripts/release.sh` script automates the following:

1. **Validation**
   - Checks that you're not on the main branch
   - Checks if Fabric CLI is installed
   - Checks if the Fabric pattern is installed
   - Verifies working directory is clean
   - Validates semver format

2. **Changelog Generation**
   - Gets commits since last tag: `git log v0.3.0..HEAD --oneline --no-merges`
   - Pipes to Fabric: `fabric -p create_git_changelog`
   - AI categorizes commits into: Added, Changed, Fixed, Security, Removed

3. **Changelog Update**
   - Prepends new version entry to `CHANGELOG.md`
   - Includes version number and release date
   - Maintains Keep a Changelog format

4. **Git Operations**
   - Commits `CHANGELOG.md`
   - Creates annotated tag with version number

5. **Next Steps**
   - Displays instructions for pushing to GitHub

## Changelog Format

The generated changelog follows [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [0.4.0] - 2025-11-30

### Added
- New features and functionality

### Changed
- Improvements and updates to existing features

### Fixed
- Bug fixes and corrections

### Security
- Security fixes and improvements
```

## Fabric Pattern Details

The custom Fabric pattern (`create_git_changelog`) is designed to:

- **Categorize commits** into meaningful sections (Added, Changed, Fixed, Security)
- **Remove duplicates** and merge related changes
- **Rewrite technical messages** into user-friendly descriptions
- **Focus on user impact** rather than implementation details
- **Order by importance** within each category
- **Follow conventions** of Keep a Changelog format

The pattern is stored in `scripts/fabric-patterns/create_git_changelog/system.md` and can be customized to fit your project's needs.

## Troubleshooting

### "Cannot create releases from the main branch"

You must create releases from a feature branch:

```bash
git checkout -b release-prep
make release
```

### "fabric: command not found"

Install Fabric CLI following the [installation instructions](https://github.com/danielmiessler/fabric?tab=readme-ov-file#installation).

### "Fabric pattern not installed"

Run `make install-fabric-pattern` to install the custom pattern.

### "Working directory is not clean"

Commit or stash your changes before creating a release:

```bash
git status
git add .
git commit -m "Your commit message"
```

### Changelog generation fails

Ensure Fabric is properly configured with an AI provider:

```bash
fabric --setup
```

### Release workflow didn't trigger after PR merge

The release workflow looks for a version tag (e.g., `v0.4.0`) on HEAD~1 (the commit before the merge commit). Verify:

1. The tag was pushed: `git tag -l 'v*'`
2. The tag points to the correct commit: `git show v0.4.0`
3. Check the Actions tab in GitHub for workflow run details

## Example Release Session

```bash
# Start from a feature branch
$ git checkout -b release-v0.4.0

$ make release
Latest tag: v0.3.0

Enter new version (semver format, e.g., 0.4.0 or 1.0.0-alpha):
0.4.0

Generating changelog for v0.4.0...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
New changelog entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## [0.4.0] - 2025-11-30

### Added
- macOS Docker Desktop support with platform-specific socket handling
- Comprehensive test coverage for Docker socket signal

### Changed
- Updated documentation for Docker socket signal with platform-specific guidance

### Security
- Fixed Docker socket detection to properly handle macOS symlinks
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Create tag v0.4.0 and commit CHANGELOG.md? (y/n):
y

✅ Release v0.4.0 created successfully!

Next steps:
  1. Review the changes: git show v0.4.0
  2. Push branch and tag: git push origin release-v0.4.0 --tags
  3. Create a PR to merge release-v0.4.0 into main
  4. Merge the PR - the release workflow will trigger automatically

# Push and create PR
$ git push origin release-v0.4.0 --tags
# Then create PR on GitHub and merge
```

