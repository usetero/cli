package query

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	msgs "github.com/usetero/cli/internal/app/chat/events"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	chattools "github.com/usetero/cli/internal/app/chattools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

// Model handles query tool execution and content rendering.
// Chrome (icon, name) is handled by the parent tools.Model.
type Model struct {
	theme    styles.Theme
	scope    log.Scope
	index    int
	turnID   domain.MessageID
	toolID   string
	state    tools.State
	executor *chattools.QueryTool
	width    int

	// Input accumulation
	input string

	// Parsed input
	sql            string
	status         string
	resultTemplate string

	// Results
	rows        []map[string]any
	rowsDropped int
	err         error
	duration    time.Duration
}

type queryExecutionCompletedMsg struct {
	toolID   string
	result   domaintools.QueryResult
	err      error
	duration time.Duration
}

// New creates a new query tool model.
func New(theme styles.Theme, index int, turnID domain.MessageID, toolID string, width int, executor *chattools.QueryTool, scope log.Scope) *Model {
	scope = scope.Child("query")
	return &Model{
		theme:    theme,
		scope:    scope,
		index:    index,
		turnID:   turnID,
		toolID:   toolID,
		state:    tools.StateAccumulating,
		executor: executor,
		width:    width,
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case msgs.AssistantContentUpdated:
		return m.handleContent(msg.Message.Content)
	case msgs.StreamCompleted:
		return m.handleContent(msg.Message.Content)
	case queryExecutionCompletedMsg:
		if msg.toolID != m.toolID {
			return nil
		}
		m.duration = msg.duration
		if msg.err != nil {
			m.err = msg.err
			m.state = tools.StateComplete
			m.scope.Error("query failed", "error", msg.err)
			return m.fireCompleted()
		}

		m.rows = msg.result.Rows
		m.rowsDropped = msg.result.RowsDropped
		m.state = tools.StateComplete
		m.scope.Info("query completed", "row_count", len(m.rows), "rows_dropped", m.rowsDropped, "duration", m.duration)
		return m.fireCompleted()
	}
	return nil
}

// handleContent finds this tool's data by index and updates state.
func (m *Model) handleContent(content []domain.Block) tea.Cmd {
	if m.state != tools.StateAccumulating {
		return nil
	}

	for _, b := range content {
		if b.Index == m.index && b.Type == domain.BlockTypeToolUse && b.ToolUse != nil {
			m.input = string(b.ToolUse.Input)
			if b.ToolUse.InputComplete {
				return m.execute()
			}
			return nil
		}
	}
	return nil
}

// Status returns the status message shown while executing.
func (m *Model) Status() string {
	return m.status
}

// Result returns the result message with {count} substituted.
func (m *Model) Result() string {
	if m.err != nil {
		return "Query failed"
	}
	var base string
	if m.resultTemplate == "" {
		base = fmt.Sprintf("%d rows", len(m.rows))
	} else {
		base = strings.Replace(m.resultTemplate, "{count}", fmt.Sprintf("%d", len(m.rows)), 1)
	}
	if m.duration > 0 {
		base += fmt.Sprintf(" (%.1fs)", m.duration.Seconds())
	}
	return base
}

const maxPreviewRows = 5

// View renders a table preview of query results, clipped to the available width.
func (m *Model) View() string {
	if len(m.rows) == 0 {
		return ""
	}

	// Collect column headers from first row, sorted for stable order.
	// Promote "name" to the front if it exists.
	first := m.rows[0]
	var headers []string
	hasName := false
	for k := range first {
		if k == "name" {
			hasName = true
		} else {
			headers = append(headers, k)
		}
	}
	sort.Strings(headers)
	if hasName {
		headers = append([]string{"name"}, headers...)
	}

	// Cap rows
	showRows := m.rows
	truncatedRows := 0
	if len(showRows) > maxPreviewRows {
		truncatedRows = len(showRows) - maxPreviewRows
		showRows = showRows[:maxPreviewRows]
	}

	// Build table — no explicit width so columns size to content
	tbl := table.New(m.theme, table.WithFitHeaders(), table.WithBackground(m.theme.Bg))
	tbl.Headers(headers...)

	for _, row := range showRows {
		cells := make([]string, len(headers))
		for i, h := range headers {
			cells[i] = fmt.Sprintf("%v", row[h])
		}
		tbl.Row(cells...)
	}

	result := tbl.View()

	if truncatedRows > 0 {
		muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg).PaddingLeft(1)
		result += "\n\n" + muted.Render(fmt.Sprintf("+%d more rows", truncatedRows))
	}

	// Clip each line to available width so wide tables don't wrap
	if m.width > 0 {
		lines := strings.Split(result, "\n")
		for i, line := range lines {
			lines[i] = ansi.Truncate(line, m.width, "")
		}
		result = strings.Join(lines, "\n")
	}

	return result
}

// SetRows sets the result rows directly (for testing).
func (m *Model) SetRows(rows []map[string]any) {
	m.rows = rows
	m.state = tools.StateComplete
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) execute() tea.Cmd {
	m.state = tools.StateExecuting

	// Parse input
	var in domaintools.QueryInput
	if err := json.Unmarshal([]byte(m.input), &in); err == nil {
		m.sql = in.SQL
		m.status = in.Status
		m.resultTemplate = in.Result
	}

	m.scope.Info("executing query", "sql", m.sql, "status", m.status)

	start := time.Now()

	if m.executor == nil {
		m.err = fmt.Errorf("no executor")
		m.state = tools.StateComplete
		m.scope.Error("query failed", "error", m.err)
		return m.fireCompleted()
	}
	input := append([]byte(nil), []byte(m.input)...)
	executor := m.executor
	return func() tea.Msg {
		result, err := executor.Execute(json.RawMessage(input))
		return queryExecutionCompletedMsg{
			toolID:   m.toolID,
			result:   result,
			err:      err,
			duration: time.Since(start),
		}
	}
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		result := domaintools.QueryResult{Rows: m.rows, RowsDropped: m.rowsDropped}
		return msgs.ToolCompleted{
			TurnID:    m.turnID,
			ToolUseID: m.toolID,
			Result:    domaintools.Result{Content: result.ToMap()},
			Error:     m.err,
		}
	}
}

// Name returns the tool's display name.
func (m *Model) Name() string {
	return "Query"
}

// ToolID returns the tool's ID.
func (m *Model) ToolID() string {
	return m.toolID
}

// State returns the tool's current state.
func (m *Model) State() tools.State {
	return m.state
}

// Err returns any error from execution.
func (m *Model) Err() error {
	return m.err
}
