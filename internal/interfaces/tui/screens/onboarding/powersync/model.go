package powersyncready

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/components/progressbar"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/present"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

// Model renders the final cold-start sync wait step.
type Model struct {
	theme    theme.Theme
	progress *progressbar.Model
	status   accountruntime.Status
	width    int
}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{
		theme:    appTheme,
		progress: progressbar.New(appTheme, 32),
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *Model) View() tea.View {
	parts := []string{
		m.theme.Text.Section.Render("Preparing your workspace"),
		"",
		m.theme.Text.Body.Render(m.statusLine()),
	}

	if pct, ok := m.percent(); ok {
		parts = append(parts, "", m.progress.ViewAs(pct))
	}
	if detail := m.detailLine(); detail != "" {
		parts = append(parts, "", m.theme.Text.Subtle.Render(detail))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tea.NewView(present.Panel(m.theme.OnSurface(), m.width, content))
}

func (m *Model) SetSize(width, _ int) {
	if width < 1 {
		width = 1
	}
	m.width = width
	m.progress.SetWidth(min(48, present.PanelInnerWidth(width)))
}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Label: "We're preparing your local workspace. This only blocks the first time for an account.",
	}
}

func (m *Model) SetStatus(status accountruntime.Status) {
	m.status = status
}

func (m *Model) statusLine() string {
	switch typed := m.status.Sync.(type) {
	case *pssyncer.Ready:
		return "Your workspace is ready."
	case *pssyncer.Error:
		return fmt.Sprintf("Sync failed: %v", typed.Err)
	case *pssyncer.Reconnecting:
		return "Reconnecting to continue sync."
	case *pssyncer.Syncing:
		return "Syncing your account data."
	case *pssyncer.Connecting:
		return "Connecting to sync."
	default:
		return "Starting the local sync engine."
	}
}

func (m *Model) detailLine() string {
	switch typed := m.status.Sync.(type) {
	case *pssyncer.Syncing:
		if typed.Progress != nil && typed.Progress.Total > 0 {
			return fmt.Sprintf("%d / %d rows", typed.Progress.Downloaded, typed.Progress.Total)
		}
	}
	return ""
}

func (m *Model) percent() (float64, bool) {
	switch typed := m.status.Sync.(type) {
	case *pssyncer.Syncing:
		if typed.Progress != nil && typed.Progress.Total > 0 {
			return float64(typed.Progress.Downloaded) / float64(typed.Progress.Total) * 100, true
		}
	}
	return 0, false
}
