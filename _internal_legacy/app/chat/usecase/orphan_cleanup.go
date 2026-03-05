package usecase

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite"
)

// OrphanMessageCleaner removes uncommitted/orphaned messages after cancellation/failure.
type OrphanMessageCleaner interface {
	CleanupMessages(ctx context.Context, ids []domain.MessageID) error
}

type SQLiteOrphanMessageCleaner struct {
	db sqlite.DB
}

func NewSQLiteOrphanMessageCleaner(db sqlite.DB) *SQLiteOrphanMessageCleaner {
	return &SQLiteOrphanMessageCleaner{db: db}
}

func (c *SQLiteOrphanMessageCleaner) CleanupMessages(ctx context.Context, ids []domain.MessageID) error {
	for _, id := range ids {
		if err := c.db.Messages().Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
