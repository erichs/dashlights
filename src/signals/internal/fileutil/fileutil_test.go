package fileutil

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileLimited(t *testing.T) {
	t.Run("reads file within limit", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		content := []byte("hello world")
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		data, err := ReadFileLimited(path, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "hello world" {
			t.Errorf("got %q, want %q", string(data), "hello world")
		}
	})

	t.Run("rejects file exceeding limit", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		content := []byte("hello world")
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		_, err := ReadFileLimited(path, 5)
		if err != ErrFileTooLarge {
			t.Errorf("got error %v, want ErrFileTooLarge", err)
		}
	})

	t.Run("rejects zero maxBytes", func(t *testing.T) {
		_, err := ReadFileLimited("/any/path", 0)
		if err != ErrFileTooLarge {
			t.Errorf("got error %v, want ErrFileTooLarge", err)
		}
	})

	t.Run("rejects negative maxBytes", func(t *testing.T) {
		_, err := ReadFileLimited("/any/path", -1)
		if err != ErrFileTooLarge {
			t.Errorf("got error %v, want ErrFileTooLarge", err)
		}
	})

	t.Run("rejects maxBytes exceeding MaxInt on 32-bit", func(t *testing.T) {
		// On 32-bit systems, math.MaxInt is 2^31-1, so math.MaxInt32+1 exceeds it.
		// On 64-bit systems, this value is well within math.MaxInt, so we skip.
		if math.MaxInt > math.MaxInt32 {
			t.Skip("skipping on 64-bit systems where MaxInt == MaxInt64")
		}
		_, err := ReadFileLimited("/any/path", math.MaxInt32+1)
		if err != ErrFileTooLarge {
			t.Errorf("got error %v, want ErrFileTooLarge", err)
		}
	})

	t.Run("rejects non-regular file", func(t *testing.T) {
		dir := t.TempDir()
		// dir itself is not a regular file
		_, err := ReadFileLimited(dir, 100)
		if err != ErrNotRegular {
			t.Errorf("got error %v, want ErrNotRegular", err)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := ReadFileLimited("/nonexistent/path/file.txt", 100)
		if err == nil {
			t.Error("expected error for non-existent file")
		}
		if err == ErrFileTooLarge || err == ErrNotRegular {
			t.Errorf("got %v, want os path error", err)
		}
	})

	t.Run("reads file at exact limit", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		content := []byte("12345")
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		data, err := ReadFileLimited(path, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "12345" {
			t.Errorf("got %q, want %q", string(data), "12345")
		}
	})

	t.Run("follows symlink to regular file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "target.txt")
		link := filepath.Join(dir, "link.txt")

		if err := os.WriteFile(target, []byte("content"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, link); err != nil {
			t.Skip("symlinks not supported")
		}

		data, err := ReadFileLimited(link, 100)
		if err != nil {
			t.Fatalf("symlink to regular file should succeed: %v", err)
		}
		if string(data) != "content" {
			t.Errorf("got %q, want %q", string(data), "content")
		}
	})

	t.Run("rejects symlink to directory", func(t *testing.T) {
		dir := t.TempDir()
		subdir := filepath.Join(dir, "subdir")
		link := filepath.Join(dir, "link")

		if err := os.Mkdir(subdir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(subdir, link); err != nil {
			t.Skip("symlinks not supported")
		}

		_, err := ReadFileLimited(link, 100)
		if err != ErrNotRegular {
			t.Errorf("symlink to directory: got %v, want ErrNotRegular", err)
		}
	})
}

func TestReadFileLimitedString(t *testing.T) {
	t.Run("returns string content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		content := []byte("hello world")
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		data, err := ReadFileLimitedString(path, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data != "hello world" {
			t.Errorf("got %q, want %q", data, "hello world")
		}
	})

	t.Run("propagates errors", func(t *testing.T) {
		_, err := ReadFileLimitedString("/nonexistent/path", 100)
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}
