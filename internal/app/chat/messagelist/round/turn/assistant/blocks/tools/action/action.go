// Package action provides a generic tool UI model for simple tools
// that accumulate input, execute, and show status — with no custom body.
package action

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/log"
)

// Executor runs a tool and returns a result.
type Executor func(input json.RawMessage) (domaintools.Result, error)

// Config provides display strings for the chrome wrapper.
type Config struct {
	DisplayName func(input json.RawMessage) string
	Status      func(input json.RawMessage) string
	Result      func(result domaintools.Result) string
}

// Model is a generic tool UI model implementing tools.Child.
type Model struct {
	scope    log.Scope
	index    int
	turnID   string
	toolID   string
	state    tools.State
	config   Config
	executor Executor
	width    int

	input  json.RawMessage
	result domaintools.Result
	err    error
}

type actionExecutedMsg struct {
	result domaintools.Result
	err    error
}

// New creates a new generic action tool model.
func New(index int, turnID, toolID string, width int, config Config, executor Executor, scope log.Scope) *Model {
	return &Model{
		scope:    scope,
		index:    index,
		turnID:   turnID,
		toolID:   toolID,
		state:    tools.StateAccumulating,
		config:   config,
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
	case actionExecutedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = tools.StateComplete
			m.scope.Error("action failed", "name", m.config.DisplayName(m.input), "error", msg.err)
			return m.fireCompleted()
		}

		m.result = msg.result
		m.state = tools.StateComplete
		m.scope.Info("action completed", "name", m.config.DisplayName(m.input))
		return m.fireCompleted()
	}
	return nil
}

func (m *Model) handleContent(content []domain.Block) tea.Cmd {
	if m.state != tools.StateAccumulating {
		return nil
	}

	for _, b := range content {
		if b.Index == m.index && b.Type == domain.BlockTypeToolUse && b.ToolUse != nil {
			m.input = b.ToolUse.Input
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
	m.scope.Info("executing action", "name", m.config.DisplayName(m.input), "input", string(m.input))
	input := append(json.RawMessage(nil), m.input...)
	executor := m.executor
	return func() tea.Msg {
		result, err := executor(input)
		return actionExecutedMsg{result: result, err: err}
	}
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.ToolCompleted{
			TurnID:    m.turnID,
			ToolUseID: m.toolID,
			Result:    m.result,
			Error:     m.err,
		}
	}
}

// Name returns the display name.
func (m *Model) Name() string {
	return m.config.DisplayName(m.input)
}

// Status returns the status message shown while executing.
func (m *Model) Status() string {
	return m.config.Status(m.input)
}

// Result returns the result message shown when complete.
func (m *Model) Result() string {
	return m.config.Result(m.result)
}

// View returns empty — simple tools have no body content.
func (m *Model) View() string {
	return ""
}

// SetWidth sets the width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// State returns the current state.
func (m *Model) State() tools.State {
	return m.state
}

// ToolID returns the tool use ID.
func (m *Model) ToolID() string {
	return m.toolID
}

// Err returns any error from execution.
func (m *Model) Err() error {
	return m.err
}
