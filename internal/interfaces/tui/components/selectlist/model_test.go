package selectlist

import (
	"strings"
	"testing"

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
