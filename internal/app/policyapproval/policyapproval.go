// Package policyapproval provides a wizard for bulk policy approval.
package policyapproval

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/policyapproval/categorydetail"
	"github.com/usetero/cli/internal/app/policyapproval/categorysummary"
	"github.com/usetero/cli/internal/app/policyapproval/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

const (
	sideBySideMinWidth = 125 // ~61 per table + 3 gap
	sideBySideGap      = 3
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

	// Steps
	summary *categorysummary.Model // persistent, created once
	detail  Step                   // nil when not drilled in
	step    Step                   // active step for input (summary or detail)

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

// Init starts the wizard with the category summary step.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("policy approval wizard started")
	m.summary = categorysummary.New(m.ctx, m.theme, m.db, m.scope)
	m.step = m.summary
	m.propagateSize()
	return m.summary.Init()
}

// Update handles messages and orchestrates step transitions.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.propagateSize()
		return nil

	case msgs.CategorySelected:
		m.detail = categorydetail.New(m.ctx, m.theme, m.db, msg.Category, m.scope)
		m.step = m.detail
		m.propagateSize()
		return m.detail.Init()

	case msgs.BackToSummary:
		m.detail = nil
		m.step = m.summary
		m.propagateSize()
		return nil // summary already loaded, no re-init

		// TODO: Handle remaining step completion messages
		// case msgs.PoliciesSelected:
		// case msgs.Confirmed:
	}

	// Delegate to current step
	if m.step != nil {
		return m.step.Update(msg)
	}
	return nil
}

// isWide returns true if the viewport is wide enough for side-by-side layout.
func (m *Model) isWide() bool {
	return m.width >= sideBySideMinWidth
}

// propagateSize distributes width/height to summary and detail based on layout.
func (m *Model) propagateSize() {
	if m.summary == nil {
		return
	}
	if m.detail != nil && m.isWide() {
		half := (m.width - sideBySideGap) / 2
		m.summary.SetSize(half, m.height)
		m.detail.SetSize(m.width-half-sideBySideGap, m.height)
	} else if m.detail != nil {
		m.detail.SetSize(m.width, m.height)
	} else {
		m.summary.SetSize(m.width, m.height)
	}
}

// View renders the current step.
func (m *Model) View() string {
	if m.summary == nil {
		return ""
	}

	var content string
	if m.detail != nil && m.isWide() {
		gap := strings.Repeat(" ", sideBySideGap)
		content = lipgloss.JoinHorizontal(lipgloss.Bottom,
			m.summary.View(), gap, m.detail.View())
	} else if m.detail != nil {
		content = m.detail.View()
	} else {
		content = m.summary.View()
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		AlignVertical(lipgloss.Bottom).
		Render(content)
}

// SetSize updates the model's dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.propagateSize()
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	if m.step != nil {
		return m.step.ShortHelp()
	}
	return nil
}
