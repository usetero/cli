package usecase

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestMemoryOrphanMessageCleaner_CleanupMessages_NoOp(t *testing.T) {
	t.Parallel()

	cleaner := NewMemoryOrphanMessageCleaner()

	if err := cleaner.CleanupMessages(context.Background(), []domain.MessageID{"msg-1", "msg-2"}); err != nil {
		t.Fatalf("CleanupMessages() error = %v", err)
	}
}

func TestMemoryOrphanMessageCleaner_CleanupMessages_EmptyIDs(t *testing.T) {
	t.Parallel()

	cleaner := NewMemoryOrphanMessageCleaner()

	if err := cleaner.CleanupMessages(context.Background(), nil); err != nil {
		t.Fatalf("CleanupMessages() error = %v", err)
	}
}
