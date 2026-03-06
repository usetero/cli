package organizationselect

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

func TestModel_NavigationAndSubmit(t *testing.T) {
	model := New(logging.Scope{}, theme.New(false))
	model.SetOrganizations([]tenancy.Organization{
		{ID: "org_1", Name: "Org One"},
		{ID: "org_2", Name: "Org Two"},
	}, nil)

	if got := model.SelectedOrganizationID(); got != "org_1" {
		t.Fatalf("expected initial org org_1, got %q", got)
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := model.SelectedOrganizationID(); got != "org_2" {
		t.Fatalf("expected org_2 after down, got %q", got)
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
	if selected.OrganizationID != "org_2" {
		t.Fatalf("expected selected org_2, got %q", selected.OrganizationID)
	}
}
