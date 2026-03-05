package preferences

import (
	"context"
	"path/filepath"
	"testing"

	domainprefs "github.com/usetero/cli/internal/domains/preferences"
	"github.com/usetero/cli/internal/domains/tenancy"
)

func TestStore_SaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore("dev")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	in := domainprefs.Snapshot{
		Role:         domainprefs.RoleEngineer,
		Organization: tenancy.OrganizationID("org_1"),
		Account:      tenancy.AccountID("acct_1"),
		Workspace:    tenancy.WorkspaceID("ws_1"),
	}
	if err := store.Save(context.Background(), in); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got != in {
		t.Fatalf("snapshot mismatch: got=%+v want=%+v", got, in)
	}
}

func TestStore_LoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore("dev")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != (domainprefs.Snapshot{}) {
		t.Fatalf("expected empty snapshot, got=%+v", got)
	}
}

func TestStore_Path(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := NewStore("prd")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	want := filepath.Join(home, ".tero", "environments", "prd", "preferences.json")
	if store.path != want {
		t.Fatalf("path mismatch: got=%q want=%q", store.path, want)
	}
}
