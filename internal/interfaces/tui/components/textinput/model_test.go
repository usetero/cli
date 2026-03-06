package textinput

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_ShowsPlaceholderWhenEmpty(t *testing.T) {
	model := New(theme.New(false))
	model.SetPlaceholder("Type a name")
	view := model.View().Content
	if !strings.Contains(view, "Type a name") {
		t.Fatalf("expected placeholder, got %q", view)
	}
}

func TestModel_EditAndBackspace(t *testing.T) {
	model := New(theme.New(false))

	model.Update(tea.KeyPressMsg{Text: "a", Code: 'a'})
	model.Update(tea.KeyPressMsg{Text: "b", Code: 'b'})
	model.Update(tea.KeyPressMsg{Text: "c", Code: 'c'})

	if got := model.Value(); got != "abc" {
		t.Fatalf("expected value abc, got %q", got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := model.Value(); got != "ab" {
		t.Fatalf("expected value ab after backspace, got %q", got)
	}
}
