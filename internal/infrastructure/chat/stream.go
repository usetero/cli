package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

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

func (e EventType) Valid() bool {
	switch e {
	case EventTypeMessageStart, EventTypeTextDelta, EventTypeThinkingDelta, EventTypeToolUse, EventTypeToolInputDelta, EventTypeContentBlockStop, EventTypeMessageStop, EventTypeMetadataUpdate:
		return true
	default:
		return false
	}
}

type Event struct {
	ConversationID string
	TurnID         string
	Seq            int
	Type           EventType
	Done           bool

	TextContent      string
	ThinkingContent  string
	ToolUse          *ToolUse
	ToolUseID        string
	ToolInputDelta   string
	StopReason       StopReason
	InputTokens      int
	OutputTokens     int
	Model            string
	ContextWindow    int
	ConversationName string
}

type StreamResult struct {
	ConversationID string
	TurnID         string
	LastSeq        int
}

const (
	streamDone             = "[DONE]"
	scannerInitialBuffer   = 64 * 1024
	scannerMaximumEventLen = 4 * 1024 * 1024
)

type streamHandler func(event Event) error

func readSSE(r io.Reader, handler streamHandler) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, scannerInitialBuffer), scannerMaximumEventLen)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == streamDone {
			if err := handler(Event{Done: true}); err != nil {
				return err
			}
			continue
		}
		e, err := decodeEvent([]byte(payload))
		if err != nil {
			return err
		}
		if err := handler(e); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func decodeEvent(data []byte) (Event, error) {
	var raw struct {
		ChatStreamVersion string    `json:"chat_stream_version"`
		ConversationID    string    `json:"conversation_id,omitempty"`
		TurnID            string    `json:"turn_id,omitempty"`
		Seq               int       `json:"seq,omitempty"`
		Type              EventType `json:"type"`
		Text              *struct {
			Content *string `json:"content"`
		} `json:"text,omitempty"`
		Thinking *struct {
			Content *string `json:"content"`
		} `json:"thinking,omitempty"`
		ToolUse *struct {
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"tool_use,omitempty"`
		ToolUseID      string `json:"tool_use_id,omitempty"`
		ToolInputDelta string `json:"tool_input_delta,omitempty"`
		MessageStart   *struct {
			Model         string `json:"model"`
			ContextWindow *int   `json:"context_window"`
		} `json:"message_start,omitempty"`
		MessageStop *struct {
			StopReason   StopReason `json:"stop_reason"`
			InputTokens  *int       `json:"input_tokens"`
			OutputTokens *int       `json:"output_tokens"`
		} `json:"message_stop,omitempty"`
		Metadata *struct {
			Title string `json:"title"`
		} `json:"metadata,omitempty"`
		Error any `json:"error,omitempty"`
	}
	if err := strictUnmarshal(data, &raw); err != nil {
		return Event{}, fmt.Errorf("parse event: %w", err)
	}
	if raw.ChatStreamVersion != protocolVersion {
		return Event{}, fmt.Errorf("protocol error: unsupported chat_stream_version %q", raw.ChatStreamVersion)
	}
	if msg, ok := parseErrorMessage(raw.Error); ok {
		return Event{}, fmt.Errorf("server error: %s", msg)
	}
	if !raw.Type.Valid() {
		if raw.Type == "" {
			return Event{}, fmt.Errorf("protocol error: missing event type")
		}
		return Event{}, fmt.Errorf("protocol error: unsupported event type %q", raw.Type)
	}

	e := Event{ConversationID: raw.ConversationID, TurnID: raw.TurnID, Seq: raw.Seq, Type: raw.Type, ToolUseID: raw.ToolUseID, ToolInputDelta: raw.ToolInputDelta}
	if raw.Text != nil && raw.Text.Content != nil {
		e.TextContent = *raw.Text.Content
	}
	if raw.Thinking != nil && raw.Thinking.Content != nil {
		e.ThinkingContent = *raw.Thinking.Content
	}
	if raw.ToolUse != nil {
		e.ToolUse = &ToolUse{ID: raw.ToolUse.ID, Name: raw.ToolUse.Name, Input: raw.ToolUse.Input}
	}
	if raw.MessageStart != nil {
		e.Model = raw.MessageStart.Model
		if raw.MessageStart.ContextWindow != nil {
			e.ContextWindow = *raw.MessageStart.ContextWindow
		}
	}
	if raw.MessageStop != nil {
		if !raw.MessageStop.StopReason.Valid() {
			return Event{}, fmt.Errorf("protocol error: invalid stop_reason %q", raw.MessageStop.StopReason)
		}
		e.StopReason = raw.MessageStop.StopReason
		if raw.MessageStop.InputTokens != nil {
			e.InputTokens = *raw.MessageStop.InputTokens
		}
		if raw.MessageStop.OutputTokens != nil {
			e.OutputTokens = *raw.MessageStop.OutputTokens
		}
	}
	if raw.Metadata != nil {
		e.ConversationName = raw.Metadata.Title
	}
	return e, nil
}

func strictUnmarshal(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	var trailing struct{}
	if err := dec.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected trailing json data")
	}
	return nil
}

func parseErrorMessage(raw any) (string, bool) {
	if raw == nil {
		return "", false
	}
	switch v := raw.(type) {
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
