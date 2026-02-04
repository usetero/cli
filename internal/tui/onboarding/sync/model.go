package sync

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components/progress"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const pollInterval = 200 * time.Millisecond

// pollMsg triggers a status check.
type pollMsg struct{}

// Model waits for sync to complete before transitioning to the app.
// It polls the syncer for current state - no message passing needed.
type Model struct {
	theme  *styles.Theme
	logger log.Logger

	org       domain.Organization
	account   domain.Account
	workspace domain.Workspace

	syncer powersync.Syncer

	// UI state
	spinner spinner.Model
	width   int
	height  int
}

// New creates a new sync step.
func New(
	theme *styles.Theme,
	org domain.Organization,
	account domain.Account,
	workspace domain.Workspace,
	syncer powersync.Syncer,
	logger log.Logger,
) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return Model{
		theme:     theme,
		logger:    logger,
		org:       org,
		account:   account,
		workspace: workspace,
		syncer:    syncer,
		spinner:   sp,
		width:     80,
	}
}

// Init starts polling if not already ready.
func (m Model) Init() tea.Cmd {
	if m.syncer.IsReady() {
		return nil
	}
	return tea.Batch(m.spinner.Tick, m.poll())
}

// poll returns a command that triggers a status check after the poll interval.
func (m Model) poll() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollMsg{}
	})
}

// readyMsg signals that sync is complete and we should transition.
type readyMsg struct{}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg.(type) {
	case pollMsg:
		if m.syncer.IsReady() {
			m.logger.Info("sync ready")
			// Send a message to trigger the flow to check Next()
			return m, func() tea.Msg { return readyMsg{} }
		}
		// Keep polling
		return m, m.poll()

	case readyMsg:
		// Flow will call Next() and transition
		return m, nil
	}

	// Update spinner while waiting
	if !m.syncer.IsReady() {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the sync UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	title := themeStyles.Title.Render("Getting ready")

	switch s := m.syncer.State().(type) {
	case *powersync.Ready:
		statusMsg := themeStyles.Body.Render("Ready!")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusMsg)

	case *powersync.Error:
		errorMsg := themeStyles.Error.Render(fmt.Sprintf("Error: %v", s.Err))
		return lipgloss.JoinVertical(lipgloss.Left, title, "", errorMsg)

	case *powersync.Connecting:
		statusLine := m.spinner.View() + " " + themeStyles.Body.Render(s.Message)
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	case *powersync.Syncing:
		statusLine := m.spinner.View() + " " + themeStyles.Body.Render(s.Message)
		parts := []string{title, "", statusLine}

		// Show progress bar if we have progress data
		if s.Progress != nil && s.Progress.Total > 0 {
			pct := float64(s.Progress.Downloaded) / float64(s.Progress.Total) * 100
			prog := progress.New(m.theme, 50)
			progressBar := prog.ViewAs(pct)

			countText := fmt.Sprintf("%d / %d rows", s.Progress.Downloaded, s.Progress.Total)
			parts = append(parts, "", progressBar, "", themeStyles.Help.Render(countText))
		}

		// Show warning as secondary line if present
		if s.Warning != "" {
			warningStyle := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
			warningLine := "  " + warningStyle.Render(s.Warning)
			parts = append(parts, "", warningLine)
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

	default: // Disconnected or unknown
		statusLine := m.spinner.View() + " " + themeStyles.Body.Render("Starting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
	}
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns true while waiting for sync.
func (m Model) IsBusy() bool {
	return !m.syncer.IsReady()
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	_, ok := m.syncer.State().(*powersync.Error)
	return ok
}

// Error returns the sync error.
func (m Model) Error() error {
	if e, ok := m.syncer.State().(*powersync.Error); ok {
		return e.Err
	}
	return nil
}

// Help returns empty key bindings.
func (m Model) Help() help.KeyMap {
	return keymap.Simple{Keys: []key.Binding{}}
}

// Next returns nil when sync is ready (flow complete).
func (m Model) Next() (step.Step, error) {
	if !m.syncer.IsReady() {
		return nil, step.ErrNotReady
	}
	// Flow complete - transition to app
	return nil, nil
}

// Organization returns the organization.
func (m Model) Organization() domain.Organization {
	return m.org
}

// Account returns the account.
func (m Model) Account() domain.Account {
	return m.account
}

// Workspace returns the selected workspace.
func (m Model) Workspace() domain.Workspace {
	return m.workspace
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
