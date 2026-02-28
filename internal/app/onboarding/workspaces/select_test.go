package workspaces

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestFindWorkspaceByID(t *testing.T) {
	t.Parallel()

	workspaces := []domain.Workspace{
		{ID: "ws-1", Name: "One"},
		{ID: "ws-2", Name: "Two"},
	}

	got := findWorkspaceByID(workspaces, "ws-2")
	if got == nil || got.ID != "ws-2" {
		t.Fatalf("expected ws-2, got %#v", got)
	}

	if got := findWorkspaceByID(workspaces, "missing"); got != nil {
		t.Fatalf("expected nil for missing workspace, got %#v", got)
	}

	if got := findWorkspaceByID(workspaces, ""); got != nil {
		t.Fatalf("expected nil for empty id, got %#v", got)
	}
}
