package datadog

import "charm.land/bubbles/v2/key"

var (
	regionUpKey     = key.NewBinding(key.WithKeys("up", "k"))
	regionDownKey   = key.NewBinding(key.WithKeys("down", "j"))
	regionSelectKey = key.NewBinding(key.WithKeys("enter"))
)
