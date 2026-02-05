package msgs

import (
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
)

// UserSubmittedInput is fired when user input is ready (text from command bar or tool results).
type UserSubmittedInput struct {
	Text        string
	ToolResults []tools.Result
}

// TurnStarted is fired when a new turn begins.
type TurnStarted struct {
	UserMessageID  domain.MessageID
	ConversationID domain.ConversationID
}

// AssistantContentUpdated is fired as the assistant message streams in.
type AssistantContentUpdated struct {
	TurnID  domain.MessageID // user message ID that started this turn
	Message domain.Message
}

// StreamCompleted is fired when the assistant stream finishes.
type StreamCompleted struct {
	TurnID     domain.MessageID // user message ID that started this turn
	Message    domain.Message
	StopReason string
}

// AssistantMessageCreated is fired after the assistant message is persisted.
type AssistantMessageCreated struct {
	MessageID domain.MessageID
}

// Tool-specific completion messages.
// Each tool fires its own message type for full type safety.

// QueryCompleted is fired when a query tool finishes executing.
type QueryCompleted struct {
	ToolUseID string
	Result    tools.QueryResult
	Error     error
}

// StartJourneyCompleted is fired when a start_journey tool finishes executing.
type StartJourneyCompleted struct {
	ToolUseID string
	Result    tools.StartJourneyResult
	Error     error
}

// EndJourneyCompleted is fired when an end_journey tool finishes executing.
type EndJourneyCompleted struct {
	ToolUseID string
	Result    tools.EndJourneyResult
	Error     error
}
