package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayColorList(t *testing.T) {
	var b bytes.Buffer
	listLen := len(colorMap)
	displayColorList(&b)
	commaCount := strings.Count(b.String(), ",")
	if commaCount != listLen-1 {
		t.Errorf("Expected %d commas in colorlist, got %d", listLen-1, commaCount)
	}
	// color attributes are listed in UPPER CASE...
	if !strings.Contains(b.String(), "BGWHITE") {
		t.Error("Expected to see string 'BGWHITE' in: ", b.String())
	}
}
