package chat

import "github.com/usetero/cli/internal/domain"

// Event is a single event from the Chat API response stream.
type Event struct {
	// Block contains the event data. The Type field indicates what kind:
	//   - MessageStart: Block.MessageStart has model info
	//   - TextDelta: Block.Text has the delta content
	//   - ThinkingDelta: Block.Thinking has the delta content
	//   - ToolUse: Block.ToolUse has the tool call
	//   - ToolInputDelta: Block.ToolInputDelta has JSON fragment
	//   - MessageStop: Block.MessageStop has stop_reason
	domain.Block

	// Done is true when the stream is complete.
	// This is sent after MessageStop.
	Done bool
}
