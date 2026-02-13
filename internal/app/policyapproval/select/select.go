// Package select provides the policy selection step.
package policyselect

import (
	"context"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/policyapproval/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// Model handles policy selection with multiselect.
type Model struct {
	ctx      context.Context
	theme    styles.Theme
	scope    log.Scope
	policies []msgs.Policy
	selected map[string]bool // policy ID -> selected
	cursor   int
	width    int
	height   int
}

// New creates a new policy selection step.
func New(ctx context.Context, theme styles.Theme, policies []msgs.Policy, scope log.Scope) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	return &Model{
		ctx:      ctx,
		theme:    theme,
		scope:    scope.Child("select"),
		policies: policies,
		selected: make(map[string]bool),
	}
}

// Init returns the initial command.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("select step started", "policy_count", len(m.policies))
	return nil
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		// TODO: handle keys
		// space - toggle current
		// a - select all
		// n - select none
		// j/down - cursor down
		// k/up - cursor up
		// enter - confirm selection
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.confirm()
		}
	}
	return nil
}

// confirm emits PoliciesSelected with the selected policies.
func (m *Model) confirm() tea.Cmd {
	// TODO: convert msgs.Policy to domain.PolicyDetail when this step is fully implemented
	m.scope.Info("policies selected (stub)")
	return func() tea.Msg {
		return msgs.PoliciesSelected{}
	}
}

// View renders the step.
func (m *Model) View() string {
	// TODO: render policy list with checkboxes
	return lipgloss.NewStyle().
		Foreground(m.theme.Text).
		Render("Select policies to approve (TODO: render list)")
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ShortHelp returns key bindings.
func (m *Model) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}
