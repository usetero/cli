package chrome

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestRenderCard_IncludesTitleAndBody(t *testing.T) {
	t.Parallel()

	got := RenderCard(theme.New(false), "Title", "Body")
	if !strings.Contains(got, "Title") {
		t.Fatalf("expected title in card, got %q", got)
	}
	if !strings.Contains(got, "Body") {
		t.Fatalf("expected body in card, got %q", got)
	}
	if !strings.Contains(got, "╭") {
		t.Fatalf("expected bordered card chrome, got %q", got)
	}
}

func TestRenderErrorCard_IncludesTitleAndBody(t *testing.T) {
	t.Parallel()

	got := RenderErrorCard(theme.New(false), "Error", "Details")
	if !strings.Contains(got, "Error") {
		t.Fatalf("expected error title in card, got %q", got)
	}
	if !strings.Contains(got, "Details") {
		t.Fatalf("expected error body in card, got %q", got)
	}
	if !strings.Contains(got, "╭") {
		t.Fatalf("expected bordered card chrome, got %q", got)
	}
}
