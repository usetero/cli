// Package show renders entity cards inside tool chrome.
// The parent tools.Model handles icon, name, collapse/expand, and status.
// Entity rendering is dispatched by type: policies → policycard, etc.
package show

import (
	"encoding/json"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/domain"
	domaintools "github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/policycard"
)

// Model handles show tool execution and renders entity content.
// Chrome (icon, name, expand/collapse) is handled by the parent tools.Model.
type Model struct {
	theme    styles.Theme
	scope    log.Scope
	index    int
	toolID   string
	state    tools.State
	executor *chattools.ShowTool
	width    int

	// Input accumulation
	input string

	// Parsed input
	entity domaintools.EntityType

	// Results
	result domaintools.ShowResult
	err    error
	policy *domain.Policy

	// Entity-specific child components (created on execute based on entity type)
	card *policycard.Model
}

// New creates a new show tool model.
func New(theme styles.Theme, index int, toolID string, width int, executor *chattools.ShowTool, scope log.Scope) *Model {
	return &Model{
		theme:    theme,
		scope:    scope.Child("show"),
		index:    index,
		toolID:   toolID,
		state:    tools.StateAccumulating,
		executor: executor,
		width:    width,
	}
}

// AutoExpand implements tools.AutoExpander — show results are always expanded.
func (m *Model) AutoExpand() bool { return true }

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

	var in domaintools.ShowInput
	if err := json.Unmarshal([]byte(m.input), &in); err == nil {
		m.entity = in.Entity
	}

	m.scope.Info("executing show", "entity", string(m.entity))

	if m.executor == nil {
		m.err = fmt.Errorf("no executor")
		m.state = tools.StateComplete
		return m.fireCompleted()
	}

	result, err := m.executor.Execute(json.RawMessage(m.input))
	if err != nil {
		m.err = err
		m.state = tools.StateComplete
		m.scope.Error("show failed", "error", err)
		return m.fireCompleted()
	}

	m.result = result
	m.state = tools.StateComplete

	// Create entity-specific child components based on result type.
	if result.Entity == domaintools.EntityPolicy && result.Data != nil {
		if p, ok := result.Data["policy"].(*domain.Policy); ok {
			m.policy = p
			m.card = policycard.New(m.theme)
			m.card.SetPolicy(p)
			m.card.SetWidth(m.width)
		}
	}

	m.scope.Info("show completed", "entity", string(result.Entity), "id", result.ID)
	return m.fireCompleted()
}

func (m *Model) fireCompleted() tea.Cmd {
	return func() tea.Msg {
		return msgs.ToolCompleted{
			ToolUseID: m.toolID,
			Result:    domaintools.Result{Content: m.result.ToMap()},
			Error:     m.err,
		}
	}
}

// ─── tools.Child interface ──────────────────────────────────────────────────

// Name returns the display name for the chrome header.
func (m *Model) Name() string {
	if m.policy != nil {
		categoryName := m.policy.CategoryDisplayName
		if categoryName == "" {
			categoryName = format.TitleCase(string(m.policy.Category))
		}
		if categoryName != "" {
			return "Policy · pol-" + m.result.IDShort + " · " + categoryName
		}
	}
	return "Policy"
}

// Status returns the message shown while executing.
func (m *Model) Status() string {
	return "Fetching policy"
}

// Result returns the message shown when complete.
func (m *Model) Result() string {
	if m.err != nil {
		return "Failed"
	}
	if m.policy != nil {
		if m.policy.ServiceName != "" && m.policy.LogEventName != "" {
			return m.policy.ServiceName + " / " + m.policy.LogEventName
		}
		if m.policy.ServiceName != "" {
			return m.policy.ServiceName
		}
	}
	return ""
}

// View renders the policy body content (shown when expanded).
func (m *Model) View() string {
	if m.policy == nil {
		return ""
	}
	switch m.result.Entity {
	case domaintools.EntityPolicy:
		return m.card.View()
	default:
		return ""
	}
}

// SetWidth sets the available width for content rendering.
func (m *Model) SetWidth(width int) {
	m.width = width
	if m.card != nil {
		m.card.SetWidth(width)
	}
}

// State returns the current execution state.
func (m *Model) State() tools.State { return m.state }

// ToolID returns the tool use ID.
func (m *Model) ToolID() string { return m.toolID }

// Err returns any execution error.
func (m *Model) Err() error { return m.err }
