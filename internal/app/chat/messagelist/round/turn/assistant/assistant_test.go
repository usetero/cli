package assistant

import (
	"fmt"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/query"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/teatest"
)

func TestBlocksNoWrapping(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	// Wide 10-column query result — the exact data that was wrapping
	rows := []map[string]any{
		{"name": "accounting", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 6, "log_percent_complete": 96.74829044065488, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 6},
		{"name": "ad", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 5, "log_percent_complete": 95.47126675672867, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 5},
		{"name": "cart", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 11, "log_percent_complete": 100, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 11},
		{"name": "checkout", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 41, "log_percent_complete": 100, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 41},
		{"name": "currency", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 1, "log_percent_complete": 95.20486903597435, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 1},
	}

	for _, termWidth := range []int{80, 120, 160, 200} {
		t.Run(fmt.Sprintf("term_%d", termWidth), func(t *testing.T) {
			// Real width chain: app subtracts 2 for horizontal padding
			assistantWidth := termWidth - 2
			contentWidth := assistantWidth - block.BorderWidth

			// Real assistant model
			m := New(theme, "turn-1", "test-msg", assistantWidth, nil, scope)
			// Real query model — pass contentWidth (same as production in newToolBlock)
			q := query.New(theme, 0, "turn-1", "tool-1", contentWidth, nil, scope)
			q.SetRows(rows)

			// Real tool model wrapping query (same as production)
			tool := tools.New(theme, 0, "turn-1", "tool-1", contentWidth, q)
			tool.ForceStatus(tools.StatusSuccess)

			// Add tool block to assistant
			m.AddBlock(tool)

			// Verify each block renders within contentWidth.
			// The viewport applies a border; blocks handle their own internal padding.
			// Blocks must fit within contentWidth.
			for _, b := range m.Blocks() {
				b.SetWidth(contentWidth)
				output := b.View()
				teatest.AssertMaxWidth(t, contentWidth, output)
			}
		})
	}
}

func TestCancel(t *testing.T) {
	t.Parallel()

	t.Run("cancels tool blocks", func(t *testing.T) {
		t.Parallel()
		theme := styles.NewTheme(true)
		scope := logtest.NewScope(t)

		m := New(theme, "turn-1", "test-msg", 80, nil, scope)

		// Add a tool block in pending state
		q := query.New(theme, 0, "turn-1", "tool-1", 78, nil, scope)
		tool := tools.New(theme, 0, "turn-1", "tool-1", 78, q)
		m.AddBlock(tool)

		m.Cancel()

		// Tool should render as cancelled — padding + icon/name + padding (3 lines)
		view := tool.View()
		if lines := strings.Count(view, "\n"); lines != 2 {
			t.Errorf("expected 3-line render for cancelled tool, got %d lines", lines+1)
		}
		if !strings.Contains(view, "Query") {
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
