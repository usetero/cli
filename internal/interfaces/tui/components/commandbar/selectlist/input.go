package selectlist

import (
	"charm.land/bubbles/v2/key"
	baselist "github.com/usetero/cli/internal/interfaces/tui/components/selectlist"
)

func SelectBinding() key.Binding { return baselist.SelectBinding() }

func (m *Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}
