package role

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_Navigation(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))

	if got := model.SelectedRole(); got != preferences.RoleEngineer {
		t.Fatalf("expected initial role %q, got %q", preferences.RoleEngineer, got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedRole(); got != preferences.RolePlatform {
		t.Fatalf("expected role %q after down, got %q", preferences.RolePlatform, got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedRole(); got != preferences.RolePlatform {
		t.Fatalf("expected clamped role %q, got %q", preferences.RolePlatform, got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if got := model.SelectedRole(); got != preferences.RoleEngineer {
		t.Fatalf("expected role %q after up, got %q", preferences.RoleEngineer, got)
	}
}

func TestModel_EnterEmitsSubmittedMessage(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected submit command")
	}

	msg := cmd()
	submitted, ok := msg.(SubmittedMsg)
	if !ok {
		t.Fatalf("expected SubmittedMsg, got %T", msg)
	}
	if submitted.Role != preferences.RoleEngineer {
		t.Fatalf("expected submitted role %q, got %q", preferences.RoleEngineer, submitted.Role)
	}
}
