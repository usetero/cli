package block_test

import (
	"encoding/json"
	"testing"

	"github.com/usetero/cli/internal/chat/block"
)

func TestAccumulator_Apply(t *testing.T) {
	t.Parallel()

	t.Run("returns true when state changes", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()

		if !acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "hello"}}) {
			t.Error("Apply() returned false for text delta, want true")
		}
	})

	t.Run("returns false for unknown block type", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()

		if acc.Apply(block.Block{Type: "unknown"}) {
			t.Error("Apply() returned true for unknown type, want false")
		}
	})

	t.Run("returns false for nil content", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()

		if acc.Apply(block.Block{Type: block.TypeTextDelta, Text: nil}) {
			t.Error("Apply() returned true for nil text, want false")
		}
	})
}

func TestAccumulator_TextDeltas(t *testing.T) {
	t.Parallel()

	t.Run("merges consecutive text deltas", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "Hello"}})
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: " world"}})
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "!"}})

		blocks := acc.Blocks()
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
		if blocks[0].Text.Content != "Hello world!" {
			t.Errorf("text content = %q, want %q", blocks[0].Text.Content, "Hello world!")
		}
	})

	t.Run("stores as TypeText not TypeTextDelta", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "test"}})

		blocks := acc.Blocks()
		if blocks[0].Type != block.TypeText {
			t.Errorf("block type = %q, want %q", blocks[0].Type, block.TypeText)
		}
	})
}

func TestAccumulator_ThinkingDeltas(t *testing.T) {
	t.Parallel()

	t.Run("merges consecutive thinking deltas", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "Let me "}})
		acc.Apply(block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "think..."}})

		blocks := acc.Blocks()
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
		if blocks[0].Thinking.Content != "Let me think..." {
			t.Errorf("thinking content = %q, want %q", blocks[0].Thinking.Content, "Let me think...")
		}
	})

	t.Run("stores as TypeThinking not TypeThinkingDelta", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "test"}})

		blocks := acc.Blocks()
		if blocks[0].Type != block.TypeThinking {
			t.Errorf("block type = %q, want %q", blocks[0].Type, block.TypeThinking)
		}
	})
}

func TestAccumulator_MixedBlocks(t *testing.T) {
	t.Parallel()

	t.Run("creates new block when type changes", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeThinkingDelta, Thinking: &block.Thinking{Content: "thinking..."}})
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "response"}})

		blocks := acc.Blocks()
		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		if blocks[0].Type != block.TypeThinking {
			t.Errorf("first block type = %q, want %q", blocks[0].Type, block.TypeThinking)
		}
		if blocks[1].Type != block.TypeText {
			t.Errorf("second block type = %q, want %q", blocks[1].Type, block.TypeText)
		}
	})

	t.Run("tool use creates new block", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "Let me check..."}})
		acc.Apply(block.Block{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolQuery}})

		blocks := acc.Blocks()
		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		if blocks[1].ToolUse.Name != block.ToolQuery {
			t.Errorf("tool name = %q, want %q", blocks[1].ToolUse.Name, block.ToolQuery)
		}
	})
}

func TestAccumulator_ToolInputDeltas(t *testing.T) {
	t.Parallel()

	t.Run("appends input to current tool use", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolQuery}})
		acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `{"sql":`})
		acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `"SELECT * FROM users"}`})

		blocks := acc.Blocks()
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
		if blocks[0].ToolUse.RawInput != `{"sql":"SELECT * FROM users"}` {
			t.Errorf("raw input = %q, want %q", blocks[0].ToolUse.RawInput, `{"sql":"SELECT * FROM users"}`)
		}
	})

	t.Run("tool use IsComplete returns false during streaming", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolQuery}})
		acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `{"sql":"test"}`})

		blocks := acc.Blocks()
		// IsComplete checks typed inputs, not RawInput
		if blocks[0].ToolUse.IsComplete() {
			t.Error("IsComplete() returned true during streaming, want false")
		}
	})

	t.Run("ignores input delta when no tool block exists", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		// This shouldn't happen in practice, but should not panic
		changed := acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `{"test":true}`})

		if changed {
			t.Error("Apply() returned true for orphan input delta, want false")
		}
	})

	t.Run("multiple tools accumulate inputs separately", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolQuery}})
		acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `{"sql":"first"}`})
		acc.Apply(block.Block{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "2", Name: block.ToolShowMetric}})
		acc.Apply(block.Block{Type: block.TypeToolInputDelta, ToolInputDelta: `{"label":"test"}`})

		blocks := acc.Blocks()
		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		if blocks[0].ToolUse.RawInput != `{"sql":"first"}` {
			t.Errorf("first tool raw input = %q, want %q", blocks[0].ToolUse.RawInput, `{"sql":"first"}`)
		}
		if blocks[1].ToolUse.RawInput != `{"label":"test"}` {
			t.Errorf("second tool raw input = %q, want %q", blocks[1].ToolUse.RawInput, `{"label":"test"}`)
		}
	})
}

func TestAccumulator_JSON(t *testing.T) {
	t.Parallel()

	t.Run("returns empty array when no blocks", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		if acc.JSON() != "[]" {
			t.Errorf("JSON() = %q, want %q", acc.JSON(), "[]")
		}
	})

	t.Run("returns valid JSON array", func(t *testing.T) {
		t.Parallel()

		acc := block.NewAccumulator()
		acc.Apply(block.Block{Type: block.TypeTextDelta, Text: &block.Text{Content: "hello"}})

		var blocks []block.Block
		if err := json.Unmarshal([]byte(acc.JSON()), &blocks); err != nil {
			t.Fatalf("JSON() produced invalid JSON: %v", err)
		}
		if len(blocks) != 1 {
			t.Errorf("unmarshaled %d blocks, want 1", len(blocks))
		}
	})
}
