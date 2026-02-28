package preflight

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/domain"
)

func TestResolveOrg(t *testing.T) {
	t.Parallel()

	orgs := []domain.Organization{
		{ID: "org-1", Name: "One"},
		{ID: "org-2", Name: "Two"},
	}

	got := resolveOrg(orgs, "org-2")
	if got == nil || got.ID != "org-2" {
		t.Fatalf("expected org-2, got %#v", got)
	}

	got = resolveOrg(orgs, "")
	if got != nil {
		t.Fatalf("expected nil when multiple orgs and no preference")
	}

	got = resolveOrg(orgs[:1], "")
	if got == nil || got.ID != "org-1" {
		t.Fatalf("expected single org fallback, got %#v", got)
	}
}

func TestResolveAccount(t *testing.T) {
	t.Parallel()

	accounts := []domain.Account{
		{ID: "acc-1", Name: "One"},
		{ID: "acc-2", Name: "Two"},
	}

	got := resolveAccount(accounts, "acc-2")
	if got == nil || got.ID != "acc-2" {
		t.Fatalf("expected acc-2, got %#v", got)
	}

	got = resolveAccount(accounts, "")
	if got != nil {
		t.Fatalf("expected nil when multiple accounts and no preference")
	}

	got = resolveAccount(accounts[:1], "")
	if got == nil || got.ID != "acc-1" {
		t.Fatalf("expected single account fallback, got %#v", got)
	}
}

func TestPreflightOutcomeForError(t *testing.T) {
	t.Parallel()

	outcome, _ := preflightOutcomeForError(context.DeadlineExceeded)
	if outcome != bootstrap.PreflightOutcomeFailed {
		t.Fatalf("expected failed outcome for deadline exceeded, got %v", outcome)
	}

	outcome, _ = preflightOutcomeForError(errors.New("boom"))
	if outcome != bootstrap.PreflightOutcomeInconclusive {
		t.Fatalf("expected inconclusive outcome for generic error, got %v", outcome)
	}
}
