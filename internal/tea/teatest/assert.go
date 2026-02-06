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
