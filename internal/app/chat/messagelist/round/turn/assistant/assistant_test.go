package assistant

import (
	"encoding/json"
	"strings"
	"testing"

	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

// actionToolBlock builds a generic action tool block for rendering tests.
func actionToolBlock(t *testing.T, width int, displayName string) *tools.Model {
	t.Helper()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)
	cfg := action.Config{
		DisplayName: func(json.RawMessage) string { return displayName },
		Status:      func(json.RawMessage) string { return "Running" },
		Result:      func(domaintools.Result) string { return "Done" },
	}
	exec := func(json.RawMessage) (domaintools.Result, error) { return domaintools.Result{}, nil }
	child := action.New(0, "turn-1", "tool-1", width, cfg, exec, scope)
	return tools.New(theme, 0, "turn-1", "tool-1", width, child)
}

func TestCancel(t *testing.T) {
	t.Parallel()

	t.Run("cancels tool blocks", func(t *testing.T) {
		t.Parallel()
		theme := styles.NewTheme(true)
		scope := logtest.NewScope(t)

		m := New(theme, "turn-1", "test-msg", 80, nil, scope)

		// Add a tool block in pending state.
		tool := actionToolBlock(t, 78, "List Services")
		m.AddBlock(tool)

		m.Cancel()

		// Tool should render as cancelled — padding + icon/name + padding (3 lines).
		view := tool.View()
		if lines := strings.Count(view, "\n"); lines != 2 {
			t.Errorf("expected 3-line render for cancelled tool, got %d lines", lines+1)
		}
		if !strings.Contains(view, "List Services") {
			t.Error("expected tool name in cancelled view")
		}
	})

	t.Run("no-op with no tool blocks", func(t *testing.T) {
		t.Parallel()
		theme := styles.NewTheme(true)
		scope := logtest.NewScope(t)

		m := New(theme, "turn-1", "test-msg", 80, nil, scope)
		m.Cancel() // should not panic
	})
}

func TestNilRegistryToolUseDoesNotPanic(t *testing.T) {
	t.Parallel()

	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)
	m := New(theme, "turn-1", "test-msg", 80, nil, scope)

	cmd := m.Update(msgs.AssistantContentUpdated{
		TurnID: "turn-1",
		Message: domain.Message{
			Content: []domain.Block{
				{
					Index: 0,
					Type:  domain.BlockTypeToolUse,
					ToolUse: &domain.ToolUse{
						ID:            "tool-1",
						Name:          "unknown_tool",
						Input:         []byte(`{}`),
						InputComplete: true,
					},
				},
			},
		},
	})

	if cmd == nil {
		t.Fatal("expected non-nil cmd to initialize tool block")
	}
	if len(m.Blocks()) != 1 {
		t.Fatalf("expected 1 block, got %d", len(m.Blocks()))
	}
}
