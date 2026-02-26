package chat

import "encoding/json"

const chatStreamVersionV2 = "v2"

// EventType identifies the kind of SSE event from the Chat API.
type EventType string

const (
	EventTypeMessageStart     EventType = "message_start"
	EventTypeTextDelta        EventType = "text_delta"
	EventTypeThinkingDelta    EventType = "thinking_delta"
	EventTypeToolUse          EventType = "tool_use"
	EventTypeToolInputDelta   EventType = "tool_input_delta"
	EventTypeContentBlockStop EventType = "content_block_stop"
	EventTypeMessageStop      EventType = "message_stop"
	EventTypeMetadataUpdate   EventType = "metadata_update"
)

// event is a single event from the Chat API response stream.
// This is internal to the chat package.
type event struct {
	ChatStreamVersion string    `json:"chat_stream_version"`
	ConversationID    string    `json:"conversation_id,omitempty"`
	TurnID            string    `json:"turn_id,omitempty"`
	Seq               int       `json:"seq,omitempty"`
	Type              EventType `json:"type"`

	Text     *textContent `json:"text,omitempty"`
	Thinking *textContent `json:"thinking,omitempty"`

	ToolUse        *toolUseEvent `json:"tool_use,omitempty"`
	ToolUseID      string        `json:"tool_use_id,omitempty"`
	ToolInputDelta string        `json:"tool_input_delta,omitempty"`

	MessageStart *messageStart `json:"message_start,omitempty"`
	MessageStop  *messageStop  `json:"message_stop,omitempty"`
	Metadata     *metadata     `json:"metadata,omitempty"`

	Done bool `json:"-"`
}

type textContent struct {
	Content *string `json:"content"`
}

type toolUseEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type messageStart struct {
	Model         string `json:"model"`
	ContextWindow *int   `json:"context_window"`
}

type messageStop struct {
	StopReason   string `json:"stop_reason"`
	InputTokens  *int   `json:"input_tokens"`
	OutputTokens *int   `json:"output_tokens"`
}

type metadata struct {
	Title string `json:"title,omitempty"`
}

// errorEnvelope represents an error frame from the stream.
type errorEnvelope struct {
	Error any `json:"error"`
}

func parseErrorMessage(raw json.RawMessage) (string, bool) {
	var frame errorEnvelope
	if err := json.Unmarshal(raw, &frame); err != nil {
		return "", false
	}
	if frame.Error == nil {
		return "", false
	}

	switch v := frame.Error.(type) {
	case string:
		if v == "" {
			return "", false
		}
		return v, true
	case map[string]any:
		if msg, _ := v["message"].(string); msg != "" {
			return msg, true
		}
		if msg, _ := v["error"].(string); msg != "" {
			return msg, true
		}
	}

	return "", false
}
