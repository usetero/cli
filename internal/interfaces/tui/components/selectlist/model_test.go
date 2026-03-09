package selectlist

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_EmptyView(t *testing.T) {
	model := New(theme.New(false))
	view := model.View().Content
	if !strings.Contains(view, "No options available.") {
		t.Fatalf("expected empty message, got %q", view)
	}
}

func TestModel_CustomEmptyText(t *testing.T) {
	model := New(theme.New(false))
	model.SetEmptyText("No organizations available.")
	view := model.View().Content
	if !strings.Contains(view, "No organizations available.") {
		t.Fatalf("expected custom empty message, got %q", view)
	}
}

func TestModel_NavigateAndSelect(t *testing.T) {
	model := New(theme.New(false))
	model.SetItems([]Item{
		{Title: "One"},
		{Title: "Two"},
	}, 0)

	if got := model.SelectedIndex(); got != 0 {
		t.Fatalf("expected initial index 0, got %d", got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedIndex(); got != 1 {
		t.Fatalf("expected index 1 after down, got %d", got)
	}

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected select command")
	}
	msg := cmd()
	selected, ok := msg.(SelectedMsg)
	if !ok {
		t.Fatalf("expected SelectedMsg, got %T", msg)
	}
	if selected.Index != 1 {
		t.Fatalf("expected selected index 1, got %d", selected.Index)
	}
}

func TestModel_ShortHelpUsesLegacySelectTerminology(t *testing.T) {
	model := New(theme.New(false))

	bindings := model.ShortHelp()
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	if got := bindings[0].Help(); got.Key != "↑/↓" || got.Desc != "select" {
		t.Fatalf("unexpected move help: %+v", got)
	}
	if got := bindings[1].Help(); got.Key != "enter" || got.Desc != "confirm" {
		t.Fatalf("unexpected select help: %+v", got)
	}
	if !key.Matches(tea.KeyPressMsg{Code: tea.KeyDown}, bindings[0]) {
		t.Fatal("expected move help binding to still match down arrow")
	}
}
