package usecase

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
)

func TestSQLiteOrphanMessageCleaner_CleanupMessages_Success(t *testing.T) {
	t.Parallel()

	db := dbtest.OpenTestDB(t)
	cleaner := NewSQLiteOrphanMessageCleaner(db)

	ids := make([]domain.MessageID, 0, 2)
	for _, text := range []string{"hello", "world"} {
		id, err := db.Messages().CreateUserMessage(context.Background(), "acct-1", "conv-1", text)
		if err != nil {
			t.Fatalf("CreateUserMessage() error = %v", err)
		}
		ids = append(ids, id)
	}

	if err := cleaner.CleanupMessages(context.Background(), ids); err != nil {
		t.Fatalf("CleanupMessages() error = %v", err)
	}

	for _, id := range ids {
		if _, err := db.Messages().Get(context.Background(), id); err == nil {
			t.Fatalf("message %s still exists after cleanup", id)
		}
	}
}

func TestSQLiteOrphanMessageCleaner_CleanupMessages_ErrorOnMissingSchema(t *testing.T) {
	t.Parallel()

	db := sqlitetest.OpenBareDB(t)
	cleaner := NewSQLiteOrphanMessageCleaner(db)

	err := cleaner.CleanupMessages(context.Background(), []domain.MessageID{"msg-1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
