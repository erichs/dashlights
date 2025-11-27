package main

import (
	"io"
	"sort"
)

// emojiAliasMap maps human-readable emoji labels to their UTF-8 hex codes
var emojiAliasMap = map[string]string{
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

// resolveEmojiAlias attempts to resolve an emoji alias to its hex code.
// If the input is already a hex code or not a recognized alias, returns the input unchanged.
func resolveEmojiAlias(input string) string {
	if hexCode, exists := emojiAliasMap[input]; exists {
		return hexCode
	}
	return input
}

// displayEmojiList outputs the supported emoji aliases and their corresponding glyphs
func displayEmojiList(w io.Writer) {
	type emojiEntry struct {
		label   string
		hexCode string
		glyph   string
	}

	entries := make([]emojiEntry, 0, len(emojiAliasMap))
	for label, hexCode := range emojiAliasMap {
		glyph, err := utf8HexToString(hexCode)
		if err != nil {
			glyph = "?"
		}
		entries = append(entries, emojiEntry{
			label:   label,
			hexCode: hexCode,
			glyph:   glyph,
		})
	}

	// Sort by label for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].label < entries[j].label
	})

	flexPrintln(w, "Supported emoji aliases:")
	flexPrintf(w, "%-20s %-10s %s\n", "LABEL", "HEX CODE", "EMOJI")
	flexPrintln(w, "--------------------------------------------")
	for _, entry := range entries {
		flexPrintf(w, "%-20s %-10s %s\n", entry.label, entry.hexCode, entry.glyph)
	}
}

