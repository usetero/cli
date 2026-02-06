package msgs

import "github.com/usetero/cli/internal/domain"

// ConversationCreated is fired when a new conversation is created.
type ConversationCreated struct {
	ConversationID domain.ConversationID
}

// RoundStarted is fired when a new round begins.
type RoundStarted struct {
	RoundID        domain.MessageID
	ConversationID domain.ConversationID
}

// AssistantMessageCreated is fired after the assistant message is persisted.
type AssistantMessageCreated struct {
	MessageID domain.MessageID
}
