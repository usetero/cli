package helpbar

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestModel_Short(t *testing.T) {
	m := New()
	m.SetWidth(80)

	out := m.Short([]key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "continue")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	})

	if !strings.Contains(out, "enter") || !strings.Contains(out, "continue") {
		t.Fatalf("expected enter binding in help output, got %q", out)
	}
	if !strings.Contains(out, "q") || !strings.Contains(out, "quit") {
		t.Fatalf("expected quit binding in help output, got %q", out)
	}
}
