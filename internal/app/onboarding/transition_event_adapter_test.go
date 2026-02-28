package onboarding

import (
	"testing"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func TestBootstrapEventFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  any
		kind bootstrap.EventKind
	}{
		{name: "authenticated", msg: msgs.Authenticated{}, kind: bootstrap.EventAuthenticated},
		{name: "org selected", msg: msgs.OrgSelected{}, kind: bootstrap.EventOrgSelected},
		{name: "runtime ready", msg: msgs.RuntimeReady{}, kind: bootstrap.EventRuntimeReady},
		{name: "sync complete", msg: msgs.SyncComplete{}, kind: bootstrap.EventSyncComplete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, ok := bootstrapEventFor(tt.msg)
			if !ok {
				t.Fatalf("expected event for %T", tt.msg)
			}
			if event.Kind != tt.kind {
				t.Fatalf("kind = %q, want %q", event.Kind, tt.kind)
			}
		})
	}
}

func TestBootstrapEventForPreflight(t *testing.T) {
	t.Parallel()

	org := &domain.Organization{ID: "org-1"}
	account := &domain.Account{ID: "acc-1"}
	msg := msgs.PreflightResolved{
		State: msgs.PreflightState{
			Outcome:      msgs.PreflightOutcomeResolved,
			HasValidAuth: true,
			Role:         msgs.RolePlatform,
			Org:          org,
			Account:      account,
		},
	}
	event, ok := bootstrapEventFor(msg)
	if !ok {
		t.Fatal("expected preflight event")
	}
	if event.Kind != bootstrap.EventPreflightResolved {
		t.Fatalf("kind = %q, want %q", event.Kind, bootstrap.EventPreflightResolved)
	}
	if event.Preflight.Org == nil || event.Preflight.Org.ID != org.ID {
		t.Fatalf("preflight org = %#v, want %q", event.Preflight.Org, org.ID)
	}
	if event.Preflight.Account == nil || event.Preflight.Account.ID != account.ID {
		t.Fatalf("preflight account = %#v, want %q", event.Preflight.Account, account.ID)
	}
}

func TestBootstrapEventForUnknownMessage(t *testing.T) {
	t.Parallel()

	_, ok := bootstrapEventFor(struct{}{})
	if ok {
		t.Fatal("expected unknown message to be ignored")
	}
}
