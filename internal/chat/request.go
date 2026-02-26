package chat

import "github.com/usetero/cli/internal/domain"

// Request is the input to the Chat API.
// The client sends the full conversation history on every request.
type Request struct {
	ConversationID  string                 `json:"conversation_id"`
	Messages        []domain.Message       `json:"messages"`
	ContextEntities []domain.ContextEntity `json:"context_entities,omitempty"`
	Tools           []Tool                 `json:"tools"`
}
