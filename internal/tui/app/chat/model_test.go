package chat

import (
	"context"
	"testing"

	"github.com/usetero/cli/internal/chat/chattest"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/sqlite/sqlitetest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	"github.com/usetero/cli/internal/tui/app/chat/turn"
	apptools "github.com/usetero/cli/internal/tui/app/tools"
)

func TestModel_handleStreamDone(t *testing.T) {
	t.Parallel()

	t.Run("transitions to idle on end_turn", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		m.state = StateStreaming

		m, _ = m.handleStreamDone(turn.StreamDoneMsg{
			Message: domain.Message{
				ID:   "msg-1",
				Role: domain.RoleAssistant,
				Content: []domain.Block{
					{Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "Hello"}},
				},
			},
			StopReason: "end_turn",
		})

		if m.state != StateIdle {
			t.Errorf("state = %v, want StateIdle", m.state)
		}
		if len(m.rawMessages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(m.rawMessages))
		}
	})

	t.Run("transitions to awaiting tools on tool_use", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		m.state = StateStreaming

		m, _ = m.handleStreamDone(turn.StreamDoneMsg{
			Message: domain.Message{
				ID:   "msg-1",
				Role: domain.RoleAssistant,
				Content: []domain.Block{
					{Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-1", Name: "unknown_tool"}},
				},
			},
			StopReason: "tool_use",
		})

		if m.state != StateAwaitingTools {
			t.Errorf("state = %v, want StateAwaitingTools", m.state)
		}
		// Tool model is created even for unknown tools (uses generic)
		if len(m.pendingToolIDs) != 1 {
			t.Fatalf("expected 1 pending tool, got %d", len(m.pendingToolIDs))
		}
		if !m.pendingToolIDs["tool-1"] {
			t.Error("expected tool-1 to be pending")
		}
	})

	t.Run("handles error", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		m.state = StateStreaming

		m, _ = m.handleStreamDone(turn.StreamDoneMsg{
			Err: context.Canceled,
		})

		if m.state != StateIdle {
			t.Errorf("state = %v, want StateIdle", m.state)
		}
	})
}

func TestModel_handleToolResult(t *testing.T) {
	t.Parallel()

	t.Run("removes tool from pending", func(t *testing.T) {
		t.Parallel()

		m := newTestModel(t)
		m.state = StateAwaitingTools
		m.pendingToolIDs = map[string]bool{
			"tool-1": true,
			"tool-2": true,
		}

		// Need to set currentAssistant for handleToolResult to process
		msg := domain.Message{
			ID:   "msg-1",
			Role: domain.RoleAssistant,
			Content: []domain.Block{
				{Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-1", Name: "unknown"}},
				{Type: domain.BlockTypeToolUse, ToolUse: &domain.ToolUse{ID: "tool-2", Name: "unknown"}},
			},
		}
		m.currentAssistant = assistant.New(m.theme, logtest.New(t), msg, apptools.Tools{})

		m, _ = m.handleToolResult(tools.ResultMsg{
			ToolUseID: "tool-1",
			Result:    &domain.ToolResult{ToolUseID: "tool-1"},
		})

		if len(m.pendingToolIDs) != 1 {
			t.Fatalf("expected 1 pending tool, got %d", len(m.pendingToolIDs))
		}
		if m.pendingToolIDs["tool-1"] {
			t.Error("tool-1 should not be pending")
		}
		if !m.pendingToolIDs["tool-2"] {
			t.Error("tool-2 should still be pending")
		}
	})
}

func newTestModel(t *testing.T) Model {
	t.Helper()

	db := sqlitetest.OpenBareDB(t)
	theme := styles.NewTheme(true)

	return New(
		context.Background(),
		theme,
		db,
		&chattest.MockClient{},
		"acc-1",
		"ws-1",
		apptools.Tools{},
		logtest.New(t),
	)
}
