package usecase

import (
	"context"

	corechat "github.com/usetero/cli/internal/core/chat"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
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

// MemoryToolLoop prepares the next turn entirely in memory, minting a local
// message ID and appending the tool results to the in-memory session.
type MemoryToolLoop struct{}

func NewMemoryToolLoop() *MemoryToolLoop {
	return &MemoryToolLoop{}
}

func (t *MemoryToolLoop) PrepareNextTurn(_ context.Context, input PrepareNextTurnInput) (PreparedNextTurn, error) {
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

	msgID := domain.NewMessageID()

	session := input.Session
	if session == nil {
		session = corechat.NewSession(input.ConversationID, nil)
	}
	toolResultMessage := session.AppendUserToolResultsMessage(msgID, domainResults)

	return PreparedNextTurn{
		MessageID:         msgID,
		Messages:          session.Messages(),
		ToolResultMessage: toolResultMessage,
	}, nil
}
