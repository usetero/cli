package powersync

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/components/loading"
	"github.com/usetero/cli/internal/interfaces/tui/components/progress"
	"github.com/usetero/cli/internal/interfaces/tui/present"
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
	loading  *loading.Model
	progress *progress.Model
}

var _ screen.Model = (*Model)(nil)

func New(session Session, appTheme theme.Theme) *Model {
	if session == nil {
		panic("powersync session is required")
	}
	return &Model{
		session:  session,
		theme:    appTheme,
		loading:  loading.NewSpinner(appTheme, "Initializing sync..."),
		progress: progress.New(appTheme, 40),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.loading.Update(msg)
	if model, ok := next.(*loading.Model); ok {
		m.loading = model
	}
	return m, cmd
}

func (m *Model) View() tea.View {
	switch state := m.session.Status().Sync.(type) {
	case *pssyncer.Ready:
		return present.View(m.theme, present.StatusBlock(
			"Syncing your account data...",
			present.Success("PowerSync is ready."),
		))
	case *pssyncer.Error:
		return present.View(m.theme, present.ErrorCard(present.BlockGap(
			1,
			present.Error("Sync failed."),
			present.Body("Sync failed: "+state.Err.Error()),
		)))
	case *pssyncer.Connecting:
		m.loading.SetLabel("Connecting...")
		return present.View(m.theme, present.StatusBlock(
			"Syncing your account data...",
			present.Raw(m.loading.View().Content),
		))
	case *pssyncer.Syncing:
		m.loading.SetLabel("Syncing...")
		parts := []present.BlockItem{present.Raw(m.loading.View().Content)}
		if state.Progress != nil && state.Progress.Total > 0 {
			percent := float64(state.Progress.Downloaded) / float64(state.Progress.Total) * 100.0
			parts = append(parts,
				present.Raw(m.progress.ViewAs(percent)),
				present.Muted(fmt.Sprintf("%d / %d rows", state.Progress.Downloaded, state.Progress.Total)),
			)
		}
		return present.View(m.theme, present.StatusBlock("Syncing your account data...", parts...))
	case *pssyncer.Reconnecting:
		m.loading.SetLabel("Reconnecting...")
		return present.View(m.theme, present.StatusBlock(
			"Syncing your account data...",
			present.Raw(m.loading.View().Content),
		))
	default:
		m.loading.SetLabel("Initializing sync...")
		return present.View(m.theme, present.StatusBlock(
			"Syncing your account data...",
			present.Raw(m.loading.View().Content),
		))
	}
}

func (m *Model) SetSize(width, _ int) {
	barWidth := width - 12
	if barWidth > 60 {
		barWidth = 60
	}
	m.progress.SetWidth(barWidth)
}

func (m *Model) ShortHelp() []key.Binding { return nil }
