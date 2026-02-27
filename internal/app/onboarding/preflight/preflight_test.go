package preflight

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestResolveOrg(t *testing.T) {
	t.Parallel()

	orgs := []domain.Organization{
		{ID: "org-1", Name: "One"},
		{ID: "org-2", Name: "Two"},
	}

	if got := resolveOrg(orgs, "org-2"); got == nil || got.ID != "org-2" {
		t.Fatalf("resolveOrg by id failed: %+v", got)
	}
	if got := resolveOrg(orgs, "missing"); got != nil {
		t.Fatalf("resolveOrg with missing id should be nil, got %+v", got)
	}
	if got := resolveOrg([]domain.Organization{{ID: "only", Name: "Only"}}, ""); got == nil || got.ID != "only" {
		t.Fatalf("resolveOrg single fallback failed: %+v", got)
	}
}

func TestResolveAccount(t *testing.T) {
	t.Parallel()

	accounts := []domain.Account{
		{ID: "acc-1", Name: "One"},
		{ID: "acc-2", Name: "Two"},
	}

	if got := resolveAccount(accounts, "acc-1"); got == nil || got.ID != "acc-1" {
		t.Fatalf("resolveAccount by id failed: %+v", got)
	}
	if got := resolveAccount(accounts, "missing"); got != nil {
		t.Fatalf("resolveAccount with missing id should be nil, got %+v", got)
	}
	if got := resolveAccount([]domain.Account{{ID: "only", Name: "Only"}}, ""); got == nil || got.ID != "only" {
		t.Fatalf("resolveAccount single fallback failed: %+v", got)
	}
}
