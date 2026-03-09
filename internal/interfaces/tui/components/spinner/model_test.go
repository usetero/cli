package spinner

import (
	"testing"

	bubblespinner "charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_InitReturnsTick(t *testing.T) {
	model := New(theme.New(false))
	if model.Init() == nil {
		t.Fatal("expected initial spinner tick")
	}
}

func TestModel_UpdateIgnoresNonTick(t *testing.T) {
	model := New(theme.New(false))
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil command for non-tick")
	}
}

func TestModel_UpdateTickReturnsCmd(t *testing.T) {
	model := New(theme.New(false))
	_, cmd := model.Update(bubblespinner.TickMsg{})
	if cmd == nil {
		t.Fatal("expected tick command")
	}
}
