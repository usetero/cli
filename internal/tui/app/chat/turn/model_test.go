package turn_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/turn"
)

func TestModel_Send(t *testing.T) {
	// Not parallel - thinking.New() uses a package-level ID counter
	theme := styles.NewTheme(true)

	t.Run("returns StreamDoneMsg on success", func(t *testing.T) {

		client := &chattest.MockClient{
			StreamFunc: func(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) error {
				onMessage(&domain.Message{
					Role:  domain.RoleAssistant,
					Model: "test-model",
					Content: []domain.Block{
						{Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "Hello"}},
					},
					StopReason: "end_turn",
				})
				return nil
			},
		}

		m := turn.New(theme, client, logtest.New(t))
		cmd := m.Send(context.Background(), "conv-1", nil, nil)

		msg := executeBatchCmd(t, cmd)
		done, ok := msg.(turn.StreamDoneMsg)
		if !ok {
			t.Fatalf("expected StreamDoneMsg, got %T", msg)
		}

		if done.Err != nil {
			t.Fatalf("unexpected error: %v", done.Err)
		}
		if done.StopReason != "end_turn" {
			t.Errorf("stop_reason = %s, want end_turn", done.StopReason)
		}
		if done.Message.Role != domain.RoleAssistant {
			t.Errorf("role = %s, want assistant", done.Message.Role)
		}
	})

	t.Run("returns error on client failure", func(t *testing.T) {

		expectedErr := errors.New("connection failed")
		client := &chattest.MockClient{
			StreamFunc: func(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) error {
				return expectedErr
			},
		}

		m := turn.New(theme, client, logtest.New(t))
		cmd := m.Send(context.Background(), "conv-1", nil, nil)

		msg := executeBatchCmd(t, cmd)
		done, ok := msg.(turn.StreamDoneMsg)
		if !ok {
			t.Fatalf("expected StreamDoneMsg, got %T", msg)
		}

		if done.Err == nil {
			t.Fatal("expected error")
		}
		if done.Err.Error() != expectedErr.Error() {
			t.Errorf("error = %v, want %v", done.Err, expectedErr)
		}
	})

	t.Run("returns tool_use stop reason", func(t *testing.T) {

		client := &chattest.MockClient{
			StreamFunc: func(ctx context.Context, req chat.Request, onMessage func(*domain.Message)) error {
				onMessage(&domain.Message{
					Role:  domain.RoleAssistant,
					Model: "test-model",
					Content: []domain.Block{
						{Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-1", Name: "query"}},
					},
					StopReason: "tool_use",
				})
				return nil
			},
		}

		m := turn.New(theme, client, logtest.New(t))
		cmd := m.Send(context.Background(), "conv-1", nil, nil)

		msg := executeBatchCmd(t, cmd)
		done, ok := msg.(turn.StreamDoneMsg)
		if !ok {
			t.Fatalf("expected StreamDoneMsg, got %T", msg)
		}

		if done.StopReason != "tool_use" {
			t.Errorf("stop_reason = %s, want tool_use", done.StopReason)
		}
		if len(done.Message.Content) != 1 {
			t.Fatalf("expected 1 block, got %d", len(done.Message.Content))
		}
		if done.Message.Content[0].ToolUse == nil {
			t.Error("expected tool_use block")
		}
	})
}

// executeBatchCmd executes a batch command and returns the first StreamDoneMsg found.
func executeBatchCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()

	result := cmd()

	// tea.BatchMsg is []tea.Cmd
	if batch, ok := result.(tea.BatchMsg); ok {
		for _, fn := range batch {
			if fn == nil {
				continue
			}
			if msg := fn(); msg != nil {
				if _, ok := msg.(turn.StreamDoneMsg); ok {
					return msg
				}
			}
		}
	}

	return result
}
