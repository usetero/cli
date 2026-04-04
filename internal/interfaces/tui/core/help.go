package core

import "charm.land/bubbles/v2/key"

// HelpProvider exposes key bindings handled by a model.
type HelpProvider interface {
	ShortHelp() []key.Binding
}
