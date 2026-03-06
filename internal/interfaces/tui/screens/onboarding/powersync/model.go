package powersync

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/components/progress"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
	sessionruntime "github.com/usetero/cli/internal/runtime/session"
)

type Session interface {
	Status() sessionruntime.Status
}

// Model renders the onboarding PowerSync readiness screen.
type Model struct {
	session  Session
	theme    theme.Theme
	spinner  spinner.Model
	progress *progress.Model
}

var _ screen.Model = (*Model)(nil)

func New(session Session, appTheme theme.Theme) *Model {
	if session == nil {
		panic("powersync session is required")
	}
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &Model{
		session:  session,
		theme:    appTheme,
		spinner:  sp,
		progress: progress.New(appTheme, 40),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tick, ok := msg.(spinner.TickMsg)
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(tick)
	return m, cmd
}

func (m *Model) View() tea.View {
	lines := []string{
		m.theme.Text.Section.Render("Syncing your account data..."),
	}

	switch state := m.session.Status().Sync.(type) {
	case *pssyncer.Ready:
		lines = append(lines, "", m.theme.Text.Body.Render("PowerSync is ready."))
	case *pssyncer.Error:
		lines = append(lines, "", m.theme.Text.Error.Render("Sync failed: "+state.Err.Error()))
	case *pssyncer.Connecting:
		lines = append(lines, "", m.theme.Text.Body.Render(m.spinner.View()+" Connecting..."))
	case *pssyncer.Syncing:
		lines = append(lines, "", m.theme.Text.Body.Render(m.spinner.View()+" Syncing..."))
		if state.Progress != nil && state.Progress.Total > 0 {
			percent := float64(state.Progress.Downloaded) / float64(state.Progress.Total) * 100.0
			lines = append(lines,
				"",
				m.progress.ViewAs(percent),
				m.theme.Text.Muted.Render(fmt.Sprintf("%d / %d rows", state.Progress.Downloaded, state.Progress.Total)),
			)
		}
	case *pssyncer.Reconnecting:
		lines = append(lines, "", m.theme.Text.Body.Render(m.spinner.View()+" Reconnecting..."))
	default:
		lines = append(lines, "", m.theme.Text.Body.Render(m.spinner.View()+" Initializing sync..."))
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m *Model) SetSize(width, _ int) {
	barWidth := width - 12
	if barWidth > 60 {
		barWidth = 60
	}
	m.progress.SetWidth(barWidth)
}

func (m *Model) ShortHelp() []key.Binding { return nil }
