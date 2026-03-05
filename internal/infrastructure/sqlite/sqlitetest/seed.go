package sqlitetest

import (
	"context"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func SeedAccount(t *testing.T, db *sqlite.DB, id, name string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), "INSERT INTO accounts (id, name, created_at) VALUES (?, ?, ?)", id, name, time.Now().UTC().Format(time.RFC3339)); err != nil {
		t.Fatalf("seed account: %v", err)
	}
}

func SeedWorkspace(t *testing.T, db *sqlite.DB, id, accountID, name string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), "INSERT INTO workspaces (id, account_id, name, created_at) VALUES (?, ?, ?, ?)", id, accountID, name, time.Now().UTC().Format(time.RFC3339)); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
}
