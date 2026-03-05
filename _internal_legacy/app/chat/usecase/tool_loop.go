package usecase

import (
	"context"

	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/sqlite"
)

type PrepareNextTurnInput struct {
	AccountID      domain.AccountID
	ConversationID domain.ConversationID
	Results        []domaintools.Result
	Session        *corechat.Session
}

type PreparedNextTurn struct {
	MessageID         domain.MessageID
	Messages          []domain.Message
	ToolResultMessage domain.Message
}

type ToolLoop interface {
	PrepareNextTurn(ctx context.Context, input PrepareNextTurnInput) (PreparedNextTurn, error)
}

type SQLiteToolLoop struct {
	db sqlite.DB
}

func NewSQLiteToolLoop(db sqlite.DB) *SQLiteToolLoop {
	return &SQLiteToolLoop{db: db}
}

func (t *SQLiteToolLoop) PrepareNextTurn(ctx context.Context, input PrepareNextTurnInput) (PreparedNextTurn, error) {
	domainResults := make([]domain.ToolResult, len(input.Results))
	for i, r := range input.Results {
		domainResults[i] = domain.ToolResult{
			ToolUseID: r.ToolUseID,
			IsError:   r.IsError(),
			Content:   r.ToMap(),
		}
		if r.Error != nil {
			domainResults[i].Error = r.Error.Message
		}
	}

	msgID, err := t.db.Messages().CreateToolResultMessage(ctx, input.AccountID, input.ConversationID, domainResults)
	if err != nil {
		msgID = domain.NewMessageID()
	}

	session := input.Session
	if session == nil {
		session = corechat.NewSession(input.ConversationID, nil)
	}
	toolResultMessage := session.AppendUserToolResultsMessage(msgID, domainResults)

	return PreparedNextTurn{
		MessageID:         msgID,
		Messages:          session.Messages(),
		ToolResultMessage: toolResultMessage,
	}, err
}
