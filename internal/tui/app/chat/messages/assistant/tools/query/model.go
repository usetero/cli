package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/chat/messages/assistant/tools"
	appquery "github.com/usetero/cli/internal/tui/app/tools/query"
	"github.com/usetero/cli/internal/tui/components/table"
)

const maxCollapsedRows = 5

// Model renders and executes a query tool.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	use      *domain.ToolUse
	executor *appquery.Tool
	result   *domain.ToolResult
	expanded bool

	sql  string
	rows []map[string]any
}

// Compile-time interface check
var _ tools.Body = (*Model)(nil)

// New creates a new query model.
// The executor is required - this model executes itself.
func New(theme *styles.Theme, logger log.Logger, use *domain.ToolUse, executor *appquery.Tool) *Model {
	m := &Model{
		theme:    theme,
		logger:   logger,
		use:      use,
		executor: executor,
	}

	// Parse SQL from input
	var in struct {
		SQL string `json:"sql"`
	}
	if err := json.Unmarshal(use.Input, &in); err == nil {
		m.sql = in.SQL
	}

	return m
}

// Init executes the query. Returns a Cmd that produces ResultMsg.
func (m *Model) Init() tea.Cmd {
	use := m.use
	executor := m.executor
	logger := m.logger
	sql := m.sql

	return func() tea.Msg {
		logger.Info("executing query")
		logger.Debug("query sql", "sql", sql)

		result, err := executor.Execute(use.Input)

		if err != nil {
			logger.Error("query failed", "error", err)
			return tools.ResultMsg{
				ToolUseID: use.ID,
				Result: &domain.ToolResult{
					ToolUseID: use.ID,
					IsError:   true,
					Error:     err.Error(),
				},
			}
		}

		logger.Info("query completed")

		content, err := json.Marshal(result)
		if err != nil {
			return tools.ResultMsg{
				ToolUseID: use.ID,
				Result: &domain.ToolResult{
					ToolUseID: use.ID,
					IsError:   true,
					Error:     "failed to marshal result: " + err.Error(),
				},
			}
		}

		return tools.ResultMsg{
			ToolUseID: use.ID,
			Result: &domain.ToolResult{
				ToolUseID: use.ID,
				Content:   content,
			},
		}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tools.ResultMsg:
		if msg.ToolUseID == m.use.ID {
			m.result = msg.Result
			// Parse rows from result
			if msg.Result != nil && len(msg.Result.Content) > 0 {
				if err := json.Unmarshal(msg.Result.Content, &m.rows); err != nil {
					m.logger.Error("failed to parse query result", "error", err)
				} else {
					m.logger.Debug("query result parsed", "rows", len(m.rows))
				}
			}
		}
	}
	return nil
}

// Result returns the tool result.
func (m *Model) Result() *domain.ToolResult {
	return m.result
}

// Render returns the rendered body.
func (m *Model) Render(width int) string {
	colors := m.theme.Colors
	var parts []string

	// SQL query
	if m.sql != "" {
		sqlStyle := lipgloss.NewStyle().
			Foreground(colors.Brand.GradientStart).
			Width(width)
		parts = append(parts, sqlStyle.Render(m.sql))
	}

	// Status-specific content
	if m.result == nil {
		running := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Italic(true).
			Render("Executing query...")
		parts = append(parts, running)
	} else if m.result.IsError {
		parts = append(parts, m.renderError(width))
	} else if len(m.rows) > 0 {
		parts = append(parts, m.renderTable(width))
	} else {
		empty := lipgloss.NewStyle().
			Foreground(colors.Page.TextMuted).
			Italic(true).
			Render("No results")
		parts = append(parts, empty)
	}

	return strings.Join(parts, "\n")
}

func (m *Model) renderTable(width int) string {
	var cols []string
	for col := range m.rows[0] {
		cols = append(cols, col)
	}

	rowsToShow := m.rows
	truncated := false
	if !m.expanded && len(m.rows) > maxCollapsedRows {
		rowsToShow = m.rows[:maxCollapsedRows]
		truncated = true
	}

	tbl := table.New(m.theme).
		Headers(cols...).
		Width(width)

	for _, row := range rowsToShow {
		var cells []string
		for _, col := range cols {
			cells = append(cells, fmt.Sprintf("%v", row[col]))
		}
		tbl.Row(cells...)
	}

	result := tbl.Render()

	if truncated {
		hint := lipgloss.NewStyle().
			Foreground(m.theme.Colors.Page.TextMuted).
			Italic(true).
			Render(fmt.Sprintf("... %d more rows [space to expand]", len(m.rows)-maxCollapsedRows))
		result = result + "\n" + hint
	}

	return result
}

func (m *Model) renderError(width int) string {
	colors := m.theme.Colors

	errTag := lipgloss.NewStyle().
		Background(colors.Error.Bg).
		Foreground(colors.Error.Fg).
		Padding(0, 1).
		Render("ERROR")

	errMsg := lipgloss.NewStyle().
		Foreground(colors.Error.Fg).
		Width(width - 10).
		Render(m.result.Error)

	return fmt.Sprintf("%s %s", errTag, errMsg)
}

// Params returns header params.
func (m *Model) Params() []string {
	if len(m.rows) > 0 {
		return []string{fmt.Sprintf("%d rows", len(m.rows))}
	}
	return nil
}

// SetExpanded sets the expanded state.
func (m *Model) SetExpanded(expanded bool) {
	m.expanded = expanded
}

// SQL returns the SQL query for copying.
func (m *Model) SQL() string {
	return m.sql
}
