package chat

import (
	"encoding/json"
	"testing"

	"github.com/usetero/cli/internal/domain"
)

func TestAccumulator_TextDeltas(t *testing.T) {
	t.Parallel()

	acc := NewAccumulator()

	// Stream: MessageStart, TextDelta, TextDelta, TextDelta, MessageStop, Done
	acc.Handle(Event{Block: domain.Block{
		Type:         domain.BlockTypeMessageStart,
		MessageStart: &domain.MessageStart{Model: "claude-3"},
	}})

	// First delta
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeTextDelta,
		Text: &domain.TextBlock{Content: "Hello"},
	}})

	blocks := acc.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Text.Content != "Hello" {
		t.Errorf("expected 'Hello', got %q", blocks[0].Text.Content)
	}

	// Second delta
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeTextDelta,
		Text: &domain.TextBlock{Content: " world"},
	}})

	blocks = acc.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Text.Content != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", blocks[0].Text.Content)
	}

	// MessageStop
	acc.Handle(Event{Block: domain.Block{
		Type:        domain.BlockTypeMessageStop,
		MessageStop: &domain.MessageStop{StopReason: "end_turn"},
	}})

	// Done
	acc.Handle(Event{Done: true})

	if !acc.Done() {
		t.Error("expected Done() to be true")
	}
	if acc.Model() != "claude-3" {
		t.Errorf("expected model 'claude-3', got %q", acc.Model())
	}
	if acc.StopReason() != "end_turn" {
		t.Errorf("expected stop_reason 'end_turn', got %q", acc.StopReason())
	}

	blocks = acc.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != domain.BlockTypeText {
		t.Errorf("expected text block, got %s", blocks[0].Type)
	}
}

func TestAccumulator_MixedBlocks(t *testing.T) {
	t.Parallel()

	acc := NewAccumulator()

	// Thinking first
	acc.Handle(Event{Block: domain.Block{
		Type:     domain.BlockTypeThinkingDelta,
		Thinking: &domain.Thinking{Content: "Let me think..."},
	}})

	// Then text
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeTextDelta,
		Text: &domain.TextBlock{Content: "Here's my answer"},
	}})

	blocks := acc.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].Type != domain.BlockTypeThinking {
		t.Errorf("expected thinking block first, got %s", blocks[0].Type)
	}
	if blocks[0].Thinking.Content != "Let me think..." {
		t.Errorf("unexpected thinking content: %q", blocks[0].Thinking.Content)
	}

	if blocks[1].Type != domain.BlockTypeText {
		t.Errorf("expected text block second, got %s", blocks[1].Type)
	}
	if blocks[1].Text.Content != "Here's my answer" {
		t.Errorf("unexpected text content: %q", blocks[1].Text.Content)
	}
}

func TestAccumulator_ToolUse(t *testing.T) {
	t.Parallel()

	acc := NewAccumulator()

	// Text first
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeTextDelta,
		Text: &domain.TextBlock{Content: "Let me check that"},
	}})

	// Tool use (comes as complete block)
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeToolUse,
		ToolUse: &domain.ToolUse{
			ID:    "tool-1",
			Name:  "query",
			Input: json.RawMessage(`{"limit": 10}`),
		},
	}})

	blocks := acc.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].Type != domain.BlockTypeText {
		t.Errorf("expected text block first, got %s", blocks[0].Type)
	}
	if blocks[1].Type != domain.BlockTypeToolUse {
		t.Errorf("expected tool_use block second, got %s", blocks[1].Type)
	}
}

func TestAccumulator_Reset(t *testing.T) {
	t.Parallel()

	acc := NewAccumulator()

	acc.Handle(Event{Block: domain.Block{
		Type:         domain.BlockTypeMessageStart,
		MessageStart: &domain.MessageStart{Model: "claude-3"},
	}})
	acc.Handle(Event{Block: domain.Block{
		Type: domain.BlockTypeTextDelta,
		Text: &domain.TextBlock{Content: "Hello"},
	}})
	acc.Handle(Event{Done: true})

	acc.Reset()

	if acc.Done() {
		t.Error("expected Done() to be false after Reset")
	}
	if acc.Model() != "" {
		t.Errorf("expected empty model after Reset, got %q", acc.Model())
	}
	if len(acc.Blocks()) != 0 {
		t.Errorf("expected no blocks after Reset, got %d", len(acc.Blocks()))
	}
}
