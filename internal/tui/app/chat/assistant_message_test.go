package chat_test

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
	chatui "github.com/usetero/cli/internal/tui/app/chat"
)

func TestAssistantMessage(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("starts in sending state", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessage(theme)
		m.SetWidth(80)

		if m.State() != chatui.StateSending {
			t.Errorf("expected StateSending, got %v", m.State())
		}
		if !m.Spinning() {
			t.Error("expected message to be spinning in sending state")
		}
		if m.ID() != "" {
			t.Error("expected empty ID before SetMessageID")
		}
	})

	t.Run("transitions to thinking state on SetMessageID", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessage(theme)
		m.SetWidth(80)
		m.SetMessageID("msg-123")

		if m.State() != chatui.StateThinking {
			t.Errorf("expected StateThinking, got %v", m.State())
		}
		if m.ID() != "msg-123" {
			t.Errorf("expected ID 'msg-123', got %q", m.ID())
		}
		if !m.Spinning() {
			t.Error("expected message to be spinning in thinking state")
		}
	})

	t.Run("transitions to ready state on content", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessage(theme)
		m.SetWidth(80)
		m.SetMessageID("msg-123")
		m.SetContent([]block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: "Hello"}},
		})

		if m.State() != chatui.StateReady {
			t.Errorf("expected StateReady, got %v", m.State())
		}
		if m.Spinning() {
			t.Error("expected message to stop spinning in ready state")
		}

		result := m.View()
		if !strings.Contains(result, "Hello") {
			t.Error("expected content in view")
		}
	})

	t.Run("shows label in all states", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessage(theme)
		m.SetWidth(80)

		result := m.View()
		if !strings.Contains(result, "Tero") {
			t.Error("expected 'Tero' label in output")
		}
	})

	t.Run("created with ID starts in thinking state", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessageWithID(theme, "existing-id")
		m.SetWidth(80)

		if m.State() != chatui.StateThinking {
			t.Errorf("expected StateThinking, got %v", m.State())
		}
		if m.ID() != "existing-id" {
			t.Errorf("expected ID 'existing-id', got %q", m.ID())
		}
	})

	t.Run("renders thinking blocks truncated", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewAssistantMessageWithID(theme, "msg-1")
		m.SetWidth(80)

		longThinking := strings.Repeat("x", 150)
		m.SetContent([]block.Block{
			{Type: block.TypeThinking, Thinking: &block.Thinking{Content: longThinking}},
			{Type: block.TypeText, Text: &block.Text{Content: "done"}},
		})

		result := m.View()

		if !strings.Contains(result, "Thinking:") {
			t.Error("expected 'Thinking:' label in output")
		}
		if !strings.Contains(result, "...") {
			t.Error("expected truncation '...' for long thinking")
		}
	})
}
