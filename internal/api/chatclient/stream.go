package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const streamDone = "[DONE]"

const (
	streamScannerInitialBuffer = 64 * 1024
	streamScannerMaxBuffer     = 4 * 1024 * 1024
)

type eventHandler func(event) error

func readStream(r io.Reader, handler eventHandler) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, streamScannerInitialBuffer), streamScannerMaxBuffer)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == streamDone {
			if err := handler(event{Done: true}); err != nil {
				return err
			}
			continue
		}

		e, err := decodeEventData([]byte(data))
		if err != nil {
			return err
		}
		if err := handler(e); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func decodeEventData(data []byte) (event, error) {
	var header struct {
		ChatStreamVersion string    `json:"chat_stream_version"`
		Type              EventType `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return event{}, fmt.Errorf("parse event: %w", err)
	}
	if header.ChatStreamVersion != chatStreamVersionV2 {
		return event{}, fmt.Errorf("protocol error: unsupported chat_stream_version %q", header.ChatStreamVersion)
	}
	if msg, ok := parseErrorMessage(data); ok {
		return event{}, fmt.Errorf("server error: %s", msg)
	}

	var e event
	if err := strictUnmarshal(data, &e); err != nil {
		return event{}, fmt.Errorf("parse event: %w", err)
	}
	if e.Type == "" {
		return event{}, fmt.Errorf("protocol error: missing event type")
	}

	switch e.Type {
	case EventTypeMessageStart:
		if e.MessageStart == nil || e.MessageStart.Model == "" || e.MessageStart.ContextWindow == nil {
			return event{}, fmt.Errorf("protocol error: message_start missing required fields")
		}
	case EventTypeTextDelta:
		if e.Text == nil || e.Text.Content == nil {
			return event{}, fmt.Errorf("protocol error: text_delta missing content")
		}
	case EventTypeThinkingDelta:
		if e.Thinking == nil || e.Thinking.Content == nil {
			return event{}, fmt.Errorf("protocol error: thinking_delta missing content")
		}
	case EventTypeToolUse:
		if e.ToolUse == nil || e.ToolUse.ID == "" || e.ToolUse.Name == "" {
			return event{}, fmt.Errorf("protocol error: tool_use missing id/name")
		}
	case EventTypeToolInputDelta:
		if e.ToolUseID == "" {
			return event{}, fmt.Errorf("protocol error: tool_input_delta missing tool_use_id")
		}
	case EventTypeContentBlockStop:
		// tool_use_id is optional and required only for tool blocks.
	case EventTypeMessageStop:
		if e.MessageStop == nil || e.MessageStop.StopReason == "" || e.MessageStop.InputTokens == nil || e.MessageStop.OutputTokens == nil {
			return event{}, fmt.Errorf("protocol error: message_stop missing required fields")
		}
	case EventTypeMetadataUpdate:
		if e.Metadata == nil || e.Metadata.Title == "" {
			return event{}, fmt.Errorf("protocol error: metadata_update missing title")
		}
	default:
		return event{}, fmt.Errorf("protocol error: unknown event type %q", e.Type)
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
		return fmt.Errorf("unexpected trailing JSON data")
	}
	return nil
}
