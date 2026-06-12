package domain

import (
	"time"

	"github.com/google/uuid"
)

// ConversationID is a typed identifier for conversations.
type ConversationID string

// NewConversationID generates a new unique ConversationID.
func NewConversationID() ConversationID {
	return ConversationID(uuid.New().String())
}

// String returns the string representation of the ConversationID.
func (id ConversationID) String() string {
	return string(id)
}

// Conversation represents a chat conversation.
type Conversation struct {
	ID        ConversationID `json:"id"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// ContextSource indicates who added an entity to context.
type ContextSource string

const (
	ContextSourceUser      ContextSource = "user"
	ContextSourceAssistant ContextSource = "assistant"
)

// ContextEntityType identifies supported entity kinds for chat context.
type ContextEntityType string

const (
	ContextEntityTypeService  ContextEntityType = "service"
	ContextEntityTypeLogEvent ContextEntityType = "log_event"
)

// ContextEntity is an entity attached to a conversation for the AI to reference.
// The client sends entity IDs; the server loads full entity data for the system prompt.
type ContextEntity struct {
	EntityType ContextEntityType `json:"entity_type"`
	EntityID   string            `json:"entity_id"`
}
