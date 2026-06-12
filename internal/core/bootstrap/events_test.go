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

func TestApplyEventWorkspaceSelectedCompletes(t *testing.T) {
	t.Parallel()

	// Workspace selection is the terminal onboarding step (no sync gate).
	state := State{
		User:    &auth.User{ID: "user-1"},
		Org:     &domain.Organization{ID: "org-1"},
		Account: &domain.Account{ID: "acc-1"},
	}
	got := ApplyEvent(state, Event{Kind: EventWorkspaceSelected, Workspace: domain.Workspace{ID: "ws-1"}})
	if got.Kind != TransitionComplete {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionComplete)
	}
	if got.Completion.User == nil || got.Completion.User.ID != "user-1" {
		t.Fatalf("completion user = %#v", got.Completion.User)
	}
	if got.Completion.Org.ID != "org-1" || got.Completion.Account.ID != "acc-1" || got.Completion.Workspace.ID != "ws-1" {
		t.Fatalf("unexpected completion payload: %#v", got.Completion)
	}
}

func TestApplyEventWorkspaceSelectedMissingStateNoops(t *testing.T) {
	t.Parallel()

	got := ApplyEvent(State{User: &auth.User{ID: "user-1"}}, Event{Kind: EventWorkspaceSelected, Workspace: domain.Workspace{ID: "ws-1"}})
	if got.Kind != TransitionNoop {
		t.Fatalf("kind = %q, want %q", got.Kind, TransitionNoop)
	}
}
