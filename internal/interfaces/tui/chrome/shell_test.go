package chrome

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestRender_IncludesShellSections(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "╱╱ TERO",
		Body:   BodySlot{Content: "Body content"},
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
		Body:   BodySlot{Content: "Body content"},
	}, Viewport{}).Content
	if strings.Contains(view, "q quit") {
		t.Fatalf("did not expect help content, got %q", view)
	}
}

func TestRender_UsesViewportHeight(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "hdr",
		Body:   BodySlot{Content: "body"},
		Footer: "help",
	}, Viewport{Width: 80, Height: 20}).Content
	view = strings.TrimRight(view, "\n")
	if got := strings.Count(view, "\n") + 1; got != 20 {
		t.Fatalf("expected 20 rendered lines, got %d", got)
	}
}

func TestRender_BottomAlignsIntrinsicBody(t *testing.T) {
	view := Render(theme.New(false), Slots{
		Header: "hdr",
		Body: BodySlot{
			Content: "body",
			Layout: BodyLayout{
				WidthMode:     WidthIntrinsic,
				HeightMode:    HeightIntrinsic,
				VerticalAlign: AlignBottom,
				MaxWidth:      40,
			},
		},
		Footer: "help",
	}, Viewport{Width: 40, Height: 10}).Content

	lines := strings.Split(view, "\n")
	bodyLine := -1
	helpLine := -1
	for i, line := range lines {
		if strings.Contains(line, "body") {
			bodyLine = i
		}
		if strings.Contains(line, "help") {
			helpLine = i
		}
	}
	if bodyLine == -1 || helpLine == -1 {
		t.Fatalf("expected body and help lines, got %q", view)
	}
	if bodyLine >= helpLine {
		t.Fatalf("expected body above help, got body=%d help=%d in %q", bodyLine, helpLine, view)
	}
	if helpLine-bodyLine > 2 {
		t.Fatalf("expected body docked near footer, got body=%d help=%d in %q", bodyLine, helpLine, view)
	}
}
