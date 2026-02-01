package chat_test

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/styles"
	chatui "github.com/usetero/cli/internal/tui/app/chat"
)

func TestUserMessage(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)

	t.Run("shows label and content", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewUserMessage(theme, "msg-1")
		m.SetWidth(80)
		m.SetContent([]block.Block{
			{Type: block.TypeText, Text: &block.Text{Content: "Hello world"}},
		})

		result := m.View()

		if !strings.Contains(result, "You") {
			t.Error("expected user label 'You' in output")
		}
		if !strings.Contains(result, "Hello world") {
			t.Error("expected message content in output")
		}
	})

	t.Run("never spins", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewUserMessage(theme, "msg-2")
		m.SetWidth(80)

		if m.Spinning() {
			t.Error("expected user message to never spin")
		}
	})

	t.Run("ID returns the message ID", func(t *testing.T) {
		t.Parallel()

		m := chatui.NewUserMessage(theme, "my-id")

		if m.ID() != "my-id" {
			t.Errorf("expected ID 'my-id', got %q", m.ID())
		}
	})
}
