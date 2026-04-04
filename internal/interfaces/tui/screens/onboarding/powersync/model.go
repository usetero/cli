package powersyncready

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
	accountruntime "github.com/usetero/cli/internal/runtime/account"
)

// Model renders the final cold-start sync wait step.
type Model struct {
	theme  theme.Theme
	status accountruntime.Status
}

var _ core.Model = (*Model)(nil)
var _ core.BusyProvider = (*Model)(nil)
var _ core.ErrorProvider = (*Model)(nil)

func New(appTheme theme.Theme) *Model {
	return &Model{
		theme: appTheme,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *Model) View() tea.View { return tea.NewView("") }

func (m *Model) SetSize(_, _ int) {}

func (m *Model) Busy() *core.Busy {
	if _, ok := m.status.Sync.(*pssyncer.Error); ok {
		return nil
	}
	busy := &core.Busy{
		Label:  "Preparing your workspace",
		Detail: "We're preparing your local workspace. This only blocks the first time for an account.",
		Status: m.statusLine(),
	}
	if progress := m.progress(); progress != nil {
		busy.Progress = progress
	}
	return busy
}

func (m *Model) Error() *core.Error {
	switch typed := m.status.Sync.(type) {
	case *pssyncer.Error:
		message := "Failed to prepare your workspace."
		if typed.Err == nil {
			return &core.Error{Message: message}
		}
		return &core.Error{
			Message: message,
			Detail:  typed.Err.Error(),
		}
	default:
		return nil
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

func (m *Model) progress() *core.Progress {
	switch typed := m.status.Sync.(type) {
	case *pssyncer.Syncing:
		if typed.Progress != nil && typed.Progress.Total > 0 {
			return &core.Progress{
				Current: typed.Progress.Downloaded,
				Total:   typed.Progress.Total,
			}
		}
	}
	return nil
}
