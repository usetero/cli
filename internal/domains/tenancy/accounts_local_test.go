package tenancy

import (
	"context"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/infrastructure/sqlite/sqlitetest"
)

func openTenancyTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	return sqlitetest.Open(t)
}

func TestLocalAccountService_CRUD(t *testing.T) {
	db := openTenancyTestDB(t)
	svc := NewLocalAccountService(db.Raw())

	createdID, err := svc.Create(context.Background(), "Primary")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if createdID == "" {
		t.Fatalf("expected non-empty account id")
	}

	accounts, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected one account, got %d", len(accounts))
	}
	if accounts[0].ID != createdID || accounts[0].Name != "Primary" {
		t.Fatalf("unexpected account row: %+v", accounts[0])
	}

	if err := svc.Delete(context.Background(), createdID); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	accounts, err = svc.List(context.Background())
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("expected no accounts after delete, got %d", len(accounts))
	}
}

func TestLocalAccountService_ValidationAndUninitialized(t *testing.T) {
	db := openTenancyTestDB(t)
	svc := NewLocalAccountService(db.Raw())

	if _, err := svc.Create(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account name is required") {
		t.Fatalf("expected account name validation error, got %v", err)
	}
	if err := svc.Delete(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "account id is required") {
		t.Fatalf("expected account id validation error, got %v", err)
	}

	var nilSvc *LocalAccountService
	if _, err := nilSvc.List(context.Background()); err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected uninitialized error, got %v", err)
	}
}
