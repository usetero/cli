package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

func TestApplyAuthenticated(t *testing.T) {
	t.Parallel()

	user := auth.User{ID: "u-1"}
	state, next := ApplyAuthenticated(State{}, user)
	if next != GateRoleSelect {
		t.Fatalf("next gate = %q, want %q", next, GateRoleSelect)
	}
	if state.User == nil || state.User.ID != user.ID {
		t.Fatalf("user not applied: %#v", state.User)
	}
}

func TestApplyDatadogTransitions(t *testing.T) {
	t.Parallel()

	state, next := ApplyDatadogRegionSelected(State{}, "US1")
	if next != GateDatadogAPIKey {
		t.Fatalf("next gate = %q, want %q", next, GateDatadogAPIKey)
	}
	if state.DDSite != "US1" {
		t.Fatalf("ddSite = %q, want US1", state.DDSite)
	}

	state, next = ApplyDatadogAPIKeyEntered(state, "api-key")
	if next != GateDatadogAppKey {
		t.Fatalf("next gate = %q, want %q", next, GateDatadogAppKey)
	}
	if state.DDAPIKey != "api-key" {
		t.Fatalf("ddAPIKey = %q, want api-key", state.DDAPIKey)
	}

	state, next = ApplyDatadogAccountCreated(state, "dd-1")
	if next != GateDatadogDiscovery {
		t.Fatalf("next gate = %q, want %q", next, GateDatadogDiscovery)
	}
	if state.DDAccount != "dd-1" {
		t.Fatalf("ddAccount = %q, want dd-1", state.DDAccount)
	}
}

func TestApplyWorkspaceSelected(t *testing.T) {
	t.Parallel()

	workspace := domain.Workspace{ID: "ws-1", Name: "Workspace 1"}
	state, next := ApplyWorkspaceSelected(State{}, workspace)
	if next != GateSync {
		t.Fatalf("next gate = %q, want %q", next, GateSync)
	}
	if state.Workspace == nil || state.Workspace.ID != workspace.ID {
		t.Fatalf("workspace not applied: %#v", state.Workspace)
	}
}
