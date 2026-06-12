package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestEventFromMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  Message
		kind EventKind
	}{
		{name: "authenticated", msg: Authenticated{}, kind: EventAuthenticated},
		{name: "org selected", msg: OrgSelected{}, kind: EventOrgSelected},
		{name: "runtime ready", msg: RuntimeReady{}, kind: EventRuntimeReady},
		{name: "workspace selected", msg: WorkspaceSelected{}, kind: EventWorkspaceSelected},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, ok := EventFromMessage(tt.msg)
			if !ok {
				t.Fatalf("expected event for %T", tt.msg)
			}
			if event.Kind != tt.kind {
				t.Fatalf("kind = %q, want %q", event.Kind, tt.kind)
			}
		})
	}
}

func TestEventFromMessagePreflight(t *testing.T) {
	t.Parallel()

	org := &domain.Organization{ID: "org-1"}
	account := &domain.Account{ID: "acc-1"}
	msg := PreflightResolved{
		State: PreflightState{
			Outcome:      PreflightOutcomeResolved,
			HasValidAuth: true,
			Role:         RolePlatform,
			Org:          org,
			Account:      account,
		},
	}
	event, ok := EventFromMessage(msg)
	if !ok {
		t.Fatal("expected preflight event")
	}
	if event.Kind != EventPreflightResolved {
		t.Fatalf("kind = %q, want %q", event.Kind, EventPreflightResolved)
	}
	if event.Preflight.Org == nil || event.Preflight.Org.ID != org.ID {
		t.Fatalf("preflight org = %#v, want %q", event.Preflight.Org, org.ID)
	}
	if event.Preflight.Account == nil || event.Preflight.Account.ID != account.ID {
		t.Fatalf("preflight account = %#v, want %q", event.Preflight.Account, account.ID)
	}
}

func TestEventFromMessageNonTransitionMessage(t *testing.T) {
	t.Parallel()

	var msg Message
	_, ok := EventFromMessage(msg)
	if ok {
		t.Fatal("expected non-transition message to be ignored")
	}
}
