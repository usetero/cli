package stepkit

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/styles"
)

func TestCreateInputShortHelp(t *testing.T) {
	t.Parallel()

	base := []key.Binding{key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next"))}
	got := CreateInputShortHelp(false, base)
	if len(got) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(got))
	}
	if got[1].Help().Key != "enter" {
		t.Fatalf("expected enter binding, got %q", got[1].Help().Key)
	}

	if got := CreateInputShortHelp(true, base); got != nil {
		t.Fatalf("expected nil while creating, got %#v", got)
	}
}

func TestParseCreateSubmit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	if name, ok := ParseCreateSubmit(msg, false, "Acme"); !ok || name != "Acme" {
		t.Fatalf("expected valid submit, got (%q, %v)", name, ok)
	}
	if _, ok := ParseCreateSubmit(msg, true, "Acme"); ok {
		t.Fatalf("expected creating submit to be blocked")
	}
	if _, ok := ParseCreateSubmit(msg, false, ""); ok {
		t.Fatalf("expected empty submit to be blocked")
	}
	if _, ok := ParseCreateSubmit(tea.KeyPressMsg{Code: tea.KeyTab}, false, "Acme"); ok {
		t.Fatalf("expected non-enter submit to be blocked")
	}
}

func TestRenderCreateForm(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(false)
	view := RenderCreateForm(theme, "Create x", "subtitle", "input-view", false, nil, "fallback")
	if !strings.Contains(view, "Create x") || !strings.Contains(view, "input-view") {
		t.Fatalf("expected title and input in view, got %q", view)
	}

	creatingView := RenderCreateForm(theme, "Create x", "subtitle", "input-view", true, nil, "fallback")
	if !strings.Contains(creatingView, "Creating...") {
		t.Fatalf("expected creating status in view, got %q", creatingView)
	}

	errorView := RenderCreateForm(theme, "Create x", "subtitle", "input-view", false, errors.New("boom"), "fallback")
	if !strings.Contains(errorView, "boom") {
		t.Fatalf("expected error text in view, got %q", errorView)
	}
}
