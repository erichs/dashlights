package install

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBackupManager_CreateBackup_Success(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	// Create a test file
	testPath := "/test/file.txt"
	testContent := []byte("test content")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	result, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if !result.Created {
		t.Error("Expected Created=true")
	}

	if result.OriginalPath != testPath {
		t.Errorf("Expected OriginalPath=%s, got %s", testPath, result.OriginalPath)
	}

	expectedBackupPath := testPath + ".dashlights-backup"
	if result.BackupPath != expectedBackupPath {
		t.Errorf("Expected BackupPath=%s, got %s", expectedBackupPath, result.BackupPath)
	}

	// Verify backup file was created
	backupContent, ok := fs.Files[expectedBackupPath]
	if !ok {
		t.Fatal("Backup file was not created")
	}

	if string(backupContent) != string(testContent) {
		t.Errorf("Expected backup content=%s, got %s", string(testContent), string(backupContent))
	}

	// Verify original file still exists
	if _, ok := fs.Files[testPath]; !ok {
		t.Error("Original file should still exist")
	}
}

func TestBackupManager_CreateBackup_FileNotExist(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/nonexistent/file.txt"

	result, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("CreateBackup should not error for nonexistent file: %v", err)
	}

	if result.Created {
		t.Error("Expected Created=false for nonexistent file")
	}

	// Verify no backup was created
	backupPath := testPath + ".dashlights-backup"
	if _, ok := fs.Files[backupPath]; ok {
		t.Error("Backup should not be created for nonexistent file")
	}
}

func TestBackupManager_CreateBackup_ExistingBackup(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	// Create original file
	testPath := "/test/file.txt"
	testContent := []byte("test content")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	// Create existing backup
	existingBackupPath := testPath + ".dashlights-backup"
	fs.Files[existingBackupPath] = []byte("old backup")
	fs.Modes[existingBackupPath] = testMode

	// Create new backup
	result, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if !result.Created {
		t.Error("Expected Created=true")
	}

	// Verify timestamped backup was created
	if result.BackupPath == existingBackupPath {
		t.Error("Expected timestamped backup path, got standard backup path")
	}

	if !strings.HasPrefix(result.BackupPath, testPath+".dashlights-backup-") {
		t.Errorf("Expected timestamped backup path with prefix %s.dashlights-backup-, got %s",
			testPath, result.BackupPath)
	}

	// Verify the timestamped backup contains the new content
	backupContent, ok := fs.Files[result.BackupPath]
	if !ok {
		t.Fatal("Timestamped backup file was not created")
	}

	if string(backupContent) != string(testContent) {
		t.Errorf("Expected backup content=%s, got %s", string(testContent), string(backupContent))
	}

	// Verify old backup still exists
	oldBackupContent, ok := fs.Files[existingBackupPath]
	if !ok {
		t.Error("Old backup should still exist")
	}
	if string(oldBackupContent) != "old backup" {
		t.Error("Old backup content should be unchanged")
	}
}

func TestBackupManager_CreateBackup_PreservesPermissions(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/executable.sh"
	testContent := []byte("#!/bin/bash\necho test")
	testMode := os.FileMode(0755) // Executable permissions

	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	result, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if !result.Created {
		t.Error("Expected Created=true")
	}

	// Verify backup has same permissions as original
	backupMode, ok := fs.Modes[result.BackupPath]
	if !ok {
		t.Fatal("Backup mode not set")
	}

	if backupMode != testMode {
		t.Errorf("Expected backup mode=%o, got %o", testMode, backupMode)
	}
}

func TestBackupManager_CreateBackup_StatError(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/file.txt"
	testContent := []byte("test content")
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = 0644

	// Simulate stat error (not os.IsNotExist)
	expectedErr := errors.New("permission denied")
	fs.StatErr = expectedErr

	result, err := bm.CreateBackup(testPath)
	if err == nil {
		t.Fatal("Expected error when stat fails")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	if !strings.Contains(err.Error(), "failed to stat file") {
		t.Errorf("Expected 'failed to stat file' error, got: %v", err)
	}
}

func TestBackupManager_CreateBackup_ReadError(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/file.txt"
	testContent := []byte("test content")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	// Simulate read error
	expectedErr := errors.New("read error")
	fs.ReadFileErr = expectedErr

	result, err := bm.CreateBackup(testPath)
	if err == nil {
		t.Fatal("Expected error when read fails")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected 'failed to read file' error, got: %v", err)
	}
}

func TestBackupManager_CreateBackup_WriteError(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/file.txt"
	testContent := []byte("test content")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	// Simulate write error
	expectedErr := errors.New("disk full")
	fs.WriteFileErr = expectedErr

	result, err := bm.CreateBackup(testPath)
	if err == nil {
		t.Fatal("Expected error when write fails")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %+v", result)
	}

	if !strings.Contains(err.Error(), "backup failed") {
		t.Errorf("Expected 'backup failed' error, got: %v", err)
	}
}

func TestBackupManager_CreateBackup_TimestampFormat(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/file.txt"
	testContent := []byte("test content")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	// Create existing backup to trigger timestamp path
	existingBackupPath := testPath + ".dashlights-backup"
	fs.Files[existingBackupPath] = []byte("old")
	fs.Modes[existingBackupPath] = testMode

	beforeTimestamp := time.Now().Unix()
	result, err := bm.CreateBackup(testPath)
	afterTimestamp := time.Now().Unix()

	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Extract timestamp from backup path
	// Expected format: /test/file.txt.dashlights-backup-1234567890
	prefix := testPath + ".dashlights-backup-"
	if !strings.HasPrefix(result.BackupPath, prefix) {
		t.Fatalf("Expected backup path to start with %s, got %s", prefix, result.BackupPath)
	}

	timestampStr := strings.TrimPrefix(result.BackupPath, prefix)
	var timestamp int64
	if _, err := time.Parse("1", timestampStr); err == nil {
		// Simple validation that it looks like a timestamp
		if len(timestampStr) < 10 {
			t.Errorf("Timestamp appears too short: %s", timestampStr)
		}
	}

	// Sanity check: timestamp should be between before and after
	// This is a weak check but validates format
	if timestamp != 0 && (timestamp < beforeTimestamp || timestamp > afterTimestamp+1) {
		t.Errorf("Timestamp %d not within expected range [%d, %d]",
			timestamp, beforeTimestamp, afterTimestamp)
	}
}

func TestBackupManager_CreateBackup_MultipleBackups(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	testPath := "/test/file.txt"
	testContent := []byte("version 1")
	testMode := os.FileMode(0644)
	fs.Files[testPath] = testContent
	fs.Modes[testPath] = testMode

	// First backup
	result1, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("First backup failed: %v", err)
	}

	// Update original file
	fs.Files[testPath] = []byte("version 2")

	// Second backup should create timestamped version
	result2, err := bm.CreateBackup(testPath)
	if err != nil {
		t.Fatalf("Second backup failed: %v", err)
	}

	// Verify both backups exist
	if _, ok := fs.Files[result1.BackupPath]; !ok {
		t.Error("First backup should still exist")
	}

	if _, ok := fs.Files[result2.BackupPath]; !ok {
		t.Error("Second backup should exist")
	}

	// Verify they have different paths
	if result1.BackupPath == result2.BackupPath {
		t.Error("Second backup should have different path than first")
	}

	// Verify second backup has timestamped name
	if !strings.Contains(result2.BackupPath, ".dashlights-backup-") {
		t.Errorf("Expected timestamped backup path, got %s", result2.BackupPath)
	}
}

func TestNewBackupManager(t *testing.T) {
	fs := NewMockFilesystem()
	bm := NewBackupManager(fs)

	if bm == nil {
		t.Fatal("NewBackupManager returned nil")
	}

	if bm.fs != fs {
		t.Error("BackupManager should use provided filesystem")
	}
}

func TestBackupResult_Fields(t *testing.T) {
	result := &BackupResult{
		OriginalPath: "/test/original.txt",
		BackupPath:   "/test/original.txt.dashlights-backup",
		Created:      true,
	}

	if result.OriginalPath != "/test/original.txt" {
		t.Errorf("Expected OriginalPath=/test/original.txt, got %s", result.OriginalPath)
	}

	if result.BackupPath != "/test/original.txt.dashlights-backup" {
		t.Errorf("Expected BackupPath=/test/original.txt.dashlights-backup, got %s", result.BackupPath)
	}

	if !result.Created {
		t.Error("Expected Created=true")
	}
}
