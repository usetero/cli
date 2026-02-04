package onboarding

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/layouts/header"
	"github.com/usetero/cli/internal/tui/onboarding/auth/check"
	"github.com/usetero/cli/internal/tui/onboarding/step"
	"github.com/usetero/cli/internal/tui/onboarding/sync"
	workspaceselect "github.com/usetero/cli/internal/tui/onboarding/workspace/select"
)

// Model orchestrates the onboarding flow.
type Model struct {
	ctx    context.Context
	theme  *styles.Theme
	logger log.Logger
	layout header.Model
	flow   step.Flow

	// Syncer - passed in, used to create sync step
	syncer powersync.Syncer

	// Track if we've injected the sync step
	syncStepDone bool

	// Final state for parent to read
	org       domain.Organization
	account   domain.Account
	workspace domain.Workspace

	ready  bool
	width  int
	height int
}

// New creates a new onboarding model.
func New(ctx context.Context, theme *styles.Theme, authService auth.Auth, prefs preferences.Preferences, apiEndpoint string, syncer powersync.Syncer, logger log.Logger) Model {
	// Start with auth check
	startStep := check.New(ctx, theme, authService, prefs, apiEndpoint, logger)
	flow := step.NewFlow(startStep, logger)

	return Model{
		ctx:    ctx,
		theme:  theme,
		logger: logger,
		layout: header.New(theme, logger),
		flow:   flow,
		syncer: syncer,
	}
}

// Init initializes the onboarding flow.
func (m Model) Init() tea.Cmd {
	return m.flow.Init()
}

// Update handles messages.
// Rule: return early ONLY if this model is the sole consumer of the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle messages we care about
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = m.layout.SetSize(msg.Width, msg.Height)
		contentWidth, contentHeight := m.layout.ContentSize()
		m.flow = m.flow.SetSize(contentWidth, contentHeight)
		m.ready = true // children also need this
	}

	// Forward to children

	m.layout, cmd = m.layout.Update(msg)
	cmds = append(cmds, cmd)

	m.flow, cmd = m.flow.Update(msg)
	cmds = append(cmds, cmd)

	// If flow "completed" at workspace step, inject the sync step
	if m.flow.IsComplete() && !m.syncStepDone {
		lastStep := m.flow.LastStep()
		if ws, ok := lastStep.(workspaceselect.Model); ok {
			m.org = ws.Organization()
			m.account = ws.Account()
			m.workspace = ws.Workspace()

			syncStep := sync.New(
				m.theme,
				m.org,
				m.account,
				m.workspace,
				m.syncer,
				m.logger,
			)
			m.flow = step.NewFlow(syncStep, m.logger)
			m.flow = m.flow.SetSize(m.width, m.height)
			m.syncStepDone = true
			cmds = append(cmds, m.flow.Init())
		}
	}

	// Update layout with current step's help and error
	m.layout = m.layout.SetKeyBindings(m.flow.Help().ShortHelp())
	m.layout = m.layout.SetError(m.flow.Error())

	return m, tea.Batch(cmds...)
}

// View renders the onboarding.
func (m Model) View() string {
	if !m.ready {
		return ""
	}

	// Get content dimensions from layout
	contentWidth, contentHeight := m.layout.ContentSize()

	// Get step content
	stepContent := m.flow.View()

	// Bottom-align step content in available space
	content := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		AlignVertical(lipgloss.Bottom).
		Render(stepContent)

	// Wrap in layout
	return m.layout.Render(content)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.layout = m.layout.SetSize(width, height)

	// Pass content size to flow
	contentWidth, contentHeight := m.layout.ContentSize()
	m.flow = m.flow.SetSize(contentWidth, contentHeight)

	m.ready = true
	return m
}

// IsComplete returns true when onboarding is done.
func (m Model) IsComplete() bool {
	return m.flow.IsComplete()
}

// IsBusy returns true if performing background work.
func (m Model) IsBusy() bool {
	return m.flow.IsBusy()
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.flow.HasError()
}

// Error returns the current error.
func (m Model) Error() error {
	return m.flow.Error()
}

// Close releases resources.
func (m Model) Close() error {
	return m.flow.Close()
}

// Organization returns the selected organization.
func (m Model) Organization() domain.Organization {
	return m.org
}

// Account returns the selected account.
func (m Model) Account() domain.Account {
	return m.account
}

// Workspace returns the selected workspace.
func (m Model) Workspace() domain.Workspace {
	return m.workspace
}
