package chat

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestAccumulator(t *testing.T) {
	t.Parallel()

	t.Run("happy path text-only stream", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()
		events := []event{
			{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}},
			{Type: EventTypeTextDelta, Text: &textContent{Content: strPtr("Hello")}},
			{Type: EventTypeTextDelta, Text: &textContent{Content: strPtr(" world")}},
			{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn", InputTokens: intPtr(9), OutputTokens: intPtr(2)}},
			{Done: true},
		}
		for _, e := range events {
			if err := acc.handle(e); err != nil {
				t.Fatalf("handle(%s) error = %v", e.Type, err)
			}
		}

		msg := acc.message()
		if msg.Model != "claude-3" {
			t.Fatalf("model = %q", msg.Model)
		}
		if msg.StopReason != "end_turn" {
			t.Fatalf("stop_reason = %q", msg.StopReason)
		}
		if len(msg.Content) != 1 || msg.Content[0].Type != domain.BlockTypeText {
			t.Fatalf("unexpected content: %#v", msg.Content)
		}
		if msg.Content[0].Text.Content != "Hello world" {
			t.Fatalf("text = %q", msg.Content[0].Text.Content)
		}
	})

	t.Run("single tool call", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()
		events := []event{
			{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}},
			{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "tool-1", Name: "query"}},
			{Type: EventTypeToolInputDelta, ToolUseID: "tool-1", ToolInputDelta: `{"sql":`},
			{Type: EventTypeToolInputDelta, ToolUseID: "tool-1", ToolInputDelta: `"SELECT 1"}`},
			{Type: EventTypeContentBlockStop, ToolUseID: "tool-1"},
			{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use", InputTokens: intPtr(10), OutputTokens: intPtr(2)}},
			{Done: true},
		}
		for _, e := range events {
			if err := acc.handle(e); err != nil {
				t.Fatalf("handle(%s) error = %v", e.Type, err)
			}
		}

		msg := acc.message()
		if len(msg.Content) != 1 || msg.Content[0].Type != domain.BlockTypeToolUse {
			t.Fatalf("unexpected content: %#v", msg.Content)
		}
		if string(msg.Content[0].ToolUse.Input) != `{"sql":"SELECT 1"}` {
			t.Fatalf("input = %s", string(msg.Content[0].ToolUse.Input))
		}
		if !msg.Content[0].ToolUse.InputComplete {
			t.Fatal("expected InputComplete=true")
		}
	})

	t.Run("multiple interleaved tool calls", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()
		events := []event{
			{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}},
			{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "a", Name: "query"}},
			{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "b", Name: "query"}},
			{Type: EventTypeToolInputDelta, ToolUseID: "a", ToolInputDelta: `{"sql":"SELECT `},
			{Type: EventTypeToolInputDelta, ToolUseID: "b", ToolInputDelta: `{"sql":"SELECT `},
			{Type: EventTypeToolInputDelta, ToolUseID: "a", ToolInputDelta: `1"}`},
			{Type: EventTypeToolInputDelta, ToolUseID: "b", ToolInputDelta: `2"}`},
			{Type: EventTypeContentBlockStop, ToolUseID: "b"},
			{Type: EventTypeContentBlockStop, ToolUseID: "a"},
			{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use", InputTokens: intPtr(10), OutputTokens: intPtr(2)}},
			{Done: true},
		}
		for _, e := range events {
			if err := acc.handle(e); err != nil {
				t.Fatalf("handle(%s) error = %v", e.Type, err)
			}
		}

		msg := acc.message()
		if len(msg.Content) != 2 {
			t.Fatalf("blocks = %d, want 2", len(msg.Content))
		}
		if string(msg.Content[0].ToolUse.Input) != `{"sql":"SELECT 1"}` {
			t.Fatalf("tool a input = %s", string(msg.Content[0].ToolUse.Input))
		}
		if string(msg.Content[1].ToolUse.Input) != `{"sql":"SELECT 2"}` {
			t.Fatalf("tool b input = %s", string(msg.Content[1].ToolUse.Input))
		}
	})

	t.Run("malformed ordering errors", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()
		if err := acc.handle(event{Type: EventTypeToolInputDelta, ToolUseID: "missing", ToolInputDelta: "{}"}); err == nil {
			t.Fatal("expected unknown tool error")
		}

		acc = newAccumulator()
		_ = acc.handle(event{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		if err := acc.handle(event{Type: EventTypeContentBlockStop, ToolUseID: "missing"}); err == nil {
			t.Fatal("expected unknown stop id error")
		}
	})

	t.Run("rejects invalid tool input JSON", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()
		_ = acc.handle(event{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3", ContextWindow: intPtr(200000)}})
		_ = acc.handle(event{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "tool-1", Name: "query"}})
		_ = acc.handle(event{Type: EventTypeToolInputDelta, ToolUseID: "tool-1", ToolInputDelta: `{`})
		err := acc.handle(event{Type: EventTypeContentBlockStop, ToolUseID: "tool-1"})
		if err == nil || !strings.Contains(err.Error(), "invalid JSON") {
			t.Fatalf("error = %v, want invalid JSON", err)
		}
	})
}
