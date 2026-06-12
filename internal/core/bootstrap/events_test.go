package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

func TestApplyEventAuthenticated(t *testing.T) {
	t.Parallel()

	got := ApplyEvent(State{}, Event{
		Kind: EventAuthenticated,
		User: auth.User{ID: "user-1"},
	})
	if got.Kind != TransitionAdvance {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionAdvance)
	}
	if got.Next != GateRoleSelect {
		t.Fatalf("next = %q, want %q", got.Next, GateRoleSelect)
	}
	if got.State.User == nil || got.State.User.ID != "user-1" {
		t.Fatalf("user not applied: %#v", got.State.User)
	}
}

func TestApplyEventDatadogDiscoveryDoneCompletes(t *testing.T) {
	t.Parallel()

	// Datadog discovery is the terminal onboarding step; the account is the
	// working context (no workspace).
	state := State{
		User:    &auth.User{ID: "user-1"},
		Org:     &domain.Organization{ID: "org-1"},
		Account: &domain.Account{ID: "acc-1"},
	}
	got := ApplyEvent(state, Event{Kind: EventDatadogDiscoveryDone})
	if got.Kind != TransitionComplete {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionComplete)
	}
	if got.Completion.User == nil || got.Completion.User.ID != "user-1" {
		t.Fatalf("completion user = %#v", got.Completion.User)
	}
	if got.Completion.Org.ID != "org-1" || got.Completion.Account.ID != "acc-1" {
		t.Fatalf("unexpected completion payload: %#v", got.Completion)
	}
}

func TestApplyEventDatadogReadyCompletes(t *testing.T) {
	t.Parallel()

	state := State{
		User:    &auth.User{ID: "user-1"},
		Org:     &domain.Organization{ID: "org-1"},
		Account: &domain.Account{ID: "acc-1"},
	}
	got := ApplyEvent(state, Event{Kind: EventDatadogReady})
	if got.Kind != TransitionComplete {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionComplete)
	}
}

func TestApplyEventDatadogDiscoveryDoneMissingStateNoops(t *testing.T) {
	t.Parallel()

	got := ApplyEvent(State{User: &auth.User{ID: "user-1"}}, Event{Kind: EventDatadogDiscoveryDone})
	if got.Kind != TransitionNoop {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionNoop)
	}
}
