package usecase

import "github.com/usetero/cli/internal/domain"

// StreamRequest is the app-level request for one assistant turn stream.
type StreamRequest struct {
	ConversationID  domain.ConversationID
	Messages        []domain.Message
	ContextEntities []domain.ContextEntity
}
