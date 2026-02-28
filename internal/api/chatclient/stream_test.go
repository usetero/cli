package chat

import (
	"errors"
	"strings"
	"testing"
)

func TestReadStream(t *testing.T) {
	t.Parallel()

	t.Run("decodes happy-path v2 events", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
data: {"chat_stream_version":"v2","type":"text_delta","text":{"content":"Hello"}}
data: {"chat_stream_version":"v2","type":"message_stop","message_stop":{"stop_reason":"end_turn","input_tokens":1,"output_tokens":1}}
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
		if len(events) != 4 {
			t.Fatalf("events = %d, want 4", len(events))
		}
		if events[0].Type != EventTypeMessageStart {
			t.Fatalf("event[0].Type = %q", events[0].Type)
		}
		if events[1].Type != EventTypeTextDelta {
			t.Fatalf("event[1].Type = %q", events[1].Type)
		}
		if events[2].Type != EventTypeMessageStop {
			t.Fatalf("event[2].Type = %q", events[2].Type)
		}
		if !events[3].Done {
			t.Fatal("event[3].Done = false, want true")
		}
	})

	t.Run("terminates immediately on error frame", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
data: {"chat_stream_version":"v2","error":"internal error"}
data: [DONE]
`
		calls := 0
		err := readStream(strings.NewReader(stream), func(e event) error {
			calls++
			return nil
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "server error: internal error") {
			t.Fatalf("error = %q", err.Error())
		}
		if calls != 1 {
			t.Fatalf("handler calls = %d, want 1", calls)
		}
	})

	t.Run("rejects unknown event types", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"bogus"}
`
		err := readStream(strings.NewReader(stream), func(e event) error { return nil })
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown event type") {
			t.Fatalf("error = %q", err.Error())
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"text_delta","text":{"content":"x"},"unexpected":1}
`
		err := readStream(strings.NewReader(stream), func(e event) error { return nil })
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("error = %q", err.Error())
		}
	})

	t.Run("rejects missing protocol version", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
`
		err := readStream(strings.NewReader(stream), func(e event) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "chat_stream_version") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects mismatched protocol version", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v1","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
`
		err := readStream(strings.NewReader(stream), func(e event) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "unsupported chat_stream_version") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects message_stop without token fields", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_stop","message_stop":{"stop_reason":"end_turn"}}
`
		err := readStream(strings.NewReader(stream), func(e event) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "message_stop missing required fields") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("stops on handler error", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
data: {"chat_stream_version":"v2","type":"text_delta","text":{"content":"x"}}
`
		handlerErr := errors.New("stop")
		calls := 0
		err := readStream(strings.NewReader(stream), func(e event) error {
			calls++
			if calls == 2 {
				return handlerErr
			}
			return nil
		})
		if !errors.Is(err, handlerErr) {
			t.Fatalf("error = %v, want %v", err, handlerErr)
		}
	})
}
