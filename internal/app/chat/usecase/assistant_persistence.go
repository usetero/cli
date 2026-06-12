package usecase

import (
	"context"

	"github.com/usetero/cli/internal/domain"
)

type PersistAssistantInput struct {
	AccountID      domain.AccountID
	ConversationID domain.ConversationID
	Message        domain.Message
}

type AssistantPersister interface {
	PersistAssistant(ctx context.Context, input PersistAssistantInput) (domain.MessageID, error)
}

// MemoryAssistantPersister mints assistant message IDs without persisting.
// Chat is ephemeral: the rendered content lives in the in-memory message list
// for the duration of the session and is intentionally not stored.
type MemoryAssistantPersister struct{}

func NewMemoryAssistantPersister() *MemoryAssistantPersister {
	return &MemoryAssistantPersister{}
}

func (p *MemoryAssistantPersister) PersistAssistant(_ context.Context, _ PersistAssistantInput) (domain.MessageID, error) {
	return domain.NewMessageID(), nil
}
