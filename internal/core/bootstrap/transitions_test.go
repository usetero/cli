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

func TestApplyOrgSelectedClearsAccountScopedState(t *testing.T) {
	t.Parallel()

	org := domain.Organization{ID: "org-2", Name: "Org 2"}
	state, next := ApplyOrgSelected(State{
		Org:       &domain.Organization{ID: "org-1"},
		Account:   &domain.Account{ID: "acc-1"},
		DDSite:    "US1",
		DDAPIKey:  "api-key",
		DDAccount: "dd-1",
	}, org)
	if next != GateAccountSelect {
		t.Fatalf("next gate = %q, want %q", next, GateAccountSelect)
	}
	if state.Org == nil || state.Org.ID != org.ID {
		t.Fatalf("org not applied: %#v", state.Org)
	}
	if state.Account != nil || state.DDSite != "" || state.DDAPIKey != "" || state.DDAccount != "" {
		t.Fatalf("account-scoped state not cleared: %#v", state)
	}
}

func TestApplyAccountSelectedClearsScopedStateBeforeSettingAccount(t *testing.T) {
	t.Parallel()

	org := domain.Organization{ID: "org-1", Name: "Org 1"}
	account := domain.Account{ID: "acc-2", Name: "Account 2"}
	state, next := ApplyAccountSelected(State{
		Account:   &domain.Account{ID: "acc-1"},
		DDSite:    "US1",
		DDAPIKey:  "api-key",
		DDAccount: "dd-1",
	}, org, account)
	if next != GateRuntimeInit {
		t.Fatalf("next gate = %q, want %q", next, GateRuntimeInit)
	}
	if state.Account == nil || state.Account.ID != account.ID {
		t.Fatalf("account not applied: %#v", state.Account)
	}
	if state.DDSite != "" || state.DDAPIKey != "" || state.DDAccount != "" {
		t.Fatalf("account-scoped state not reset: %#v", state)
	}
}
