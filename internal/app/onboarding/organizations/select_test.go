package organizations

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestFindOrgByID(t *testing.T) {
	t.Parallel()

	orgs := []domain.Organization{
		{ID: "org-1", Name: "One"},
		{ID: "org-2", Name: "Two"},
	}

	got := findOrgByID(orgs, "org-2")
	if got == nil || got.ID != "org-2" {
		t.Fatalf("expected org-2, got %#v", got)
	}

	if got := findOrgByID(orgs, "missing"); got != nil {
		t.Fatalf("expected nil for missing org, got %#v", got)
	}

	if got := findOrgByID(orgs, ""); got != nil {
		t.Fatalf("expected nil for empty id, got %#v", got)
	}
}
