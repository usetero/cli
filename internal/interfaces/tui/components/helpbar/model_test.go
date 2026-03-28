package helpbar

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func TestModel_Short(t *testing.T) {
	m := New(theme.New(false))
	m.SetSize(80, 1)
	m.SetBindings([]key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	})

	out := ansi.Strip(m.View().Content)

	if !strings.Contains(out, "enter") || !strings.Contains(out, "confirm") {
		t.Fatalf("expected enter binding in help output, got %q", out)
	}
	if !strings.Contains(out, "q") || !strings.Contains(out, "quit") {
		t.Fatalf("expected quit binding in help output, got %q", out)
	}
	if strings.Contains(out, "╱╱") {
		t.Fatalf("did not expect decorative motif in help output, got %q", out)
	}
}
