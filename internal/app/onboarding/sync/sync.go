// Package sync provides the sync step for onboarding.
package sync

import (
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/progress"
)

const pollInterval = 200 * time.Millisecond

// pollMsg triggers a sync status check.
type pollMsg struct{}

// Model waits for sync to complete.
type Model struct {
	theme    *styles.Theme
	syncer   powersync.Syncer
	logger   log.Logger
	spinner  spinner.Model
	progress *progress.Model
	width    int
	height   int
}

// New creates a new sync step.
func New(theme *styles.Theme, syncer powersync.Syncer, logger log.Logger) *Model {
	if theme == nil {
		panic("theme is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	logger = logger.With(slog.String("step", "sync"))
	logger.Debug("initialized")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	return &Model{
		theme:    theme,
		syncer:   syncer,
		logger:   logger,
		spinner:  sp,
		progress: progress.New(theme, 50),
	}
}

// Init starts polling if not already ready.
func (m *Model) Init() tea.Cmd {
	if m.syncer.IsReady() {
		m.logger.Info("sync already complete")
		return func() tea.Msg { return msgs.SyncComplete{} }
	}
	m.logger.Debug("starting sync poll")
	return tea.Batch(m.spinner.Tick, m.poll())
}

func (m *Model) poll() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollMsg{}
	})
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case pollMsg:
		if m.syncer.IsReady() {
			m.logger.Info("sync completed")
			return func() tea.Msg { return msgs.SyncComplete{} }
		}
		return m.poll()

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
	colors := m.theme.Colors

	title := s.Title.Render("Getting ready")

	switch state := m.syncer.State().(type) {
	case *powersync.Ready:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Success.Render("Ready!"))

	case *powersync.Error:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", s.Error.Render(fmt.Sprintf("Error: %v", state.Err)))

	case *powersync.Connecting:
		statusLine := m.spinner.View() + " " + s.Body.Render(state.Message)
		return lipgloss.JoinVertical(lipgloss.Left, title, "", statusLine)

	case *powersync.Syncing:
		statusLine := m.spinner.View() + " " + s.Body.Render(state.Message)
		parts := []string{title, "", statusLine}

		if state.Progress != nil && state.Progress.Total > 0 {
			pct := float64(state.Progress.Downloaded) / float64(state.Progress.Total) * 100
			progressBar := m.progress.ViewAs(pct)
			countText := fmt.Sprintf("%d / %d rows", state.Progress.Downloaded, state.Progress.Total)
			parts = append(parts, "", progressBar, "", s.Help.Render(countText))
		}

		if state.Warning != "" {
			warningStyle := lipgloss.NewStyle().Foreground(colors.Warning.Fg)
			parts = append(parts, "", "  "+warningStyle.Render(state.Warning))
		}

		return lipgloss.JoinVertical(lipgloss.Left, parts...)

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
