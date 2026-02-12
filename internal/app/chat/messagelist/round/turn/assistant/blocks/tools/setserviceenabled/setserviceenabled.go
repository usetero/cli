package setserviceenabled

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
)

// Model handles set_service_enabled tool execution and rendering.
type Model struct {
	scope    log.Scope
	index    int
	toolID   string
	state    tools.State
	executor *chattools.SetServiceEnabledTool
	width    int

	// Input accumulation
	input string

	// Parsed input
	serviceID string
	enabled   bool

	// Results
	result domaintools.SetServiceEnabledResult
	err    error
}

// New creates a new set_service_enabled tool model.
func New(index int, toolID string, width int, executor *chattools.SetServiceEnabledTool, scope log.Scope) *Model {
	return &Model{
		scope:    scope.Child("set_service_enabled"),
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

	var in domaintools.SetServiceEnabledInput
	if err := json.Unmarshal([]byte(m.input), &in); err == nil {
		m.serviceID = in.ServiceID
		m.enabled = in.Enabled
	}

	m.scope.Info("executing", "service_id", m.serviceID, "enabled", m.enabled)

	if m.executor == nil {
		m.err = fmt.Errorf("no executor")
		m.state = tools.StateComplete
		return m.fireCompleted()
	}

	result, err := m.executor.Execute(json.RawMessage(m.input))
	if err != nil {
		m.err = err
		m.state = tools.StateComplete
		m.scope.Error("failed", "error", err)
		return m.fireCompleted()
	}

	m.result = result
	m.state = tools.StateComplete
	m.scope.Info("completed", "service_name", result.ServiceName, "enabled", result.Enabled)
	return m.fireCompleted()
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.SetServiceEnabledCompleted{
			ToolUseID: m.toolID,
			Result:    m.result,
			Error:     m.err,
		}
	}
}

// Name returns the display name.
func (m *Model) Name() string {
	if m.enabled {
		return "Enable Service"
	}
	return "Disable Service"
}

// Status returns the status message shown while executing.
func (m *Model) Status() string {
	name := m.result.ServiceName
	if name == "" {
		name = m.serviceID
	}
	if m.enabled {
		return fmt.Sprintf("Enabling %s", name)
	}
	return fmt.Sprintf("Disabling %s", name)
}

// Result returns the result message shown when complete.
func (m *Model) Result() string {
	name := m.result.ServiceName
	if name == "" {
		name = m.serviceID
	}
	if m.result.Enabled {
		return fmt.Sprintf("%s enabled", name)
	}
	return fmt.Sprintf("%s disabled", name)
}

// View returns the body content (empty for this tool).
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
