package accountselect

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_NavigationAndSubmit(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetAccounts([]tenancy.Account{
		{ID: "acct_1", Name: "Account One"},
		{ID: "acct_2", Name: "Account Two"},
	}, nil)

	if got := model.SelectedAccountID(); got != "acct_1" {
		t.Fatalf("expected initial account acct_1, got %q", got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedAccountID(); got != "acct_2" {
		t.Fatalf("expected acct_2 after down, got %q", got)
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
	if selected.AccountID != "acct_2" {
		t.Fatalf("expected selected acct_2, got %q", selected.AccountID)
	}
}
