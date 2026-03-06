package tenancyflow

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	accountcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/create"
	accountselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/account/select"
	organizationcreate "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/create"
	organizationselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/organization/select"
	workspaceselect "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding/tenancy/workspace/select"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	"github.com/usetero/cli/internal/runtime/onboarding"
)

func newModel() *Model {
	appTheme := theme.New(false)
	return New(
		organizationselect.New(logging.Scope{}, appTheme),
		organizationcreate.New(logging.Scope{}, appTheme),
		accountselect.New(logging.Scope{}, appTheme),
		accountcreate.New(logging.Scope{}, appTheme),
		workspaceselect.New(logging.Scope{}, appTheme),
		appTheme,
	)
}

func TestModel_ApplyStateRoutes(t *testing.T) {
	model := newModel()
	if !model.ApplyState(onboarding.State{NextStep: onboarding.StepOrganizationCreate}) {
		t.Fatal("expected state to be handled")
	}
	if !strings.Contains(model.View().Content, "Create your organization:") {
		t.Fatalf("expected organization create view, got %q", model.View().Content)
	}
}

func TestModel_WorkspaceSelectionLift(t *testing.T) {
	model := newModel()
	if !model.ApplyState(onboarding.State{
		NextStep: onboarding.StepWorkspaceSelect,
		Workspaces: []tenancy.Workspace{
			{ID: "ws_1", Name: "One"},
			{ID: "ws_2", Name: "Two"},
		},
	}) {
		t.Fatal("expected workspace select state to be handled")
	}

	model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected select command")
	}
	msg := cmd()
	selected, ok := msg.(WorkspaceSelectedMsg)
	if !ok {
		t.Fatalf("expected WorkspaceSelectedMsg, got %T", msg)
	}
	if selected.WorkspaceID != "ws_2" {
		t.Fatalf("expected workspace ws_2, got %q", selected.WorkspaceID)
	}
}
