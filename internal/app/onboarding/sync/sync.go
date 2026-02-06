// Package sync provides the sync step for onboarding.
package sync

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/progress"
)

// Model waits for sync to complete.
type Model struct {
	theme    *styles.Theme
	syncer   powersync.Syncer
	scope    log.Scope
	spinner  spinner.Model
	progress *progress.Model
	width    int
	height   int
}

// New creates a new sync step.
func New(theme *styles.Theme, syncer powersync.Syncer, scope log.Scope) *Model {
	if theme == nil {
		panic("theme is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}

	scope.Debug("initialized")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return &Model{
		theme:    theme,
		syncer:   syncer,
		scope:    scope,
		spinner:  sp,
		progress: progress.New(theme, 50),
	}
}

// Init starts the spinner and checks if already ready.
func (m *Model) Init() tea.Cmd {
	if m.syncer.IsReady() {
		m.scope.Info("sync already complete")
		return func() tea.Msg { return msgs.SyncComplete{} }
	}
	m.scope.Debug("waiting for sync")
	return m.spinner.Tick
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case appmsg.SyncStateChanged:
		if _, ok := msg.State.(*powersync.Ready); ok {
			m.scope.Info("sync completed")
			return func() tea.Msg { return msgs.SyncComplete{} }
		}
		return nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}

	return m.progress.Update(msg)
}

// View renders the sync UI.
func (m *Model) View() string {
	s := m.theme.Styles
	title := s.Title.Render("Getting ready")

	switch state := m.syncer.State().(type) {
	case *powersync.Ready:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Success.Render("Ready!"))

	case *powersync.Error:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Error.Render(fmt.Sprintf("Error: %v", state.Err)))

	case *powersync.Connecting:
		statusLine := m.spinner.View() + " " + s.Body.Render("Connecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	case *powersync.Syncing:
		msg := "Syncing your data..."
		if state.Progress != nil && state.Progress.Total > 0 {
			msg = fmt.Sprintf("Syncing your data... (%s)", state.Progress)
		}
		statusLine := m.spinner.View() + " " + s.Body.Render(msg)
		parts := []string{title, "", statusLine}

		if state.Progress != nil && state.Progress.Total > 0 {
			pct := float64(state.Progress.Downloaded) / float64(state.Progress.Total) * 100
			progressBar := m.progress.ViewAs(pct)
			countText := fmt.Sprintf("%d / %d rows", state.Progress.Downloaded, state.Progress.Total)
			parts = append(parts, "", progressBar, "", s.Help.Render(countText))
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

	case *powersync.Reconnecting:
		statusLine := m.spinner.View() + " " + s.Body.Render("Reconnecting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	default:
		statusLine := m.spinner.View() + " " + s.Body.Render("Starting...")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)
	}
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progress.SetWidth(min(width, 50))
}

// ShortHelp returns the key bindings for the short help view.
func (m *Model) ShortHelp() []key.Binding {
	// Sync is automatic, no user action needed
	return nil
}
