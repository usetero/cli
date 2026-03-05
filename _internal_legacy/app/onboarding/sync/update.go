package sync

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	appevents "github.com/usetero/cli/internal/app/events"
	"github.com/usetero/cli/internal/core/bootstrap"
	"github.com/usetero/cli/internal/powersync"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case appevents.SyncStateChanged:
		if _, ok := msg.State.(*powersync.Ready); ok {
			m.scope.Info("sync completed")
			return func() tea.Msg { return bootstrap.SyncComplete{} }
		}
		return nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}

	return m.progress.Update(msg)
}
