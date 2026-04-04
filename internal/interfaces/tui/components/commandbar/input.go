package commandbar

import (
	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

func confirmBinding(action string) key.Binding {
	return key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", action),
	)
}

// ShortHelp exposes the active command-bar child bindings when that child owns interaction.
func (m *Model) ShortHelp() []key.Binding {
	if bindings := m.commandShortHelp(); bindings != nil {
		return bindings
	}
	if m.err != nil {
		if action := m.err.Action; action != "" {
			return []key.Binding{confirmBinding(action)}
		}
		return nil
	}

	if m.busy != nil {
		return nil
	}

	if m.mode == ModeAction && m.input != nil && m.input.Kind == core.InputConfirm && m.input.Action != "" {
		return []key.Binding{confirmBinding(m.input.Action)}
	}

	if bindings := m.children.action.ShortHelp(); bindings != nil {
		return bindings
	}

	return nil
}
