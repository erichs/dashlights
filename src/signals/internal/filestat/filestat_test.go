package filestat

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMatchFile_Extensions(t *testing.T) {
	patterns := DefaultSensitivePatterns()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"sql file", "dump.sql", true},
		{"SQL uppercase", "BACKUP.SQL", true},
		{"sqlite file", "data.sqlite", true},
		{"db file", "app.db", true},
		{"bak file", "old.bak", true},
		{"har file", "request.har", true},
		{"pcap file", "capture.pcap", true},
		{"keychain", "login.keychain", true},
		{"pem file", "cert.pem", true},
		{"pfx file", "cert.pfx", true},
		{"jks file", "keystore.jks", true},
		{"txt file", "notes.txt", false},
		{"go file", "main.go", false},
		{"no extension", "Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := patterns.MatchFile(tt.filename); got != tt.want {
				t.Errorf("MatchFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMatchFile_Prefixes(t *testing.T) {
	patterns := DefaultSensitivePatterns()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"dump prefix", "dump-2024-01-01.tar.gz", true},
		{"DUMP uppercase", "DUMP-latest.tar", true},
		{"backup prefix", "backup-db.tar.gz", true},
		{"nodump prefix", "nodump.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := patterns.MatchFile(tt.filename); got != tt.want {
				t.Errorf("MatchFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMatchFile_Substrings(t *testing.T) {
	patterns := DefaultSensitivePatterns()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"prod in name", "prod-data.csv", true},
		{"PROD uppercase", "PROD_BACKUP.tar", true},
		{"production", "production-dump.sql", true}, // contains "prod"
		{"my-product", "my-product.csv", true},      // contains "prod"
		{"dev file", "dev-data.csv", false},
		{"test file", "test-data.csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := patterns.MatchFile(tt.filename); got != tt.want {
				t.Errorf("MatchFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestScanDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	sensitiveFiles := []string{"dump.sql", "backup-data.db", "prod-config.json"}
	normalFiles := []string{"readme.txt", "main.go", "config.yaml"}

	for _, f := range sensitiveFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range normalFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a subdirectory (should be skipped)
	subDir := filepath.Join(tmpDir, "subdir.sql")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	patterns := DefaultSensitivePatterns()
	results, err := patterns.ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(results))
	}
}

func TestScanDirectory_NonExistent(t *testing.T) {
	patterns := DefaultSensitivePatterns()
	_, err := patterns.ScanDirectory("/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestIsOlderThan(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		modTime   time.Time
		threshold time.Duration
		want      bool
	}{
		{"1 day old, 7 day threshold", now.Add(-24 * time.Hour), 7 * 24 * time.Hour, false},
		{"10 days old, 7 day threshold", now.Add(-10 * 24 * time.Hour), 7 * 24 * time.Hour, true},
		// Use 6 days 23 hours to safely be under threshold (avoids timing issues)
		{"under 7 days", now.Add(-6*24*time.Hour - 23*time.Hour), 7 * 24 * time.Hour, false},
		{"just over 7 days", now.Add(-7*24*time.Hour - time.Hour), 7 * 24 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MatchResult{ModTime: tt.modTime}
			if got := m.IsOlderThan(tt.threshold); got != tt.want {
				t.Errorf("IsOlderThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHotZoneDirectories(t *testing.T) {
	dirs := GetHotZoneDirectories()

	// Should have at least /tmp
	if len(dirs) == 0 {
		t.Error("Expected at least one directory")
	}

	// Check /tmp is included
	hasTemp := false
	for _, d := range dirs {
		if d == "/tmp" {
			hasTemp = true
			break
		}
	}
	if !hasTemp {
		t.Error("Expected /tmp in hot zone directories")
	}
}

// Note: Not using slices.Contains to avoid Go 1.21+ dependency
