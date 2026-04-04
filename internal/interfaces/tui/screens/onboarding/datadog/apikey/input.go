package datadogapikey

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var openDocsBinding = key.NewBinding(
	key.WithKeys("ctrl+o"),
	key.WithHelp("ctrl+o", "open Datadog"),
)

var openTeroDocsBinding = key.NewBinding(
	key.WithKeys("ctrl+d"),
	key.WithHelp("ctrl+d", "Tero docs"),
)

func keyMatchesOpenDocs(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, openDocsBinding)
}

func keyMatchesOpenTeroDocs(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, openTeroDocsBinding)
}
