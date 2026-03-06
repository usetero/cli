package workspaceselect

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_NavigationAndSubmit(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetWorkspaces([]tenancy.Workspace{
		{ID: "ws_1", Name: "Workspace One"},
		{ID: "ws_2", Name: "Workspace Two"},
	}, nil)

	if got := model.SelectedWorkspaceID(); got != "ws_1" {
		t.Fatalf("expected initial workspace ws_1, got %q", got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedWorkspaceID(); got != "ws_2" {
		t.Fatalf("expected ws_2 after down, got %q", got)
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
	if selected.WorkspaceID != "ws_2" {
		t.Fatalf("expected selected ws_2, got %q", selected.WorkspaceID)
	}
}
