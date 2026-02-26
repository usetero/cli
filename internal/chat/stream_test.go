package chat

import (
	"errors"
	"strings"
	"testing"
)

func TestReadStream(t *testing.T) {
	t.Parallel()

	t.Run("parses text delta events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"text_delta","text":{"content":"Hello"}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}

		if events[0].Type != EventTypeTextDelta {
			t.Errorf("Type = %q, want %q", events[0].Type, EventTypeTextDelta)
		}
		if events[0].Text == nil || events[0].Text.Content != "Hello" {
			t.Errorf("Text.Content = %v, want 'Hello'", events[0].Text)
		}
		if !events[1].Done {
			t.Error("last event Done = false, want true")
		}
	})

	t.Run("parses thinking delta events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"thinking_delta","thinking":{"content":"Let me think..."}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != EventTypeThinkingDelta {
			t.Errorf("Type = %q, want %q", events[0].Type, EventTypeThinkingDelta)
		}
		if events[0].Thinking == nil || events[0].Thinking.Content != "Let me think..." {
			t.Errorf("Thinking.Content = %v, want 'Let me think...'", events[0].Thinking)
		}
	})

	t.Run("parses message_start events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_start","message_start":{"model":"claude-3"}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != EventTypeMessageStart {
			t.Errorf("Type = %q, want %q", events[0].Type, EventTypeMessageStart)
		}
		if events[0].MessageStart == nil || events[0].MessageStart.Model != "claude-3" {
			t.Errorf("MessageStart.Model = %v, want 'claude-3'", events[0].MessageStart)
		}
	})

	t.Run("parses message_stop events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_stop","message_stop":{"stop_reason":"end_turn"}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != EventTypeMessageStop {
			t.Errorf("Type = %q, want %q", events[0].Type, EventTypeMessageStop)
		}
		if events[0].MessageStop == nil || events[0].MessageStop.StopReason != "end_turn" {
			t.Errorf("MessageStop.StopReason = %v, want 'end_turn'", events[0].MessageStop)
		}
	})

	t.Run("parses tool_use events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"tool_use","tool_use":{"id":"tool-1","name":"get_weather"}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != EventTypeToolUse {
			t.Errorf("Type = %q, want %q", events[0].Type, EventTypeToolUse)
		}
		if events[0].ToolUse == nil || events[0].ToolUse.Name != "get_weather" {
			t.Errorf("ToolUse.Name = %v, want 'get_weather'", events[0].ToolUse)
		}
	})

	t.Run("skips empty lines and comments", func(t *testing.T) {
		t.Parallel()

		stream := `: this is a comment

data: {"type":"text_delta","text":{"content":"Hi"}}

: another comment
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		t.Parallel()

		stream := `data: {invalid json}
`
		err := readStream(strings.NewReader(stream), func(e event) error {
			return nil
		})

		if err == nil {
			t.Fatal("readStream() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "parse event") {
			t.Errorf("error = %q, want to contain 'parse event'", err.Error())
		}
	})

	t.Run("stops on handler error", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"text_delta","text":{"content":"1"}}
data: {"type":"text_delta","text":{"content":"2"}}
data: {"type":"text_delta","text":{"content":"3"}}
data: [DONE]
`
		handlerErr := errors.New("stop")
		callCount := 0
		err := readStream(strings.NewReader(stream), func(e event) error {
			callCount++
			if callCount == 2 {
				return handlerErr
			}
			return nil
		})

		if !errors.Is(err, handlerErr) {
			t.Errorf("readStream() error = %v, want %v", err, handlerErr)
		}
		if callCount != 2 {
			t.Errorf("handler called %d times, want 2", callCount)
		}
	})

	t.Run("handles large event payloads over default scanner limit", func(t *testing.T) {
		t.Parallel()

		large := strings.Repeat("a", 70*1024)
		stream := `data: {"type":"text_delta","text":{"content":"` + large + `"}}
data: [DONE]
`
		var events []event
		err := readStream(strings.NewReader(stream), func(e event) error {
			events = append(events, e)
			return nil
		})
		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}
		if events[0].Text == nil || len(events[0].Text.Content) != len(large) {
			t.Fatalf("unexpected payload length: got %d, want %d", len(events[0].Text.Content), len(large))
		}
	})
}
