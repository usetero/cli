package domain

import (
	"time"

	"github.com/google/uuid"
)

// MessageID is a typed identifier for messages.
type MessageID string

// NewMessageID generates a new unique MessageID.
func NewMessageID() MessageID {
	return MessageID(uuid.New().String())
}

// String returns the string representation of the MessageID.
func (id MessageID) String() string {
	return string(id)
}

// Role identifies who sent a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a chat message.
type Message struct {
	ID             MessageID      `json:"id"`
	ConversationID ConversationID `json:"conversation_id"`
	Role           Role           `json:"role"`
	Content        []Block        `json:"content"`
	Model          string         `json:"model,omitempty"`
	StopReason     string         `json:"stop_reason,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}
