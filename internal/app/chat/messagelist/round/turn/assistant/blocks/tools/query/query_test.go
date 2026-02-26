package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/teatest"
)

func TestViewClipsToWidth(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	rows := wideRows()

	m := &Model{
		theme: theme,
		scope: scope,
		rows:  rows,
		width: 80,
	}

	output := m.View()
	teatest.AssertMaxWidth(t, 80, output)
}

func TestViewClipsToWidth_NarrowTerminal(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	rows := []map[string]any{
		{"name": "accounting", "status": "READY", "count": 42},
		{"name": "checkout", "status": "READY", "count": 11},
	}

	for _, width := range []int{40, 60, 80, 120} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			m := &Model{
				theme: theme,
				scope: scope,
				rows:  rows,
				width: width,
			}

			output := m.View()
			teatest.AssertMaxWidth(t, width, output)
		})
	}
}

func TestViewWidthZero(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	rows := []map[string]any{
		{"name": "test", "value": 123},
	}

	m := &Model{
		theme: theme,
		scope: scope,
		rows:  rows,
		width: 0,
	}

	output := m.View()
	t.Logf("width=0 output:\n%s", output)

	// With width 0, no clipping happens — just make sure it doesn't panic
	if output == "" {
		t.Error("expected non-empty output even with width=0")
	}
}

func TestUpdate_IgnoresExecutionCompletionFromDifferentToolInstance(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)
	scope := logtest.NewScope(t)

	m := New(theme, 0, "turn-1", "tool-1", 80, nil, scope)
	cmd := m.Update(queryExecutedMsg{
		toolID:   "tool-2",
		result:   domaintools.QueryResult{Rows: []map[string]any{{"name": "wrong"}}},
		duration: 100 * time.Millisecond,
	})
	if cmd != nil {
		t.Fatal("expected nil cmd for foreign completion")
	}
	if m.state != tools.StateAccumulating {
		t.Fatalf("state = %d, want StateAccumulating", m.state)
	}
	if len(m.rows) != 0 {
		t.Fatalf("rows = %d, want 0", len(m.rows))
	}
}

func wideRows() []map[string]any {
	return []map[string]any{
		{"name": "accounting", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 6, "log_percent_complete": 96.74829044065488, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 6},
		{"name": "ad", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 5, "log_percent_complete": 95.47126675672867, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 5},
		{"name": "cart", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 11, "log_percent_complete": 100, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 11},
		{"name": "checkout", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 41, "log_percent_complete": 100, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 41},
		{"name": "currency", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 1, "log_percent_complete": 95.20486903597435, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 1},
		{"name": "email", "log_analyzing_count": 0, "log_discovering_count": 0, "log_error": nil, "log_event_count": 2, "log_percent_complete": 100, "log_saved_count": 0, "log_status": "READY", "log_valuable_count": 0, "log_waste_count": 2},
	}
}
