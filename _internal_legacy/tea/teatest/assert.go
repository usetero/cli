package teatest

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// AssertMaxWidth fails if any line in output exceeds maxWidth.
func AssertMaxWidth(t *testing.T, maxWidth int, output string) {
	t.Helper()
	for i, line := range strings.Split(output, "\n") {
		w := ansi.StringWidth(line)
		if w > maxWidth {
			t.Errorf("line %d: width %d > max %d: %q", i, w, maxWidth, line)
		}
	}
}

// AssertExactWidth fails if the widest line in output doesn't equal expected.
func AssertExactWidth(t *testing.T, expected int, output string) {
	t.Helper()
	var widest int
	for _, line := range strings.Split(output, "\n") {
		widest = max(widest, ansi.StringWidth(line))
	}
	if widest != expected {
		t.Errorf("expected output width == %d, got %d", expected, widest)
	}
}

// AssertNoRawEscapes fails if any line contains a broken/visible escape sequence.
// A raw escape looks like "38;2;110;231;183m" — ANSI parameters without the leading ESC[.
func AssertNoRawEscapes(t *testing.T, output string) {
	t.Helper()
	for i, line := range strings.Split(output, "\n") {
		// After stripping valid ANSI sequences, no escape-like fragments should remain
		stripped := ansi.Strip(line)
		for j, r := range stripped {
			if r == '\x1b' {
				t.Errorf("line %d col %d: contains raw ESC byte in stripped output: %q", i, j, stripped)
				break
			}
		}
		// Check for orphaned CSI parameters (e.g. "38;2;110;231;183m")
		if strings.Contains(stripped, ";") && strings.ContainsAny(stripped, "m") {
			// Heuristic: if stripping ANSI left behind something like "38;2;...m", it's broken
			for _, part := range strings.Split(stripped, " ") {
				if looksLikeBrokenCSI(part) {
					t.Errorf("line %d: contains raw ANSI parameters in stripped output: %q (full: %q)", i, part, stripped)
					break
				}
			}
		}
	}
}

// looksLikeBrokenCSI checks if a string looks like orphaned CSI parameters.
// e.g. "38;2;110;231;183m>" or "38;2;110;231;183m"
func looksLikeBrokenCSI(s string) bool {
	if len(s) < 3 {
		return false
	}
	// Must end with a letter (CSI final byte) and contain semicolons
	if !strings.Contains(s, ";") {
		return false
	}
	// Check if it looks like "digits;digits;...letter"
	semiCount := strings.Count(s, ";")
	if semiCount < 2 {
		return false
	}
	// High confidence: 3+ semicolons with 'm' somewhere = broken color code
	if semiCount >= 3 && strings.Contains(s, "m") {
		return true
	}
	return false
}
