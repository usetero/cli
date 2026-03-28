package datadogapikey

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var openDocsBinding = key.NewBinding(
	key.WithKeys("o"),
	key.WithHelp("o", "open docs"),
)

func keyMatchesOpenDocs(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, openDocsBinding)
}
