package chat

import (
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestAccumulator(t *testing.T) {
	t.Parallel()

	t.Run("accumulates text deltas into single block", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()

		acc.handle(event{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3"}})
		acc.handle(event{Type: EventTypeTextDelta, Text: &textContent{Content: "Hello"}})
		acc.handle(event{Type: EventTypeTextDelta, Text: &textContent{Content: " world"}})
		acc.handle(event{Type: EventTypeContentBlockStop})
		acc.handle(event{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "end_turn"}})
		acc.handle(event{Done: true})

		if !acc.isDone() {
			t.Error("expected isDone")
		}

		msg := acc.message()
		if msg.Model != "claude-3" {
			t.Errorf("model = %q, want claude-3", msg.Model)
		}
		if msg.StopReason != "end_turn" {
			t.Errorf("stop_reason = %q, want end_turn", msg.StopReason)
		}

		if len(msg.Content) != 1 {
			t.Fatalf("expected 1 block, got %d", len(msg.Content))
		}
		if msg.Content[0].Type != domain.BlockTypeText {
			t.Errorf("type = %s, want text", msg.Content[0].Type)
		}
		if msg.Content[0].Text.Content != "Hello world" {
			t.Errorf("content = %q, want 'Hello world'", msg.Content[0].Text.Content)
		}
	})

	t.Run("accumulates tool_input_delta and finalizes on content_block_stop", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()

		acc.handle(event{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3"}})
		// tool_use starts with ID + name, no input
		acc.handle(event{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "tool-1", Name: "query"}})
		// input streams as deltas
		acc.handle(event{Type: EventTypeToolInputDelta, ToolInputDelta: `{"sql"`})
		acc.handle(event{Type: EventTypeToolInputDelta, ToolInputDelta: `: "SELECT 1"}`})
		// content_block_stop finalizes
		acc.handle(event{Type: EventTypeContentBlockStop})
		acc.handle(event{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use"}})
		acc.handle(event{Done: true})

		msg := acc.message()
		if len(msg.Content) != 1 {
			t.Fatalf("expected 1 block, got %d", len(msg.Content))
		}
		if msg.Content[0].Type != domain.BlockTypeToolUse {
			t.Errorf("type = %s, want tool_use", msg.Content[0].Type)
		}
		if msg.Content[0].ToolUse.ID != "tool-1" {
			t.Errorf("id = %s, want tool-1", msg.Content[0].ToolUse.ID)
		}
		if msg.Content[0].ToolUse.Name != "query" {
			t.Errorf("name = %s, want query", msg.Content[0].ToolUse.Name)
		}
		if string(msg.Content[0].ToolUse.Input) != `{"sql": "SELECT 1"}` {
			t.Errorf("input = %s, want {\"sql\": \"SELECT 1\"}", string(msg.Content[0].ToolUse.Input))
		}
	})

	t.Run("handles text then tool in sequence", func(t *testing.T) {
		t.Parallel()

		acc := newAccumulator()

		acc.handle(event{Type: EventTypeMessageStart, MessageStart: &messageStart{Model: "claude-3"}})
		acc.handle(event{Type: EventTypeTextDelta, Text: &textContent{Content: "Let me check"}})
		acc.handle(event{Type: EventTypeContentBlockStop})
		acc.handle(event{Type: EventTypeToolUse, ToolUse: &toolUseEvent{ID: "tool-1", Name: "query"}})
		acc.handle(event{Type: EventTypeToolInputDelta, ToolInputDelta: `{}`})
		acc.handle(event{Type: EventTypeContentBlockStop})
		acc.handle(event{Type: EventTypeMessageStop, MessageStop: &messageStop{StopReason: "tool_use"}})
		acc.handle(event{Done: true})

		msg := acc.message()
		if len(msg.Content) != 2 {
			t.Fatalf("expected 2 blocks, got %d", len(msg.Content))
		}
		if msg.Content[0].Type != domain.BlockTypeText {
			t.Errorf("block 0: type = %s, want text", msg.Content[0].Type)
		}
		if msg.Content[1].Type != domain.BlockTypeToolUse {
			t.Errorf("block 1: type = %s, want tool_use", msg.Content[1].Type)
		}
	})
}
