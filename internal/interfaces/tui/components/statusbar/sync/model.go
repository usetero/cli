package sync

import (
	tea "charm.land/bubbletea/v2"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/interfaces/tui/events"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type state uint8

const (
	stateMuted state = iota
	stateConnecting
	stateReady
	stateReconnecting
	stateSyncing
	stateError
)

// Model renders the runtime connectivity indicator.
type Model struct {
	theme theme.Theme
	state state
}

func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	typed, ok := msg.(events.AccountRuntimeUpdatedMsg)
	if !ok {
		return m, nil
	}

	switch typed.Status.Sync.(type) {
	case *pssyncer.Ready:
		m.state = stateReady
	case *pssyncer.Connecting:
		m.state = stateConnecting
	case *pssyncer.Reconnecting:
		m.state = stateReconnecting
	case *pssyncer.Syncing:
		m.state = stateSyncing
	case *pssyncer.Error, *pssyncer.Disconnected:
		m.state = stateError
	default:
		m.state = stateMuted
	}

	return m, nil
}

func (m *Model) SetSize(_, _ int) {}
