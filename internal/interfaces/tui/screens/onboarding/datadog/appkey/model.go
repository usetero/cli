package datadogappkey

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	cmdinput "github.com/usetero/cli/internal/interfaces/tui/components/commandbar/input"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the Datadog app-key page state.
type Model struct {
	site integrations.DatadogSite
}

var _ core.Model = (*Model)(nil)
var _ core.InputProvider = (*Model)(nil)
var _ core.HelpProvider = (*Model)(nil)

func New(theme.Theme) *Model { return &Model{} }

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case cmdinput.SubmittedMsg:
		return m, func() tea.Msg { return SubmittedMsg{AppKey: typed.Text} }
	case tea.KeyPressMsg:
		if openDocsBinding.Enabled() && keyMatchesOpenDocs(typed) {
			return m, m.openBrowser()
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) View() tea.View { return tea.NewView("") }

func (m *Model) SetSize(width, height int) {}

func (m *Model) Input() *core.Input {
	return &core.Input{
		Kind:        core.InputText,
		Label:       "Paste a Datadog app key from a service account.",
		Placeholder: "Paste app key",
	}
}

func (m *Model) ShortHelp() []key.Binding {
	if !m.site.Valid() {
		return nil
	}
	return []key.Binding{openDocsBinding}
}

func (m *Model) SetSite(site integrations.DatadogSite) {
	m.site = site
}
