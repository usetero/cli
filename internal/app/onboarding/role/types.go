package role

import "charm.land/bubbles/v2/key"

// savedRoleMsg is sent when checking for a saved role.
type savedRoleMsg struct {
	role string
}

var (
	upKey     = key.NewBinding(key.WithKeys("up", "k"))
	downKey   = key.NewBinding(key.WithKeys("down", "j"))
	selectKey = key.NewBinding(key.WithKeys("enter"))
)
