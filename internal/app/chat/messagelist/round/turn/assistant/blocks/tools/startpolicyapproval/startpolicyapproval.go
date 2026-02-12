package startpolicyapproval

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	policyapprovalmsg "github.com/usetero/cli/internal/app/policyapproval/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model handles start_policy_approval tool execution.
// Unlike other tools, this emits a message to open the wizard UI.
type Model struct {
	theme    styles.Theme
	scope    log.Scope
	index    int
	toolID   string
	state    tools.State
	executor *chattools.StartPolicyApprovalTool
	width    int

	// Result
	started bool
	err     error
}

// New creates a new start policy approval tool model.
func New(theme styles.Theme, index int, toolID string, width int, executor *chattools.StartPolicyApprovalTool, scope log.Scope) *Model {
	return &Model{
		theme:    theme,
		scope:    scope.Child("startpolicyapproval"),
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

// handleContent finds this tool's data and triggers execution.
func (m *Model) handleContent(content []domain.Block) tea.Cmd {
	if m.state != tools.StateAccumulating {
		return nil
	}

	for _, b := range content {
		if b.Index == m.index && b.Type == domain.BlockTypeToolUse && b.ToolUse != nil {
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
	m.scope.Info("starting policy approval wizard")

	// Emit message to open the wizard UI
	// The wizard will query the DB for pending policies
	m.started = true
	m.state = tools.StateComplete

	return tea.Batch(
		// Open the wizard
		func() tea.Msg {
			return policyapprovalmsg.Start{
				ToolUseID: m.toolID,
				Policies:  nil, // Wizard will query DB
			}
		},
		// Signal tool completion for chat flow
		m.fireCompleted(),
	)
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.StartPolicyApprovalCompleted{
			ToolUseID: m.toolID,
			Started:   m.started,
			Error:     m.err,
		}
	}
}

// Name returns the tool's display name.
func (m *Model) Name() string {
	return "Policy Approval"
}

// Status returns the status message shown while executing.
func (m *Model) Status() string {
	return "Opening policy approval wizard"
}

// Result returns the result message.
func (m *Model) Result() string {
	if m.started {
		return "Wizard opened"
	}
	return "Failed to open wizard"
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
