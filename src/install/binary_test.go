package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsNonPreferredHomebrewDir(t *testing.T) {
	tests := []struct {
		dir      string
		expected bool
	}{
		{"/opt/homebrew/bin", false},                    // acceptable
		{"/opt/homebrew/lib/ruby/gems/3.3.0/bin", true}, // non-preferred
		{"/opt/homebrew/opt/openjdk@21/bin", true},      // non-preferred
		{"/opt/homebrew/Cellar/python@3.11/bin", true},  // non-preferred
		{"/usr/local/bin", false},                       // not homebrew
		{"/home/user/bin", false},                       // not homebrew
		{"/opt/homebrew", false},                        // just prefix, not a subdir
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			result := isNonPreferredHomebrewDir(tt.dir)
			if result != tt.expected {
				t.Errorf("isNonPreferredHomebrewDir(%q) = %v, want %v", tt.dir, result, tt.expected)
			}
		})
	}
}

func TestBinaryInstaller_FindExistingDashlightsInPath(t *testing.T) {
	fs := NewMockFilesystem()
	fs.PathEnv = "/usr/bin:/home/testuser/bin:/usr/local/bin"
	// Simulate dashlights existing in /home/testuser/bin
	fs.Files["/home/testuser/bin/dashlights"] = []byte("binary")

	bi := NewBinaryInstaller(fs)
	existingDir := bi.findExistingDashlightsInPath()

	if existingDir != "/home/testuser/bin" {
		t.Errorf("expected /home/testuser/bin, got %s", existingDir)
	}
}

func TestBinaryInstaller_FindExistingDashlightsInPath_NotFound(t *testing.T) {
	fs := NewMockFilesystem()
	fs.PathEnv = "/usr/bin:/home/testuser/bin:/usr/local/bin"
	// No dashlights binary exists

	bi := NewBinaryInstaller(fs)
	existingDir := bi.findExistingDashlightsInPath()

	if existingDir != "" {
		t.Errorf("expected empty string, got %s", existingDir)
	}
}

func TestBinaryInstaller_FindInstallDir_RespectsExistingLocation(t *testing.T) {
	fs := NewMockFilesystem()
	// Dashlights already exists in /usr/local/bin (even though it's a system dir)
	fs.PathEnv = "/home/testuser/bin:/usr/local/bin:/usr/bin"
	fs.Files["/usr/local/bin/dashlights"] = []byte("existing binary")
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use existing location even though /home/testuser/bin is writable
	if dir != "/usr/local/bin" {
		t.Errorf("expected /usr/local/bin (existing location), got %s", dir)
	}
	if needsExport {
		t.Error("expected needsExport=false for existing location")
	}
}

func TestBinaryInstaller_FindInstallDir_SkipsHomebrewSubdirs(t *testing.T) {
	fs := NewMockFilesystem()
	// PATH has homebrew subdirs before the preferred dir
	fs.PathEnv = "/opt/homebrew/lib/ruby/gems/3.3.0/bin:/opt/homebrew/opt/openjdk@21/bin:/home/testuser/bin"
	fs.WritableDirs["/opt/homebrew/lib/ruby/gems/3.3.0/bin"] = true
	fs.WritableDirs["/opt/homebrew/opt/openjdk@21/bin"] = true
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should skip homebrew subdirs and use /home/testuser/bin
	if dir != "/home/testuser/bin" {
		t.Errorf("expected /home/testuser/bin (skipping homebrew subdirs), got %s", dir)
	}
	if needsExport {
		t.Error("expected needsExport=false")
	}
}

func TestBinaryInstaller_FindInstallDir_AllowsHomebrewBin(t *testing.T) {
	fs := NewMockFilesystem()
	// PATH has /opt/homebrew/bin which should be allowed
	fs.PathEnv = "/opt/homebrew/bin:/usr/bin"
	fs.WritableDirs["/opt/homebrew/bin"] = true

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use /opt/homebrew/bin (it's acceptable)
	if dir != "/opt/homebrew/bin" {
		t.Errorf("expected /opt/homebrew/bin, got %s", dir)
	}
	if needsExport {
		t.Error("expected needsExport=false")
	}
}

func TestBinaryInstaller_FindInstallDir_WritableInPath(t *testing.T) {
	fs := NewMockFilesystem()
	fs.PathEnv = "/usr/bin:/home/testuser/bin:/usr/local/bin"
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/home/testuser/bin" {
		t.Errorf("expected /home/testuser/bin, got %s", dir)
	}
	if needsExport {
		t.Error("expected needsExport=false for writable dir in PATH")
	}
}

func TestBinaryInstaller_FindInstallDir_FallbackToLocalBin(t *testing.T) {
	fs := NewMockFilesystem()
	fs.PathEnv = "/usr/bin:/usr/local/bin"
	// No writable dirs

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(fs.HomeDir, ".local/bin")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
	if !needsExport {
		t.Error("expected needsExport=true when falling back to ~/.local/bin")
	}
}

func TestBinaryInstaller_FindInstallDir_LocalBinAlreadyInPath(t *testing.T) {
	fs := NewMockFilesystem()
	localBin := filepath.Join(fs.HomeDir, ".local/bin")
	fs.PathEnv = "/usr/bin:" + localBin + ":/usr/local/bin"
	// ~/.local/bin is in PATH but not writable (doesn't exist)

	bi := NewBinaryInstaller(fs)
	dir, needsExport, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != localBin {
		t.Errorf("expected %s, got %s", localBin, dir)
	}
	if needsExport {
		t.Error("expected needsExport=false when ~/.local/bin already in PATH")
	}
}

func TestBinaryInstaller_FindInstallDir_SkipsSystemDirs(t *testing.T) {
	fs := NewMockFilesystem()
	fs.PathEnv = "/usr/bin:/bin:/usr/local/bin:/home/testuser/bin"
	// Make system dirs writable - they should still be skipped
	fs.WritableDirs["/usr/bin"] = true
	fs.WritableDirs["/bin"] = true
	fs.WritableDirs["/usr/local/bin"] = true
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	dir, _, err := bi.FindInstallDir()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should skip system dirs and use the user's bin
	if dir != "/home/testuser/bin" {
		t.Errorf("expected /home/testuser/bin (skipping system dirs), got %s", dir)
	}
}

func TestBinaryInstaller_CheckBinaryState_NotInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	bi := NewBinaryInstaller(fs)

	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	state, err := bi.CheckBinaryState(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != BinaryNotInstalled {
		t.Errorf("expected BinaryNotInstalled, got %d", state)
	}
}

func TestBinaryInstaller_CheckBinaryState_Installed(t *testing.T) {
	fs := NewMockFilesystem()
	// Same content = same version
	fs.Files["/tmp/dashlights"] = []byte("binary content v1")
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("binary content v1")

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	state, err := bi.CheckBinaryState(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != BinaryInstalled {
		t.Errorf("expected BinaryInstalled, got %d", state)
	}
}

func TestBinaryInstaller_CheckBinaryState_Outdated(t *testing.T) {
	fs := NewMockFilesystem()
	// Different content = different version
	fs.Files["/tmp/dashlights"] = []byte("binary content v2")
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("binary content v1")

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	state, err := bi.CheckBinaryState(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != BinaryOutdated {
		t.Errorf("expected BinaryOutdated, got %d", state)
	}
}

func TestBinaryInstaller_CheckBinaryState_Symlink(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("binary content")
	// Target is a symlink
	fs.Symlinks["/home/testuser/.local/bin/dashlights"] = "/opt/dashlights/bin/dashlights"
	// Also add to Files so Exists returns true
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("symlink target")

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	state, err := bi.CheckBinaryState(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != BinaryIsSymlink {
		t.Errorf("expected BinaryIsSymlink, got %d", state)
	}
}

func TestBinaryInstaller_CompareVersions_Same(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/src/binary"] = []byte("identical content")
	fs.Files["/dst/binary"] = []byte("identical content")

	bi := NewBinaryInstaller(fs)
	same, err := bi.CompareVersions("/src/binary", "/dst/binary")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !same {
		t.Error("expected files to be the same")
	}
}

func TestBinaryInstaller_CompareVersions_DifferentSize(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/src/binary"] = []byte("short")
	fs.Files["/dst/binary"] = []byte("this is longer content")

	bi := NewBinaryInstaller(fs)
	same, err := bi.CompareVersions("/src/binary", "/dst/binary")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if same {
		t.Error("expected files to be different (different sizes)")
	}
}

func TestBinaryInstaller_CompareVersions_SameSizeDifferentContent(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/src/binary"] = []byte("content A")
	fs.Files["/dst/binary"] = []byte("content B")

	bi := NewBinaryInstaller(fs)
	same, err := bi.CompareVersions("/src/binary", "/dst/binary")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if same {
		t.Error("expected files to be different (same size, different content)")
	}
}

func TestBinaryInstaller_InstallBinary_Success(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("binary content")
	fs.Modes["/tmp/dashlights"] = 0755

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetDir:  "/home/testuser/.local/bin",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	result, err := bi.InstallBinary(config, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	if result.WhatChanged != "install" {
		t.Errorf("expected WhatChanged=install, got %s", result.WhatChanged)
	}

	// Verify binary was copied
	if _, ok := fs.Files["/home/testuser/.local/bin/dashlights"]; !ok {
		t.Error("binary was not copied to target")
	}
}

func TestBinaryInstaller_InstallBinary_AlreadyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("binary content")
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("binary content")

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetDir:  "/home/testuser/.local/bin",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	result, err := bi.InstallBinary(config, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	if result.WhatChanged != "" {
		t.Errorf("expected no change, got WhatChanged=%s", result.WhatChanged)
	}
}

func TestBinaryInstaller_InstallBinary_Update(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("new binary v2")
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("old binary v1")
	fs.Modes["/tmp/dashlights"] = 0755

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetDir:  "/home/testuser/.local/bin",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	result, err := bi.InstallBinary(config, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	if result.WhatChanged != "update" {
		t.Errorf("expected WhatChanged=update, got %s", result.WhatChanged)
	}
	if result.BackupPath == "" {
		t.Error("expected backup to be created")
	}

	// Verify binary was updated
	content := string(fs.Files["/home/testuser/.local/bin/dashlights"])
	if content != "new binary v2" {
		t.Errorf("binary was not updated, got content: %s", content)
	}
}

func TestBinaryInstaller_InstallBinary_DryRun(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("binary content")

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetDir:  "/home/testuser/.local/bin",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	result, err := bi.InstallBinary(config, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	if result.WhatChanged != "install" {
		t.Errorf("expected WhatChanged=install for dry run, got %s", result.WhatChanged)
	}

	// Verify binary was NOT copied
	if _, ok := fs.Files["/home/testuser/.local/bin/dashlights"]; ok {
		t.Error("binary should not be copied in dry run mode")
	}
}

func TestBinaryInstaller_InstallBinary_SymlinkWarning(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/tmp/dashlights"] = []byte("binary content")
	fs.Files["/home/testuser/.local/bin/dashlights"] = []byte("symlink placeholder")
	fs.Symlinks["/home/testuser/.local/bin/dashlights"] = "/opt/dashlights"

	bi := NewBinaryInstaller(fs)
	config := &BinaryConfig{
		SourcePath: "/tmp/dashlights",
		TargetDir:  "/home/testuser/.local/bin",
		TargetPath: "/home/testuser/.local/bin/dashlights",
	}

	result, err := bi.InstallBinary(config, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	// Should not have changed anything
	if result.WhatChanged != "" {
		t.Errorf("expected no change for symlink, got WhatChanged=%s", result.WhatChanged)
	}
}

func TestBinaryInstaller_AddPathExport_Success(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/home/testuser/.bashrc"] = []byte("# existing content\n")

	bi := NewBinaryInstaller(fs)
	shellConfig := &ShellConfig{
		Shell:      ShellBash,
		ConfigPath: "/home/testuser/.bashrc",
	}

	result, err := bi.AddPathExport(shellConfig, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	if result.WhatChanged != "path_export" {
		t.Errorf("expected WhatChanged=path_export, got %s", result.WhatChanged)
	}

	// Verify PATH export was added
	content := string(fs.Files["/home/testuser/.bashrc"])
	if !contains(content, PathSentinelBegin) {
		t.Error("PATH export sentinel not found in config")
	}
	if !contains(content, "export PATH=") {
		t.Error("PATH export statement not found in config")
	}
}

func TestBinaryInstaller_AddPathExport_AlreadyPresent(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/home/testuser/.bashrc"] = []byte("# existing\n" + PathSentinelBegin + "\nexport PATH\n" + PathSentinelEnd + "\n")

	bi := NewBinaryInstaller(fs)
	shellConfig := &ShellConfig{
		Shell:      ShellBash,
		ConfigPath: "/home/testuser/.bashrc",
	}

	result, err := bi.AddPathExport(shellConfig, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}
	// Should not have changed anything
	if result.WhatChanged != "" {
		t.Errorf("expected no change when already present, got WhatChanged=%s", result.WhatChanged)
	}
}

func TestBinaryInstaller_AddPathExport_DryRun(t *testing.T) {
	fs := NewMockFilesystem()
	originalContent := "# existing content\n"
	fs.Files["/home/testuser/.bashrc"] = []byte(originalContent)

	bi := NewBinaryInstaller(fs)
	shellConfig := &ShellConfig{
		Shell:      ShellBash,
		ConfigPath: "/home/testuser/.bashrc",
	}

	result, err := bi.AddPathExport(shellConfig, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.WhatChanged != "path_export" {
		t.Errorf("expected WhatChanged=path_export for dry run, got %s", result.WhatChanged)
	}

	// Verify config was NOT modified
	content := string(fs.Files["/home/testuser/.bashrc"])
	if content != originalContent {
		t.Error("config should not be modified in dry run mode")
	}
}

func TestBinaryInstaller_CheckPathExportState_NotInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/home/testuser/.bashrc"] = []byte("# normal config\n")

	bi := NewBinaryInstaller(fs)
	state, err := bi.CheckPathExportState("/home/testuser/.bashrc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != NotInstalled {
		t.Errorf("expected NotInstalled, got %d", state)
	}
}

func TestBinaryInstaller_CheckPathExportState_FullyInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.Files["/home/testuser/.bashrc"] = []byte(PathSentinelBegin + "\n" + PathSentinelEnd + "\n")

	bi := NewBinaryInstaller(fs)
	state, err := bi.CheckPathExportState("/home/testuser/.bashrc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != FullyInstalled {
		t.Errorf("expected FullyInstalled, got %d", state)
	}
}

func TestBinaryInstaller_CheckPathExportState_PartialInstall(t *testing.T) {
	fs := NewMockFilesystem()
	// Only begin marker, no end
	fs.Files["/home/testuser/.bashrc"] = []byte(PathSentinelBegin + "\nexport PATH\n")

	bi := NewBinaryInstaller(fs)
	state, err := bi.CheckPathExportState("/home/testuser/.bashrc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != PartialInstall {
		t.Errorf("expected PartialInstall, got %d", state)
	}
}

func TestBinaryInstaller_GetBinaryConfig(t *testing.T) {
	fs := NewMockFilesystem()
	fs.ExecutablePath = "/tmp/downloaded/dashlights"
	fs.Files["/tmp/downloaded/dashlights"] = []byte("binary")
	fs.PathEnv = "/usr/bin:/home/testuser/bin"
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	shellConfig := &ShellConfig{
		Shell:      ShellBash,
		ConfigPath: "/home/testuser/.bashrc",
	}

	config, err := bi.GetBinaryConfig(shellConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.SourcePath != "/tmp/downloaded/dashlights" {
		t.Errorf("expected source path /tmp/downloaded/dashlights, got %s", config.SourcePath)
	}
	if config.TargetDir != "/home/testuser/bin" {
		t.Errorf("expected target dir /home/testuser/bin, got %s", config.TargetDir)
	}
	if config.PathNeedsExport {
		t.Error("expected PathNeedsExport=false for writable dir in PATH")
	}
}

func TestBinaryInstaller_GetBinaryConfig_ExecutableError(t *testing.T) {
	fs := NewMockFilesystem()
	fs.ExecutableErr = os.ErrNotExist

	bi := NewBinaryInstaller(fs)
	_, err := bi.GetBinaryConfig(nil)

	if err == nil {
		t.Error("expected error when executable path cannot be determined")
	}
}

func TestBinaryInstaller_EnsureBinaryInstalled(t *testing.T) {
	fs := NewMockFilesystem()
	fs.ExecutablePath = "/tmp/dashlights"
	fs.Files["/tmp/dashlights"] = []byte("binary")
	fs.PathEnv = "/home/testuser/bin"
	fs.WritableDirs["/home/testuser/bin"] = true

	bi := NewBinaryInstaller(fs)
	shellConfig := &ShellConfig{
		Shell:      ShellBash,
		ConfigPath: "/home/testuser/.bashrc",
	}

	result, err := bi.EnsureBinaryInstalled(shellConfig, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != ExitSuccess {
		t.Errorf("expected ExitSuccess, got %d", result.ExitCode)
	}

	// Verify binary was copied
	if _, ok := fs.Files["/home/testuser/bin/dashlights"]; !ok {
		t.Error("binary was not installed")
	}
}

func TestGetPathExportTemplate(t *testing.T) {
	tests := []struct {
		shell    ShellType
		contains string
	}{
		{ShellBash, "export PATH="},
		{ShellZsh, "export PATH="},
		{ShellFish, "fish_add_path"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			template := GetPathExportTemplate(tt.shell)
			if !contains(template, tt.contains) {
				t.Errorf("template for %s should contain %q", tt.shell, tt.contains)
			}
			if !contains(template, PathSentinelBegin) {
				t.Error("template should contain begin sentinel")
			}
			if !contains(template, PathSentinelEnd) {
				t.Error("template should contain end sentinel")
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
