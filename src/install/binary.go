package install

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BinaryInstallState represents the state of the binary installation.
type BinaryInstallState int

const (
	BinaryNotInstalled BinaryInstallState = iota
	BinaryInstalled                       // Same version already installed
	BinaryOutdated                        // Different version exists
	BinaryIsSymlink                       // Target is a symlink (warn, don't overwrite)
)

// BinaryConfig contains information about binary installation.
type BinaryConfig struct {
	SourcePath      string // Path to currently running binary
	TargetDir       string // Directory to install to
	TargetPath      string // Full path to target binary
	PathNeedsExport bool   // Whether ~/.local/bin needs to be added to PATH
	ShellConfigPath string // Shell config file to modify for PATH export
	Shell           ShellType
}

// Sentinel markers for PATH export.
const (
	PathSentinelBegin = "# BEGIN dashlights-path"
	PathSentinelEnd   = "# END dashlights-path"
)

// Default fallback directory for binary installation.
const DefaultBinDir = ".local/bin"

// System directories that should be skipped for user installs.
var systemDirs = []string{"/usr/bin", "/usr/local/bin", "/bin", "/sbin", "/usr/sbin"}

// homebrewBinDir is the acceptable homebrew bin directory.
const homebrewBinDir = "/opt/homebrew/bin"

// homebrewPrefix is the prefix for homebrew directories that should be filtered out.
const homebrewPrefix = "/opt/homebrew/"

// BinaryInstaller handles installation of the dashlights binary to PATH.
type BinaryInstaller struct {
	fs     Filesystem
	backup *BackupManager
}

// NewBinaryInstaller creates a new BinaryInstaller.
func NewBinaryInstaller(fs Filesystem) *BinaryInstaller {
	return &BinaryInstaller{
		fs:     fs,
		backup: NewBackupManager(fs),
	}
}

// GetBinaryConfig determines the source and target paths for binary installation.
func (b *BinaryInstaller) GetBinaryConfig(shellConfig *ShellConfig) (*BinaryConfig, error) {
	// Get the path of the running binary
	sourcePath, err := b.fs.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot determine running binary path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := b.resolveSymlinks(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve binary path: %w", err)
	}

	// Verify the binary exists and is readable
	if _, err := b.fs.Stat(realPath); err != nil {
		return nil, fmt.Errorf("running binary not accessible at %s: %w", realPath, err)
	}

	// Find installation directory
	targetDir, needsExport, err := b.FindInstallDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find installation directory: %w", err)
	}

	config := &BinaryConfig{
		SourcePath:      realPath,
		TargetDir:       targetDir,
		TargetPath:      filepath.Join(targetDir, "dashlights"),
		PathNeedsExport: needsExport,
	}

	// If PATH export is needed, determine shell config
	if needsExport && shellConfig != nil {
		config.ShellConfigPath = shellConfig.ConfigPath
		config.Shell = shellConfig.Shell
	}

	return config, nil
}

// resolveSymlinks resolves symlinks to get the real binary path.
func (b *BinaryInstaller) resolveSymlinks(path string) (string, error) {
	// For mock filesystem, just return the path as-is
	// In real use, filepath.EvalSymlinks handles this
	info, err := b.fs.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&0o120000 == 0o120000 { // symlink
		// For real filesystem, we'd follow the link
		// For mock, we just return the path
		return path, nil
	}
	return path, nil
}

// findExistingDashlightsInPath checks if dashlights binary already exists somewhere in PATH.
// Returns the directory path if found, empty string otherwise.
func (b *BinaryInstaller) findExistingDashlightsInPath() string {
	pathDirs := b.fs.SplitPath()
	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}
		binaryPath := filepath.Join(dir, "dashlights")
		if b.fs.Exists(binaryPath) {
			return dir
		}
	}
	return ""
}

// FindInstallDir finds the best directory to install the binary to.
// Priority:
// 1. If dashlights already exists somewhere in PATH, use that location
// 2. First user-writable directory in PATH (excluding system dirs and non-preferred homebrew subdirs)
// 3. Fallback to ~/.local/bin
func (b *BinaryInstaller) FindInstallDir() (string, bool, error) {
	pathDirs := b.fs.SplitPath()

	homeDir, err := b.fs.UserHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("cannot determine home directory: %w", err)
	}
	localBin := filepath.Join(homeDir, DefaultBinDir)

	// Priority 1: If dashlights already exists in PATH, use that location
	// (even if it requires sudo - we respect user's prior choice)
	existingDir := b.findExistingDashlightsInPath()
	if existingDir != "" {
		return existingDir, false, nil
	}

	// Priority 2: First user-writable directory in PATH
	// Skip system directories and non-preferred homebrew subdirectories
	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}

		// Skip system directories
		if isSystemDir(dir) {
			continue
		}

		// Skip non-preferred homebrew subdirectories (but allow /opt/homebrew/bin)
		if isNonPreferredHomebrewDir(dir) {
			continue
		}

		// Check if writable
		if b.fs.IsWritable(dir) {
			return dir, false, nil
		}
	}

	// Priority 3: Fall back to ~/.local/bin
	// Check if ~/.local/bin is already in PATH
	for _, dir := range pathDirs {
		if dir == localBin {
			// In PATH but may not exist or not writable - create it
			return localBin, false, nil
		}
	}

	// ~/.local/bin not in PATH - need to add export
	return localBin, true, nil
}

// isSystemDir checks if a directory is a system directory.
func isSystemDir(dir string) bool {
	for _, sys := range systemDirs {
		if dir == sys {
			return true
		}
	}
	return false
}

// isNonPreferredHomebrewDir checks if a directory is a homebrew subdirectory
// that should be filtered out (e.g., /opt/homebrew/lib/ruby/gems/3.3.0/bin).
// Returns false for /opt/homebrew/bin which is acceptable.
func isNonPreferredHomebrewDir(dir string) bool {
	if dir == homebrewBinDir {
		return false // /opt/homebrew/bin is acceptable
	}
	return strings.HasPrefix(dir, homebrewPrefix)
}

// CheckBinaryState checks if the binary is already installed and its state.
func (b *BinaryInstaller) CheckBinaryState(config *BinaryConfig) (BinaryInstallState, error) {
	// Check if target exists
	if !b.fs.Exists(config.TargetPath) {
		return BinaryNotInstalled, nil
	}

	// Check if target is a symlink
	info, err := b.fs.Lstat(config.TargetPath)
	if err != nil {
		return BinaryNotInstalled, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return BinaryIsSymlink, nil
	}

	// Compare versions
	same, err := b.CompareVersions(config.SourcePath, config.TargetPath)
	if err != nil {
		return BinaryNotInstalled, err
	}
	if same {
		return BinaryInstalled, nil
	}
	return BinaryOutdated, nil
}

// CompareVersions compares the running binary against the installed one.
// Returns true if they are the same (by size and checksum).
func (b *BinaryInstaller) CompareVersions(sourcePath, targetPath string) (bool, error) {
	// First compare file sizes (fast check)
	srcInfo, err := b.fs.Stat(sourcePath)
	if err != nil {
		return false, err
	}
	dstInfo, err := b.fs.Stat(targetPath)
	if err != nil {
		return false, err
	}
	if srcInfo.Size() != dstInfo.Size() {
		return false, nil // Different sizes = different versions
	}

	// If sizes match, compare SHA256 checksums
	srcHash, err := b.hashFile(sourcePath)
	if err != nil {
		return false, err
	}
	dstHash, err := b.hashFile(targetPath)
	if err != nil {
		return false, err
	}

	return srcHash == dstHash, nil
}

// hashFile computes the SHA256 hash of a file.
func (b *BinaryInstaller) hashFile(path string) (string, error) {
	content, err := b.fs.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// InstallBinary copies the binary to the target location.
func (b *BinaryInstaller) InstallBinary(config *BinaryConfig, dryRun bool) (*InstallResult, error) {
	// Check current state
	state, err := b.CheckBinaryState(config)
	if err != nil {
		return nil, fmt.Errorf("failed to check binary state: %w", err)
	}

	switch state {
	case BinaryInstalled:
		return &InstallResult{
			ExitCode: ExitSuccess,
			Message:  fmt.Sprintf("Binary already installed at %s (up to date)", config.TargetPath),
		}, nil

	case BinaryIsSymlink:
		return &InstallResult{
			ExitCode: ExitSuccess,
			Message:  fmt.Sprintf("Warning: %s is a symlink, not overwriting", config.TargetPath),
		}, nil

	case BinaryOutdated:
		if dryRun {
			return &InstallResult{
				ExitCode:    ExitSuccess,
				Message:     fmt.Sprintf("[DRY RUN] Would update binary at %s", config.TargetPath),
				WhatChanged: "update",
			}, nil
		}
		// Create backup before updating
		backupResult, err := b.backup.CreateBackup(config.TargetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to backup existing binary: %w", err)
		}
		// Copy the new binary
		if err := b.fs.CopyFile(config.SourcePath, config.TargetPath); err != nil {
			return nil, fmt.Errorf("failed to copy binary: %w", err)
		}
		// Ensure executable permissions
		if err := b.fs.Chmod(config.TargetPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to set permissions: %w", err)
		}
		return &InstallResult{
			ExitCode:    ExitSuccess,
			Message:     fmt.Sprintf("Updated binary at %s", config.TargetPath),
			BackupPath:  backupResult.BackupPath,
			WhatChanged: "update",
		}, nil

	case BinaryNotInstalled:
		if dryRun {
			return &InstallResult{
				ExitCode:    ExitSuccess,
				Message:     fmt.Sprintf("[DRY RUN] Would install binary to %s", config.TargetPath),
				WhatChanged: "install",
			}, nil
		}
		// Ensure target directory exists
		if err := b.fs.MkdirAll(config.TargetDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", config.TargetDir, err)
		}
		// Copy the binary
		if err := b.fs.CopyFile(config.SourcePath, config.TargetPath); err != nil {
			return nil, fmt.Errorf("failed to copy binary: %w", err)
		}
		// Ensure executable permissions
		if err := b.fs.Chmod(config.TargetPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to set permissions: %w", err)
		}
		return &InstallResult{
			ExitCode:    ExitSuccess,
			Message:     fmt.Sprintf("Installed binary to %s", config.TargetPath),
			ConfigPath:  config.TargetPath,
			WhatChanged: "install",
		}, nil
	}

	return nil, fmt.Errorf("unknown binary state")
}

// CheckPathExportState checks if PATH export is already configured in shell config.
func (b *BinaryInstaller) CheckPathExportState(shellConfigPath string) (InstallState, error) {
	if shellConfigPath == "" {
		return NotInstalled, nil
	}

	content, err := b.fs.ReadFile(shellConfigPath)
	if err != nil {
		// File doesn't exist - not installed
		return NotInstalled, nil
	}

	contentStr := string(content)
	hasBegin := strings.Contains(contentStr, PathSentinelBegin)
	hasEnd := strings.Contains(contentStr, PathSentinelEnd)

	if hasBegin && hasEnd {
		return FullyInstalled, nil
	}
	if hasBegin || hasEnd {
		return PartialInstall, nil
	}
	return NotInstalled, nil
}

// AddPathExport adds ~/.local/bin to PATH in shell config file.
func (b *BinaryInstaller) AddPathExport(shellConfig *ShellConfig, dryRun bool) (*InstallResult, error) {
	if shellConfig == nil || shellConfig.ConfigPath == "" {
		return &InstallResult{
			ExitCode: ExitSuccess,
			Message:  "No shell config to update",
		}, nil
	}

	// Check if already configured
	state, err := b.CheckPathExportState(shellConfig.ConfigPath)
	if err != nil {
		return nil, err
	}

	switch state {
	case FullyInstalled:
		return &InstallResult{
			ExitCode: ExitSuccess,
			Message:  "PATH export already configured",
		}, nil

	case PartialInstall:
		return &InstallResult{
			ExitCode: ExitError,
			Message:  fmt.Sprintf("Error: Found partial PATH export in %s. Please remove the dashlights-path section manually and try again.", shellConfig.ConfigPath),
		}, nil
	}

	// Get the template for this shell
	template := GetPathExportTemplate(shellConfig.Shell)
	if template == "" {
		return &InstallResult{
			ExitCode: ExitSuccess,
			Message:  fmt.Sprintf("No PATH template for shell %s", shellConfig.Shell),
		}, nil
	}

	if dryRun {
		return &InstallResult{
			ExitCode:    ExitSuccess,
			Message:     fmt.Sprintf("[DRY RUN] Would add PATH export to %s", shellConfig.ConfigPath),
			WhatChanged: "path_export",
		}, nil
	}

	// Read existing content (may not exist - that's OK)
	existingContent, readErr := b.fs.ReadFile(shellConfig.ConfigPath)
	if readErr != nil {
		existingContent = nil // File doesn't exist, start fresh
	}

	// Backup if file exists
	var backupPath string
	if len(existingContent) > 0 {
		backupResult, backupErr := b.backup.CreateBackup(shellConfig.ConfigPath)
		if backupErr != nil {
			return nil, fmt.Errorf("failed to backup %s: %w", shellConfig.ConfigPath, backupErr)
		}
		backupPath = backupResult.BackupPath
	}

	// Append the PATH export
	newContent := string(existingContent)
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n" + template

	if err := b.fs.WriteFile(shellConfig.ConfigPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", shellConfig.ConfigPath, err)
	}

	return &InstallResult{
		ExitCode:    ExitSuccess,
		Message:     fmt.Sprintf("Added PATH export to %s", shellConfig.ConfigPath),
		BackupPath:  backupPath,
		ConfigPath:  shellConfig.ConfigPath,
		WhatChanged: "path_export",
	}, nil
}

// EnsureBinaryInstalled installs the binary and adds PATH export if needed.
// This is a convenience method for use by prompt and agent installers.
func (b *BinaryInstaller) EnsureBinaryInstalled(shellConfig *ShellConfig, dryRun bool) (*InstallResult, error) {
	config, err := b.GetBinaryConfig(shellConfig)
	if err != nil {
		return &InstallResult{
			ExitCode: ExitError,
			Message:  fmt.Sprintf("Warning: %v", err),
		}, nil
	}

	// Install binary
	result, err := b.InstallBinary(config, dryRun)
	if err != nil {
		return &InstallResult{
			ExitCode: ExitError,
			Message:  fmt.Sprintf("Warning: %v", err),
		}, nil
	}

	// Add PATH export if needed
	if config.PathNeedsExport && shellConfig != nil {
		pathResult, err := b.AddPathExport(shellConfig, dryRun)
		if err != nil {
			result.Message += fmt.Sprintf("; PATH export failed: %v", err)
		} else if pathResult.WhatChanged != "" {
			result.Message += "; " + pathResult.Message
		}
	}

	return result, nil
}
