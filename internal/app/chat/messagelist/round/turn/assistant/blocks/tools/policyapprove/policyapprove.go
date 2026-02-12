package policyapprove

import (
	"encoding/json"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model handles policy approve tool execution.
type Model struct {
	theme    styles.Theme
	scope    log.Scope
	index    int
	toolID   string
	state    tools.State
	executor *chattools.PolicyApproveTool
	width    int

	// Input
	input    string
	policyID string

	// Result
	approved bool
	err      error
}

// New creates a new policy approve tool model.
func New(theme styles.Theme, index int, toolID string, width int, executor *chattools.PolicyApproveTool, scope log.Scope) *Model {
	return &Model{
		theme:    theme,
		scope:    scope.Child("policyapprove"),
		index:    index,
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
	}
	return nil
}

// handleContent finds this tool's data and updates state.
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

func (m *Model) execute() tea.Cmd {
	m.state = tools.StateExecuting

	// Parse input
	var in domaintools.PolicyApproveInput
	if err := json.Unmarshal([]byte(m.input), &in); err == nil {
		m.policyID = in.PolicyID
	}

	m.scope.Info("approving policy", "policy_id", m.policyID)

	if m.executor == nil {
		m.err = fmt.Errorf("no executor")
		m.state = tools.StateComplete
		m.scope.Error("policy approve failed", "error", m.err)
		return m.fireCompleted()
	}

	result, err := m.executor.Execute(json.RawMessage(m.input))
	if err != nil {
		m.err = err
		m.state = tools.StateComplete
		m.scope.Error("policy approve failed", "error", err)
		return m.fireCompleted()
	}

	m.approved = result.Approved
	m.state = tools.StateComplete
	m.scope.Info("policy approved", "policy_id", m.policyID)
	return m.fireCompleted()
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.PolicyApproveCompleted{
			ToolUseID: m.toolID,
			PolicyID:  m.policyID,
			Approved:  m.approved,
			Error:     m.err,
		}
	}
}

// Name returns the tool's display name.
func (m *Model) Name() string {
	return "Approve Policy"
}

// Status returns the status message shown while executing.
func (m *Model) Status() string {
	if m.policyID != "" {
		return "Approving policy " + m.policyID
	}
	return "Approving policy"
}

// Result returns the result message.
func (m *Model) Result() string {
	if m.approved {
		return "Policy approved"
	}
	return "Policy approval failed"
}

// View renders content (empty for this tool).
func (m *Model) View() string {
	return ""
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
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
