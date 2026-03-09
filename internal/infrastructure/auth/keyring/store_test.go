package keyring

import (
	"testing"
)

func TestStore_FileBackendRoundTrip(t *testing.T) {
	t.Setenv(EnvBackend, BackendFile)
	t.Setenv(EnvPath, t.TempDir()+"/secrets.json")

	store, err := NewStore("local")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if err := store.Set(KeyAccessToken, "access"); err != nil {
		t.Fatalf("Set(access) error = %v", err)
	}
	if err := store.Set(KeyRefreshToken, "refresh"); err != nil {
		t.Fatalf("Set(refresh) error = %v", err)
	}

	access, err := store.Get(KeyAccessToken)
	if err != nil {
		t.Fatalf("Get(access) error = %v", err)
	}
	if access != "access" {
		t.Fatalf("access = %q, want access", access)
	}

	refresh, err := store.Get(KeyRefreshToken)
	if err != nil {
		t.Fatalf("Get(refresh) error = %v", err)
	}
	if refresh != "refresh" {
		t.Fatalf("refresh = %q, want refresh", refresh)
	}

	if err := store.Delete(KeyAccessToken); err != nil {
		t.Fatalf("Delete(access) error = %v", err)
	}
	access, err = store.Get(KeyAccessToken)
	if err != nil {
		t.Fatalf("Get(access after delete) error = %v", err)
	}
	if access != "" {
		t.Fatalf("access after delete = %q, want empty", access)
	}
}

func TestStore_InvalidBackend(t *testing.T) {
	t.Setenv(EnvBackend, "bogus")

	if _, err := NewStore("local"); err == nil {
		t.Fatal("expected invalid backend error")
	}
}
