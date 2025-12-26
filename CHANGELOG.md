# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.2] - 2025-12-26

### Added
- Added MacOS Secure Keyboard Entry detection for supported terminals

### Changed
- Updated documentation for Secure Keyboard Entry feature
- Improved filename filtering in DumpsterFire component


## [1.1.1] - 2025-12-21

### Added
- Automated installation into prompts and agent hooks

### Changed
- Documented installation options for easier setup
- Ensured end-to-end installation tests run before release to improve reliability


## [1.1.0] - 2025-12-19

This release introduces --agentic mode, see docs/agentic_mode.md for
details.

This mode is intended to be used with coding agents that support tool
hooks, currently Claude Code and Cursor.

### Added
- Added critical threat detection for Claude configuration writes and invisible Unicode characters
- Added file redirection and tee detection heuristics for agentic mode
- Added tests to document symlink behavior in file read operations
- Added support for Cursor in agentic mode

### Changed
- Improved agentic mode debug handling to avoid swallowing JSON errors
- Improved data collection and diagnostics for invisible Unicode scanning
- Improved context cancellation behavior for multiple signals to enhance responsiveness
- Clarified supported hooks in agentic mode for better user understanding
- Refactored agentic package for improved structure and maintainability
- Hardened file and agentic input handling with bounded reads to improve safety and stability
- Upgraded to Go version 1.25 for better performance and compatibility
- Tweaked README documentation for clarity

### Fixed
- Handled error cases during debug mode propagation to prevent silent failures
- Ignored swap files to avoid unnecessary processing
- Detected use of in-place editors when modifying critical agent configuration to prevent unnoticed changes

### Security
- Improved detection of critical agent configuration modifications to enhance security monitoring

### Testing
- Increased test coverage for main application code and agentic threat detection components


## [1.0.7-slsa-2] - 2025-12-17

### Fixed
- Corrected the order of steps in the release process to ensure proper packaging and deployment


## [1.0.7-slsa] - 2025-12-17

### Changed
- Streamlined SBOM generation and cosigning process for improved efficiency


## [1.0.7] - 2025-12-17

### Changed
- Improved context cancellation handling for zombie processes detection
- Enhanced performance of Rotting Secrets and Dumpster Fire signals
- Addressed filesystem variability issues in continuous integration
- Configured Software Bill of Materials (SBOM) generation for better supply chain transparency
- Pinned GitHub Actions dependencies, limited permissions, and signed artifacts with Cosign for improved security and reliability


## [1.0.6] - 2025-12-16

### Changed
- Improved performance by reducing the 95th percentile latency by 4.5ms for SSH agent under heavy load


## [1.0.5] - 2025-12-15

### Changed
- Improved performance of PycachePollution and MissingInitPy signals for large directories
- Refactored pycache_pollution code to reduce duplication and improve maintainability
- Fixed minor issue in npmrc token signal

### Added
- Updated README documentation


## [1.0.4] - 2025-12-14

### Added
- Added a --debug flag and load testing utilities to aid in troubleshooting and performance evaluation

### Changed
- Returned partial results on timeout to improve responsiveness instead of blocking
- Improved test coverage for core application logic
- Cleaned up README for better clarity and usability
- Updated .gitignore to exclude additional unnecessary files

### Fixed
- Addressed review feedback to enhance code quality and consistency


## [1.0.3] - 2025-12-13

### Fixed
- Fixed release publishing process to support multi-step releases


## [1.0.2] - 2025-12-13

### Changed
- Bumped version to trigger release pipeline run


## [1.0.1] - 2025-12-13

### Changed
- Updated the README documentation for clarity and accuracy


## [1.0.0] - 2025-12-13

### Added
- Apple code signing and notarization to the release process for macOS
  This requires a one-time online validation with Apple's notary servers

### Fixed
- Minor issues with the macOS release process


## [1.0.0-alpha.8] - 2025-12-09

### Added
- Added DumpsterFire and RottingSecrets security signals

### Changed
- Provided detailed remediation guidance for the privileged_path signal
- Supported selective disabling of specific signal checks
- Addressed linter issues for improved code quality

### Fixed
- Updated README documentation


## [1.0.0-alpha.7] - 2025-12-04

### Added
- Added InsecureCurlPipe signal to detect insecure use of curl | bash

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

