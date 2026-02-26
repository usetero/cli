package chat

// Protocol types for the Chat API SSE stream.
// These types are internal to the chat package - the TUI and other callers
// only see the final domain.Message that gets built from these events.

// EventType identifies the kind of SSE event from the Chat API.
type EventType string

const (
	// Delta events - streaming content fragments.
	EventTypeTextDelta      EventType = "text_delta"
	EventTypeThinkingDelta  EventType = "thinking_delta"
	EventTypeToolInputDelta EventType = "tool_input_delta"

	// Complete block events.
	EventTypeText       EventType = "text"
	EventTypeThinking   EventType = "thinking"
	EventTypeToolUse    EventType = "tool_use"
	EventTypeToolResult EventType = "tool_result"

	// Stream lifecycle events.
	EventTypeMessageStart     EventType = "message_start"
	EventTypeMessageStop      EventType = "message_stop"
	EventTypeContentBlockStop EventType = "content_block_stop"

	// Post-stream metadata.
	EventTypeMetadataUpdate EventType = "metadata_update"
)

// event is a single event from the Chat API response stream.
// This is internal to the chat package.
type event struct {
	ConversationID string        `json:"conversation_id,omitempty"`
	TurnID         string        `json:"turn_id,omitempty"`
	Seq            int           `json:"seq,omitempty"`
	Type           EventType     `json:"type"`
	Text           *textContent  `json:"text,omitempty"`
	Thinking       *textContent  `json:"thinking,omitempty"`
	ToolUse        *toolUseEvent `json:"tool_use,omitempty"`
	ToolInputDelta string        `json:"tool_input_delta,omitempty"`
	MessageStart   *messageStart `json:"message_start,omitempty"`
	MessageStop    *messageStop  `json:"message_stop,omitempty"`
	Metadata       *metadata     `json:"metadata,omitempty"`
	Done           bool          `json:"-"` // Set when stream is complete (after [DONE])
}

// textContent holds text or thinking content.
type textContent struct {
	Content string `json:"content"`
}

// toolUseEvent is a tool call event from the stream.
type toolUseEvent struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Input []byte `json:"input,omitempty"` // Only populated for complete tool_use, not during streaming
}

// messageStart contains metadata sent at the start of a message stream.
type messageStart struct {
	Model         string `json:"model"`
	ContextWindow int    `json:"context_window,omitempty"`
}

// messageStop contains metadata sent at the end of a message stream.
type messageStop struct {
	StopReason   string `json:"stop_reason"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
}

// metadata contains post-stream metadata updates.
type metadata struct {
	Title string `json:"title,omitempty"`
}
