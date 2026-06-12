package usecase

import (
	"context"

	"github.com/usetero/cli/internal/domain"
)

// OrphanMessageCleaner removes uncommitted/orphaned messages after cancellation/failure.
type OrphanMessageCleaner interface {
	CleanupMessages(ctx context.Context, ids []domain.MessageID) error
}

// MemoryOrphanMessageCleaner is a no-op cleaner. With ephemeral chat there is
// no persisted store to reconcile; the message list drops cancelled rounds in
// the UI, so orphan cleanup has nothing to do.
type MemoryOrphanMessageCleaner struct{}

func NewMemoryOrphanMessageCleaner() *MemoryOrphanMessageCleaner {
	return &MemoryOrphanMessageCleaner{}
}

func (c *MemoryOrphanMessageCleaner) CleanupMessages(_ context.Context, _ []domain.MessageID) error {
	return nil
}
