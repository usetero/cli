package chat

import (
	"errors"
	"strings"
	"testing"

	corechat "github.com/usetero/cli/internal/core/chat"
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
		machine := corechat.NewStreamMachine("conv-1")
		var snapshots []corechat.StreamSnapshot
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			var (
				snap *corechat.StreamSnapshot
				err  error
			)
			if done {
				snap, err = machine.ConsumeDone()
			} else {
				snap, err = machine.ConsumeData(data)
			}
			if err != nil {
				return err
			}
			snapshots = append(snapshots, *snap)
			return nil
		})
		if err != nil {
			t.Fatalf("readStream() error = %v", err)
		}
		if len(snapshots) != 4 {
			t.Fatalf("snapshots = %d, want 4", len(snapshots))
		}
		if snapshots[0].Status != corechat.StreamStatusStreaming {
			t.Fatalf("snapshot[0].Status = %q", snapshots[0].Status)
		}
		if snapshots[1].Status != corechat.StreamStatusStreaming {
			t.Fatalf("snapshot[1].Status = %q", snapshots[1].Status)
		}
		if snapshots[2].Status != corechat.StreamStatusStreaming {
			t.Fatalf("snapshot[2].Status = %q", snapshots[2].Status)
		}
		if !snapshots[3].Done {
			t.Fatal("snapshot[3].Done = false, want true")
		}
	})

	t.Run("stops when handler returns error for error frame", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
data: {"chat_stream_version":"v2","error":"internal error"}
data: [DONE]
`
		calls := 0
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			calls++
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "server error: internal error") {
			t.Fatalf("error = %q", err.Error())
		}
		if calls != 2 {
			t.Fatalf("handler calls = %d, want 2", calls)
		}
	})

	t.Run("rejects unknown event types", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"bogus"}
`
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
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
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
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
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
		if err == nil || !strings.Contains(err.Error(), "chat_stream_version") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects mismatched protocol version", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v1","type":"message_start","message_start":{"model":"claude-3","context_window":200000}}
`
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
		if err == nil || !strings.Contains(err.Error(), "unsupported chat_stream_version") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rejects message_stop without token fields", func(t *testing.T) {
		t.Parallel()

		stream := `data: {"chat_stream_version":"v2","type":"message_stop","message_stop":{"stop_reason":"end_turn"}}
`
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
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
		machine := corechat.NewStreamMachine("conv-1")
		err := readStream(strings.NewReader(stream), func(data []byte, done bool) error {
			calls++
			if calls == 2 {
				return handlerErr
			}
			if done {
				_, err := machine.ConsumeDone()
				return err
			}
			_, err := machine.ConsumeData(data)
			return err
		})
		if !errors.Is(err, handlerErr) {
			t.Fatalf("error = %v, want %v", err, handlerErr)
		}
	})
}
