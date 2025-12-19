package ruleoftwo

import (
	"testing"
)

func TestParseWriteInput(t *testing.T) {
	input := map[string]interface{}{
		"file_path": "/path/to/file.txt",
		"content":   "file content",
	}

	result := ParseWriteInput(input)

	if result.FilePath != "/path/to/file.txt" {
		t.Errorf("Expected file_path '/path/to/file.txt', got '%s'", result.FilePath)
	}
	if result.Content != "file content" {
		t.Errorf("Expected content 'file content', got '%s'", result.Content)
	}
}

func TestParseEditInput(t *testing.T) {
	input := map[string]interface{}{
		"file_path":  "/path/to/file.txt",
		"old_string": "old text",
		"new_string": "new text",
	}

	result := ParseEditInput(input)

	if result.FilePath != "/path/to/file.txt" {
		t.Errorf("Expected file_path '/path/to/file.txt', got '%s'", result.FilePath)
	}
	if result.OldString != "old text" {
		t.Errorf("Expected old_string 'old text', got '%s'", result.OldString)
	}
	if result.NewString != "new text" {
		t.Errorf("Expected new_string 'new text', got '%s'", result.NewString)
	}
}

func TestParseBashInput(t *testing.T) {
	input := map[string]interface{}{
		"command":     "ls -la",
		"description": "List files",
		"timeout":     float64(30000), // JSON numbers come as float64
	}

	result := ParseBashInput(input)

	if result.Command != "ls -la" {
		t.Errorf("Expected command 'ls -la', got '%s'", result.Command)
	}
	if result.Description != "List files" {
		t.Errorf("Expected description 'List files', got '%s'", result.Description)
	}
	if result.Timeout != 30000 {
		t.Errorf("Expected timeout 30000, got %d", result.Timeout)
	}
}

func TestParseReadInput(t *testing.T) {
	input := map[string]interface{}{
		"file_path": "/path/to/file.txt",
		"offset":    float64(100),
		"limit":     float64(50),
	}

	result := ParseReadInput(input)

	if result.FilePath != "/path/to/file.txt" {
		t.Errorf("Expected file_path '/path/to/file.txt', got '%s'", result.FilePath)
	}
	if result.Offset != 100 {
		t.Errorf("Expected offset 100, got %d", result.Offset)
	}
	if result.Limit != 50 {
		t.Errorf("Expected limit 50, got %d", result.Limit)
	}
}

func TestParseWebFetchInput(t *testing.T) {
	input := map[string]interface{}{
		"url":    "https://example.com",
		"prompt": "Extract the content",
	}

	result := ParseWebFetchInput(input)

	if result.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", result.URL)
	}
	if result.Prompt != "Extract the content" {
		t.Errorf("Expected prompt 'Extract the content', got '%s'", result.Prompt)
	}
}

func TestParseWebSearchInput(t *testing.T) {
	input := map[string]interface{}{
		"query": "golang testing",
	}

	result := ParseWebSearchInput(input)

	if result.Query != "golang testing" {
		t.Errorf("Expected query 'golang testing', got '%s'", result.Query)
	}
}

func TestParseGrepInput(t *testing.T) {
	input := map[string]interface{}{
		"pattern": "TODO",
		"path":    "/path/to/search",
		"glob":    "*.go",
	}

	result := ParseGrepInput(input)

	if result.Pattern != "TODO" {
		t.Errorf("Expected pattern 'TODO', got '%s'", result.Pattern)
	}
	if result.Path != "/path/to/search" {
		t.Errorf("Expected path '/path/to/search', got '%s'", result.Path)
	}
	if result.Glob != "*.go" {
		t.Errorf("Expected glob '*.go', got '%s'", result.Glob)
	}
}

func TestParseGlobInput(t *testing.T) {
	input := map[string]interface{}{
		"pattern": "**/*.go",
		"path":    "/path/to/search",
	}

	result := ParseGlobInput(input)

	if result.Pattern != "**/*.go" {
		t.Errorf("Expected pattern '**/*.go', got '%s'", result.Pattern)
	}
	if result.Path != "/path/to/search" {
		t.Errorf("Expected path '/path/to/search', got '%s'", result.Path)
	}
}

func TestGetStringField_MissingKey(t *testing.T) {
	input := map[string]interface{}{}
	result := getStringField(input, "missing")
	if result != "" {
		t.Errorf("Expected empty string for missing key, got '%s'", result)
	}
}

func TestGetStringField_NonStringValue(t *testing.T) {
	input := map[string]interface{}{
		"number": 42,
	}
	result := getStringField(input, "number")
	if result != "" {
		t.Errorf("Expected empty string for non-string value, got '%s'", result)
	}
}

func TestGetIntField_MissingKey(t *testing.T) {
	input := map[string]interface{}{}
	result := getIntField(input, "missing")
	if result != 0 {
		t.Errorf("Expected 0 for missing key, got %d", result)
	}
}

func TestGetIntField_IntValue(t *testing.T) {
	input := map[string]interface{}{
		"value": 42,
	}
	result := getIntField(input, "value")
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestGetIntField_Int64Value(t *testing.T) {
	input := map[string]interface{}{
		"value": int64(42),
	}
	result := getIntField(input, "value")
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestGetIntField_Float64Value(t *testing.T) {
	input := map[string]interface{}{
		"value": float64(42.9),
	}
	result := getIntField(input, "value")
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}
