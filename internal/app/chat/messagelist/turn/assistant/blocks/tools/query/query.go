package query

import (
	"encoding/json"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model renders and executes a query tool.
type Model struct {
	theme    *styles.Theme
	logger   log.Logger
	index    int
	toolID   string
	name     string
	state    tools.State
	executor *chattools.QueryTool
	width    int

	// Input accumulation
	input string

	// Results
	sql  string
	rows []map[string]any
	err  error
}

// New creates a new query tool model.
func New(theme *styles.Theme, index int, toolID string, width int, executor *chattools.QueryTool, logger log.Logger) *Model {
	return &Model{
		theme:    theme,
		logger:   logger.With("component", "query_tool", "tool_id", toolID),
		index:    index,
		toolID:   toolID,
		name:     executor.Name(),
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

// View renders the tool block.
func (m *Model) View() string {
	switch m.state {
	case tools.StateAccumulating:
		return "⋯ " + m.name
	case tools.StateExecuting:
		return "◐ " + m.name
	case tools.StateComplete:
		if m.err != nil {
			return "✗ " + m.name + ": " + m.err.Error()
		}
		return m.renderResults()
	default:
		return m.name
	}
}

func (m *Model) renderResults() string {
	if len(m.rows) == 0 {
		return "✓ " + m.name + " (no results)"
	}
	return fmt.Sprintf("✓ %s (%d rows)", m.name, len(m.rows))
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) execute() tea.Cmd {
	m.state = tools.StateExecuting

	// Parse SQL from input
	var in struct {
		SQL string `json:"sql"`
	}
	if err := json.Unmarshal([]byte(m.input), &in); err == nil {
		m.sql = in.SQL
	}

	m.logger.Info("executing query", "sql", m.sql)

	if m.executor == nil {
		m.err = fmt.Errorf("no executor")
		m.state = tools.StateComplete
		m.logger.Error("query failed", "error", m.err)
		return m.fireCompleted()
	}

	result, err := m.executor.Execute(json.RawMessage(m.input))
	if err != nil {
		m.err = err
		m.state = tools.StateComplete
		m.logger.Error("query failed", "error", err)
		return m.fireCompleted()
	}

	m.rows = result.Rows
	m.state = tools.StateComplete
	m.logger.Info("query completed", "row_count", len(m.rows))
	return m.fireCompleted()
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.QueryCompleted{
			ToolUseID: m.toolID,
			Result:    domaintools.QueryResult{Rows: m.rows},
			Error:     m.err,
		}
	}
}

// Index returns the block index.
func (m *Model) Index() int {
	return m.index
}

// ToolID returns the tool's ID.
func (m *Model) ToolID() string {
	return m.toolID
}

// Name returns the tool's name.
func (m *Model) Name() string {
	return m.name
}

// State returns the tool's current state.
func (m *Model) State() tools.State {
	return m.state
}
