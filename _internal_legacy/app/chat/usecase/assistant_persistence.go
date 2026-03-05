package usecase

import (
	"context"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite"
)

type PersistAssistantInput struct {
	AccountID      domain.AccountID
	ConversationID domain.ConversationID
	Message        domain.Message
}

type AssistantPersister interface {
	PersistAssistant(ctx context.Context, input PersistAssistantInput) (domain.MessageID, error)
}

type SQLiteAssistantPersister struct {
	db sqlite.DB
}

func NewSQLiteAssistantPersister(db sqlite.DB) *SQLiteAssistantPersister {
	return &SQLiteAssistantPersister{db: db}
}

func (p *SQLiteAssistantPersister) PersistAssistant(ctx context.Context, input PersistAssistantInput) (domain.MessageID, error) {
	msgID, err := p.db.Messages().CreateAssistantMessage(
		ctx,
		input.AccountID,
		input.ConversationID,
		input.Message.Model,
	)
	if err != nil {
		return "", err
	}

	content, err := domain.EncodeBlocks(input.Message.Content)
	if err != nil {
		return "", err
	}
	if err := p.db.Messages().UpdateContent(ctx, msgID, content); err != nil {
		return "", err
	}
	if err := p.db.Messages().UpdateMeta(ctx, msgID, input.Message.Model, input.Message.StopReason); err != nil {
		return "", err
	}
	return msgID, nil
}
