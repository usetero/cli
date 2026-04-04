package datadogappkey

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var openDocsBinding = key.NewBinding(
	key.WithKeys("ctrl+o"),
	key.WithHelp("ctrl+o", "open Datadog"),
)

var openPermissionsDocsBinding = key.NewBinding(
	key.WithKeys("ctrl+d"),
	key.WithHelp("ctrl+d", "permissions docs"),
)

func keyMatchesOpenDocs(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, openDocsBinding)
}

func keyMatchesOpenPermissionsDocs(msg tea.KeyPressMsg) bool {
	return key.Matches(msg, openPermissionsDocsBinding)
}
