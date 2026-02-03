package chat

import (
	"errors"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestReadStream(t *testing.T) {
	t.Parallel()

	t.Run("parses text delta events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"text_delta","text":{"content":"Hello"}}
data: [DONE]
`
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if len(events) != 2 {
			t.Fatalf("got %d events, want 2", len(events))
		}

		if events[0].Type != domain.BlockTypeTextDelta {
			t.Errorf("Type = %q, want %q", events[0].Type, domain.BlockTypeTextDelta)
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
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != domain.BlockTypeThinkingDelta {
			t.Errorf("Type = %q, want %q", events[0].Type, domain.BlockTypeThinkingDelta)
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
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != domain.BlockTypeMessageStart {
			t.Errorf("Type = %q, want %q", events[0].Type, domain.BlockTypeMessageStart)
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
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != domain.BlockTypeMessageStop {
			t.Errorf("Type = %q, want %q", events[0].Type, domain.BlockTypeMessageStop)
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
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
			events = append(events, e)
			return nil
		})

		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}

		if events[0].Type != domain.BlockTypeToolUse {
			t.Errorf("Type = %q, want %q", events[0].Type, domain.BlockTypeToolUse)
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
		var events []Event
		err := readStream(strings.NewReader(stream), func(e Event) error {
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
		err := readStream(strings.NewReader(stream), func(e Event) error {
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
		err := readStream(strings.NewReader(stream), func(e Event) error {
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
}
