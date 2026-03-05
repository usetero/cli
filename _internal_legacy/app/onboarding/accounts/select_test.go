package accounts

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestFindAccountByID(t *testing.T) {
	t.Parallel()

	accounts := []domain.Account{
		{ID: "acc-1", Name: "One"},
		{ID: "acc-2", Name: "Two"},
	}

	got := findAccountByID(accounts, "acc-2")
	if got == nil || got.ID != "acc-2" {
		t.Fatalf("expected acc-2, got %#v", got)
	}

	if got := findAccountByID(accounts, "missing"); got != nil {
		t.Fatalf("expected nil for missing account, got %#v", got)
	}

	if got := findAccountByID(accounts, ""); got != nil {
		t.Fatalf("expected nil for empty id, got %#v", got)
	}
}
