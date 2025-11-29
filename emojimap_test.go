package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveEmojiAlias(t *testing.T) {
	// Test valid emoji aliases
	testCases := []struct {
		alias    string
		expected string
	}{
		{"WRENCH", "1F527"},
		{"PAPERCLIP", "1F4CE"},
		{"CHECKMARK", "2705"},
		{"CROSSMARK", "274C"},
		{"LIGHTBULB", "1F4A1"},
		{"LINK", "1F517"},
		{"SHIELD", "1F6E1"},
		{"KEY", "1F511"},
		{"LOCK", "1F512"},
		{"MAGNIFYINGGLASS", "1F50D"},
	}

	for _, tc := range testCases {
		result := resolveEmojiAlias(tc.alias)
		if result != tc.expected {
			t.Errorf("resolveEmojiAlias(%s) = %s; want %s", tc.alias, result, tc.expected)
		}
	}

	// Test that hex codes pass through unchanged
	hexCodes := []string{"1F527", "0021", "2112", "1F4A9"}
	for _, hexCode := range hexCodes {
		result := resolveEmojiAlias(hexCode)
		if result != hexCode {
			t.Errorf("resolveEmojiAlias(%s) = %s; want %s (should pass through)", hexCode, result, hexCode)
		}
	}

	// Test that unknown aliases pass through unchanged
	unknownAliases := []string{"NOTANALIAS", "FOOBAR", "UNKNOWN"}
	for _, alias := range unknownAliases {
		result := resolveEmojiAlias(alias)
		if result != alias {
			t.Errorf("resolveEmojiAlias(%s) = %s; want %s (should pass through)", alias, result, alias)
		}
	}
}

func TestDisplayEmojiList(t *testing.T) {
	var b bytes.Buffer
	displayEmojiList(&b)
	output := b.String()

	// Check that the output contains expected headers
	if !strings.Contains(output, "Supported emoji aliases:") {
		t.Error("Expected output to contain 'Supported emoji aliases:'")
	}
	if !strings.Contains(output, "LABEL") {
		t.Error("Expected output to contain 'LABEL' header")
	}
	if !strings.Contains(output, "HEX CODE") {
		t.Error("Expected output to contain 'HEX CODE' header")
	}
	if !strings.Contains(output, "EMOJI") {
		t.Error("Expected output to contain 'EMOJI' header")
	}

	// Check that some known emoji aliases are in the output
	expectedAliases := []string{"WRENCH", "PAPERCLIP", "CHECKMARK", "LIGHTBULB"}
	for _, alias := range expectedAliases {
		if !strings.Contains(output, alias) {
			t.Errorf("Expected output to contain alias '%s'", alias)
		}
	}

	// Check that corresponding hex codes are in the output
	expectedHexCodes := []string{"1F527", "1F4CE", "2705", "1F4A1"}
	for _, hexCode := range expectedHexCodes {
		if !strings.Contains(output, hexCode) {
			t.Errorf("Expected output to contain hex code '%s'", hexCode)
		}
	}
}

func TestEmojiAliasMapCompleteness(t *testing.T) {
	// Verify all required emoji aliases are present
	requiredAliases := map[string]string{
		"CRYSTALBALL":     "1F52E",
		"SHOPPINGCART":    "1F6D2",
		"NOENTRY":         "26D4",
		"NOENTRYSIGN":     "1F6AB",
		"CROSSMARK":       "274C",
		"CHECKMARK":       "2705",
		"QUESTIONMARK":    "2753",
		"EXCLAMATIONMARK": "2757",
		"ANTENNAWITHBARS": "1F4F6",
		"SQUAREDSOS":      "1F198",
		"LINK":            "1F517",
		"WRENCH":          "1F527",
		"SHIELD":          "1F6E1",
		"HAMMERANDWRENCH": "1F6E0",
		"KEY":             "1F511",
		"LOCK":            "1F512",
		"PAPERCLIP":       "1F4CE",
		"PUSHPIN":         "1F4CC",
		"FILEFOLDER":      "1F4C1",
		"SCROLL":          "1F4DC",
		"NOTEBOOK":        "1F4D3",
		"LIGHTBULB":       "1F4A1",
		"MAGNIFYINGGLASS": "1F50D",
	}

	for alias, expectedHex := range requiredAliases {
		actualHex, exists := emojiAliasMap[alias]
		if !exists {
			t.Errorf("Required alias '%s' is missing from emojiAliasMap", alias)
		} else if actualHex != expectedHex {
			t.Errorf("Alias '%s' has hex code '%s'; want '%s'", alias, actualHex, expectedHex)
		}
	}

	// Verify the map has exactly the expected number of entries
	if len(emojiAliasMap) != len(requiredAliases) {
		t.Errorf("emojiAliasMap has %d entries; want %d", len(emojiAliasMap), len(requiredAliases))
	}
}
