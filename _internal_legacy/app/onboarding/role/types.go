package role

import "charm.land/bubbles/v2/key"

// savedRoleLoadedMsg is sent when checking for a saved role.
type savedRoleLoadedMsg struct {
	role string
}

var (
	upKey     = key.NewBinding(key.WithKeys("up", "k"))
	downKey   = key.NewBinding(key.WithKeys("down", "j"))
	selectKey = key.NewBinding(key.WithKeys("enter"))
)
