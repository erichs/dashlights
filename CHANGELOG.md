# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-alpha.7] - 2025-12-04

### Added
- Added InsecureCurlPipe signal to detect insecure use of curl pipes

### Changed
- Improved PrivilegedPath signal checks for better accuracy
- Refined PrivilegedPath signal with various minor fixes
- Excluded internal packages from pre-push documentation validation to streamline checks
- Refactored code by extracting internal package helper functions for cleaner organization
- Addressed static analysis and linter suggestions to enhance code quality


## [1.0.0-alpha.6] - 2025-12-03

### Added
- Documented capability enforcement to clarify security controls
- Added support for running integration tests in network-isolated and dockerized environments

### Changed
- Standardized Go version used in continuous integration for consistency
- Hardened CI test runner and permissions to improve security and reliability
- Refactored PwnRequest to Unsafe Workflow for improved internal handling

### Fixed
- Fixed Go generate step to enhance build hardening
- Corrected Docker mount pathing issues
- Disallowed import of 'net/http' in tests to enforce stricter dependencies
- Applied various minor fixes to improve stability and correctness


## [1.0.0-alpha.5] - 2025-12-02

### Added
- Added DangerZone signal to detect when running as the root user

### Changed
- Simplified Docker socket detection signal for improved reliability


## [1.0.0-alpha.4] - 2025-12-02

### Added
- Added a new signal for detecting missing Git hooks

### Changed
- Updated and refactored SIGNALS documentation for clarity and better organization
- Documented design intent to improve understanding of project goals
- Made minor documentation improvements and adjusted visibility settings


## [1.0.0-alpha.3b] - 2025-12-01

### Fixed
- Fixed release tag detection to ensure accurate versioning during releases


## [1.0.0-alpha.3] - 2025-12-01

### Added
- Added PwnRequest GitHub Action signal to enhance security scanning

### Changed
- Increased test coverage for PwnRequest to improve reliability
- Documented the requirement for temporary quarantine authorization on Darwin systems
- Adopted a pull request-based release flow for better release management


## [1.0.0-alpha.2] - 2025-12-01

### Changed
- Ensured adherence to the 10ms hard time limit for improved performance consistency


## [1.0.0-alpha.1] - 2025-11-30

### Changed
- Simplified release platform target configurations for easier maintenance
- Updated changelog to include a note about breaking changes


## [1.0.0-alpha] - 2025-11-30

### Upgrade Notes
This is a major breaking change to the design of dashlights. It still supports the previous custom emoji and colors via ENV vars, but this is now a minor feature, and may be deprecated in a future release. See install instructions in README for install/upgrade details.

### Added
- Added a security policy document outlining vulnerability disclosure procedures
- Added verbose diagnostic mode with portable documentation links
- Added comprehensive documentation for all trouble signals
- Added new infrastructure security signals including Docker socket and SSH key permissions
- Added language-specific repository hygiene signals for better code quality monitoring
- Added operational security signals for history permissions, SSH agent key bloat, and DEBUG environment variables
- Added system health signals for time drift and dangling symlinks
- Added AWS CLI alias injection detection signal
- Added support for version flag and release automation for changelog updates
- Added badge for OpenSSF best practices compliance
- Added contributor guidelines to support community contributions
- Added detection of modern environment secrets management tools

### Changed
- Revised command line flags for improved clarity and usability
- Refactored codebase into src/ directory for better organization
- Updated README with new mascot and general improvements
- Updated prompt guidelines for p10k users
- Configured Codecov for test coverage monitoring
- Improved concurrency and build processes for better performance
- Cleaned up security override annotations for clearer code

### Fixed
- Fixed validation for pre-release and versioned pre-release tags in semantic version strings
- Fixed release target configurations
- Fixed Docker socket signal behavior on Darwin (macOS) systems
- Fixed missing initialization signal for Python projects
- Fixed security pipeline to ensure code generation runs before scanning
- Fixed and addressed multiple security findings including G115, G104/CWE-703, G602/CWE-118, and G304/CWE-22

### Security
- Ensured gosec scans run with audit mode enabled for enhanced security analysis
- Improved detection and handling of environment secrets to prevent leaks


## [0.3.0] - 2025-11-26

### Added
- Added support for emoji alias labels to enhance labeling options

## [0.2.1] - 2025-11-26

### Changed
- Modernized the project and prepared a new release

## [0.2.0] - 2017-07-03

Initial release with basic security signal detection.

[0.3.0]: https://github.com/erichs/dashlights/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/erichs/dashlights/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/erichs/dashlights/releases/tag/v0.2.0

