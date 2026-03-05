// Package domaintest provides factories for creating domain objects in tests.
// Each factory returns a valid object with unique IDs and sensible defaults.
// Mutate the returned struct to customize for your test.
package domaintest

import (
	"encoding/json"
	"time"

	"github.com/usetero/cli/internal/domain"
)

// NewMessage returns a valid assistant message with a text block.
func NewMessage() domain.Message {
	return domain.Message{
		ID:             domain.NewMessageID(),
		ConversationID: domain.ConversationID(domain.NewMessageID()),
		Role:           domain.RoleAssistant,
		Content:        []domain.Block{NewTextBlock()},
		Model:          "test-model",
		StopReason:     "end_turn",
		CreatedAt:      time.Now(),
	}
}

// NewUserMessage returns a valid user message with a text block.
func NewUserMessage() domain.Message {
	msg := NewMessage()
	msg.Role = domain.RoleUser
	msg.Model = ""
	msg.StopReason = ""
	return msg
}

// NewTextBlock returns a text block at index 0.
func NewTextBlock() domain.Block {
	return domain.Block{
		Index: 0,
		Type:  domain.BlockTypeText,
		Text:  &domain.TextBlock{Content: "Hello, world."},
	}
}

// NewThinkingBlock returns a thinking block at index 0.
func NewThinkingBlock() domain.Block {
	return domain.Block{
		Index:    0,
		Type:     domain.BlockTypeThinking,
		Thinking: &domain.Thinking{Content: "Let me think about this."},
	}
}

// NewToolUseBlock returns a tool_use block at index 0.
func NewToolUseBlock() domain.Block {
	return domain.Block{
		Index: 0,
		Type:  domain.BlockTypeToolUse,
		ToolUse: &domain.ToolUse{
			ID:    domain.NewMessageID().String(),
			Name:  "query",
			Input: json.RawMessage(`{}`),
		},
	}
}
