package turn

import (
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
)

// Blocks returns all visual blocks for the viewport.
// Includes user message (if visible) followed by assistant blocks.
func (m *Model) Blocks() []block.Block {
	var result []block.Block
	if m.userMessage.IsVisible() {
		result = append(result, m.userMessage)
	}
	result = append(result, m.assistantMessage.Blocks()...)
	return result
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
	// User block gets content width minus border+padding, same as assistant blocks.
	// The border decoration is applied by renderBlock in the message list.
	m.userMessage.SetWidth(width - block.BorderWidth)
	m.assistantMessage.SetWidth(width)
}

// State returns the turn's current state.
func (m *Model) State() State {
	return m.state
}

// UserMessageID returns the user message ID.
func (m *Model) UserMessageID() domain.MessageID {
	return m.userMessage.ID()
}

// UserInput returns the input that created this turn's user message.
func (m *Model) UserInput() msgs.UserSubmittedInput {
	return m.userMessage.Input()
}

// AssistantMessageID returns the persisted assistant message ID.
// Returns empty string if the assistant message was never persisted.
func (m *Model) AssistantMessageID() domain.MessageID {
	return m.assistantMessage.ID()
}

// Cancel stops the in-flight stream and marks the turn complete.
// The partial content remains rendered but nothing is persisted.
func (m *Model) Cancel() {
	if m.stream != nil && !m.stream.done {
		m.stream.cancel(errUserCancelled)
		m.stream.done = true
	}
	m.assistantMessage.Cancel()
	m.state = StateComplete
}
