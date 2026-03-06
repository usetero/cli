package sqlite

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountIDValidate(t *testing.T) {
	t.Parallel()

	if err := AccountID("").Validate(); err == nil {
		t.Fatalf("expected empty account id error")
	}
	if err := AccountID("acc_1").Validate(); err != nil {
		t.Fatalf("unexpected account id validation error: %v", err)
	}
}

func TestDatabasePathValidate(t *testing.T) {
	t.Parallel()

	if err := DatabasePath("").Validate(); err == nil {
		t.Fatalf("expected empty path error")
	}
	if err := DatabasePath("/tmp/tero.sqlite").Validate(); err != nil {
		t.Fatalf("unexpected database path validation error: %v", err)
	}
}

func TestStorageDatabasePathBuildsScopedPath(t *testing.T) {
	t.Parallel()

	s := Storage{
		BaseDir: "/base",
		Env:     "dev",
		OrgID:   "org_1",
	}
	got, err := s.DatabasePath(AccountID("acc_1"))
	if err != nil {
		t.Fatalf("database path: %v", err)
	}

	want := filepath.Join("/base", "dev", "orgs", "org_1", "accounts", "acc_1", "tero.sqlite")
	if got.String() != want {
		t.Fatalf("database path mismatch: got %q want %q", got, want)
	}
}

func TestStorageDatabasePathValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		storage Storage
		id      AccountID
		wantErr string
	}{
		{name: "missing account", storage: Storage{BaseDir: "/base", Env: "dev", OrgID: "org_1"}, id: "", wantErr: "account id is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.storage.DatabasePath(tt.id)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
