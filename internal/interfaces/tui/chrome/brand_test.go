package chrome

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestRenderAppWordmark(t *testing.T) {
	got := ansi.Strip(RenderAppWordmark(theme.New(false)))
	if got != "TERO" {
		t.Fatalf("expected wordmark TERO, got %q", got)
	}
}

func TestRenderSlashMotif(t *testing.T) {
	got := ansi.Strip(RenderSlashMotif(theme.New(false), 4))
	if got != strings.Repeat("╱", 4) {
		t.Fatalf("expected slash motif, got %q", got)
	}
}
