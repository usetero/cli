package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestApplyPreflight(t *testing.T) {
	t.Parallel()

	org := &domain.Organization{ID: "org-1", Name: "Org 1"}
	account := &domain.Account{ID: "acc-1", Name: "Acc 1"}

	state, next := ApplyPreflight(State{}, PreflightState{
		Outcome:      PreflightOutcomeResolved,
		HasValidAuth: true,
		Role:         RolePlatform,
		Org:          org,
		Account:      account,
	})

	if next != GateRuntimeInit {
		t.Fatalf("next gate = %q, want %q", next, GateRuntimeInit)
	}
	if state.Org == nil || state.Org.ID != org.ID {
		t.Fatalf("org not applied: %#v", state.Org)
	}
	if state.Account == nil || state.Account.ID != account.ID {
		t.Fatalf("account not applied: %#v", state.Account)
	}
}
