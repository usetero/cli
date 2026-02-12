// Package policyapproval provides a wizard for bulk policy approval.
package policyapproval

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	policyselect "github.com/usetero/cli/internal/app/policyapproval/select"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

// Model is the policy approval wizard orchestrator.
type Model struct {
	// Dependencies
	ctx   context.Context
	theme styles.Theme
	db    sqlite.DB
	scope log.Scope

	// Input from tool
	toolUseID string

	// Accumulated state from step completions
	selectedIDs   []string
	approvedCount int
	failedCount   int

	// Current step
	step   Step
	width  int
	height int
}

// New creates a new policy approval wizard.
func New(
	ctx context.Context,
	theme styles.Theme,
	db sqlite.DB,
	toolUseID string,
	scope log.Scope,
) *Model {
	return &Model{
		ctx:       ctx,
		theme:     theme,
		db:        db,
		toolUseID: toolUseID,
		scope:     scope.Child("policyapproval"),
	}
}

// Init starts the wizard with the select step.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("policy approval wizard started")
	// TODO: return m.setStep(select.New(...))
	return m.setStep(policyselect.New(m.ctx, m.theme, nil, m.scope))
}

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step != nil {
			m.step.SetSize(m.width, m.height)
		}
		return nil

		// TODO: Handle step completion messages
		// case msgs.PoliciesSelected:
		// case msgs.ApprovalConfirmed:
		// case msgs.ExecutionComplete:
	}

	// Delegate to current step
	if m.step != nil {
		return m.step.Update(msg)
	}
	return nil
}

// setStep sets the current step and initializes it.
func (m *Model) setStep(step Step) tea.Cmd {
	m.step = step
	m.step.SetSize(m.width, m.height)
	return m.step.Init()
}

// View renders the current step.
func (m *Model) View() string {
	if m.step == nil {
		return ""
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(m.step.View())
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.step != nil {
		m.step.SetSize(width, height)
	}
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.step != nil {
		return m.step.ShortHelp()
	}
	return nil
}
