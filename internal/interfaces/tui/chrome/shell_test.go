package chrome

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestRender_IncludesShellSections(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "╱╱ TERO",
		Body:   "Body content",
		Footer: "q quit",
	}, Viewport{}).Content

	if !strings.Contains(view, "TERO") {
		t.Fatalf("expected title, got %q", view)
	}
	if !strings.Contains(view, "╱╱") {
		t.Fatalf("expected legacy-inspired slash motif header, got %q", view)
	}
	if !strings.Contains(view, "Body content") {
		t.Fatalf("expected body content, got %q", view)
	}
	if !strings.Contains(view, "q quit") {
		t.Fatalf("expected help content, got %q", view)
	}
}

func TestRender_OmitsHelpWhenEmpty(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "╱╱ TERO",
		Body:   "Body content",
	}, Viewport{}).Content
	if strings.Contains(view, "q quit") {
		t.Fatalf("did not expect help content, got %q", view)
	}
}

func TestRender_UsesViewportHeight(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "hdr",
		Body:   "body",
		Footer: "help",
	}, Viewport{Width: 80, Height: 20}).Content
	if got := strings.Count(view, "\n") + 1; got != 20 {
		t.Fatalf("expected 20 rendered lines, got %d", got)
	}
}
