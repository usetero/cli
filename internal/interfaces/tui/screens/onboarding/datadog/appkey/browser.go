package datadogappkey

import (
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/interfaces/tui/ui/browser"
)

var openBrowser = browser.Open

const teroDatadogPermissionsDocsURL = "https://docs.usetero.com/integrations/datadog#permissions"

type browserOpenedMsg struct {
	Err error
}

func (m *Model) openBrowser() tea.Cmd {
	if !m.site.Valid() {
		return nil
	}
	url := integrations.DatadogAppKeyURL(m.site)
	return func() tea.Msg {
		return browserOpenedMsg{Err: openBrowser(url)}
	}
}

func (m *Model) openPermissionsDocs() tea.Cmd {
	return func() tea.Msg {
		return browserOpenedMsg{Err: openBrowser(teroDatadogPermissionsDocsURL)}
	}
}
