package stepkit

import (
	"charm.land/bubbles/v2/key"

	"github.com/usetero/cli/internal/tea/components/remotelist"
)

// RemoteListShortHelp renders a consistent help model for remotelist-based steps.
func RemoteListShortHelp(list *remotelist.Model, extras ...key.Binding) []key.Binding {
	if list.IsLoading() {
		return nil
	}
	if list.HasError() {
		return []key.Binding{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		}
	}
	bindings := list.ShortHelp()
	return append(bindings, extras...)
}
