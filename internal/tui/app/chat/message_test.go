package chat_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat"
)

func TestMessage_Render(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("returns empty string for nil role", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		result := m.Render(sqlite.Message{Role: nil})

		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("renders user message with label", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "user"
		content := mustJSON([]block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: "Hello world"}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "You") {
			t.Error("expected user label 'You' in output")
		}
		if !strings.Contains(result, "Hello world") {
			t.Error("expected message content in output")
		}
	})

	t.Run("renders assistant message with label", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := mustJSON([]block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: "Hi there"}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "Tero") {
			t.Error("expected assistant label 'Tero' in output")
		}
		if !strings.Contains(result, "Hi there") {
			t.Error("expected message content in output")
		}
	})

	t.Run("renders placeholder when assistant has no content", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := "[]"

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "...") {
			t.Error("expected placeholder '...' for empty assistant message")
		}
	})

	t.Run("renders thinking block truncated", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		longThinking := strings.Repeat("x", 150) // Over 100 chars
		content := mustJSON([]block.Block{
			{Type: block.TypeThinking, Thinking: &block.Thinking{Content: longThinking}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "Thinking:") {
			t.Error("expected 'Thinking:' label in output")
		}
		if !strings.Contains(result, "...") {
			t.Error("expected truncation '...' for long thinking")
		}
		// Should not contain the full 150 chars
		if strings.Contains(result, longThinking) {
			t.Error("expected thinking to be truncated")
		}
	})

	t.Run("renders tool use with name", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := mustJSON([]block.Block{
			{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolQuery}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "Tool:") {
			t.Error("expected 'Tool:' label in output")
		}
		if !strings.Contains(result, "query") {
			t.Error("expected tool name 'query' in output")
		}
	})

	t.Run("renders tool result success", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := mustJSON([]block.Block{
			{Type: block.TypeToolResult, ToolResult: &block.ToolResult{ToolUseID: "1", IsError: false}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "Done") {
			t.Error("expected 'Done' for successful tool result")
		}
	})

	t.Run("renders tool result error", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := mustJSON([]block.Block{
			{Type: block.TypeToolResult, ToolResult: &block.ToolResult{ToolUseID: "1", IsError: true, Error: "Something broke"}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "Error:") {
			t.Error("expected 'Error:' label for failed tool result")
		}
		if !strings.Contains(result, "Something broke") {
			t.Error("expected error message in output")
		}
	})

	t.Run("renders multiple blocks in order", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "assistant"
		content := mustJSON([]block.Block{
			{Type: block.TypeThinking, Thinking: &block.Thinking{Content: "Let me think"}},
			{Type: block.TypeText, Text: &block.Text{Content: "Here is my answer"}},
			{Type: block.TypeToolUse, ToolUse: &block.ToolUse{ID: "1", Name: block.ToolShowMetric}},
		})

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		thinkingIdx := strings.Index(result, "Thinking:")
		answerIdx := strings.Index(result, "Here is my answer")
		toolIdx := strings.Index(result, "Tool:")

		if thinkingIdx == -1 || answerIdx == -1 || toolIdx == -1 {
			t.Fatal("expected all blocks to be rendered")
		}
		if thinkingIdx >= answerIdx || answerIdx >= toolIdx {
			t.Error("expected blocks to be rendered in order: thinking, text, tool")
		}
	})

	t.Run("falls back to plain text on invalid JSON", func(t *testing.T) {
		t.Parallel()

		m := chat.NewMessage(theme)
		m.SetWidth(80)

		role := "user"
		content := "not valid json"

		result := m.Render(sqlite.Message{Role: &role, Content: &content})

		if !strings.Contains(result, "not valid json") {
			t.Error("expected fallback to plain text content")
		}
	})
}

func mustJSON(blocks []block.Block) string {
	data, err := json.Marshal(blocks)
	if err != nil {
		panic(err)
	}
	return string(data)
}
