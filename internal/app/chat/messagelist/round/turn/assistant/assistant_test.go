package assistant

import (
	"fmt"
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

func countBlockKinds(blocks []block.Block, kind block.Kind) int {
	n := 0
	for _, b := range blocks {
		if b.Kind() == kind {
			n++
		}
	}
	return n
}

func TestBlocks_ThinkingAnimation(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	t.Run("included while streaming", func(t *testing.T) {
		t.Parallel()
		m := New(theme, "test-msg", 80, nil, scope)
		// streaming is true by default from New()

		blocks := m.Blocks()
		if n := countBlockKinds(blocks, block.KindThinkingAnimation); n != 1 {
			t.Errorf("expected 1 thinking animation block while streaming, got %d", n)
		}
	})

	t.Run("removed after StreamCompleted", func(t *testing.T) {
		t.Parallel()
		m := New(theme, "test-msg", 80, nil, scope)

		m.Update(msgs.StreamCompleted{
			TurnID:  "test-msg",
			Message: domain.Message{Content: []domain.Block{{Index: 0, Type: domain.BlockTypeText, Text: &domain.TextBlock{Content: "hello"}}}},
		})

		blocks := m.Blocks()
		if n := countBlockKinds(blocks, block.KindThinkingAnimation); n != 0 {
			t.Errorf("expected 0 thinking animation blocks after StreamCompleted, got %d", n)
		}
	})
}

func TestBlocksNoWrapping(t *testing.T) {
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
			contentWidth := assistantWidth - block.AssistantPadding

			// Real assistant model
			m := New(theme, "test-msg", assistantWidth, nil, scope)
			m.streaming = false

			// Real query model — pass contentWidth (same as production in newToolBlock)
			q := query.New(theme, 0, "tool-1", contentWidth, nil, scope)
			q.SetRows(rows)

			// Real tool model wrapping query (same as production)
			tool := tools.New(theme, 0, "tool-1", contentWidth, q)
			tool.ForceStatus(tools.StatusSuccess)

			// Add tool block to assistant
			m.AddBlock(tool)

			// Verify each block renders within contentWidth.
			// The viewport wraps blocks with PaddingLeft(block.AssistantPadding), so
			// blocks themselves must fit within contentWidth.
			for _, b := range m.Blocks() {
				b.SetWidth(contentWidth)
				output := b.View()
				teatest.AssertMaxWidth(t, contentWidth, output)
			}
		})
	}
}
