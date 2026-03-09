package datadogapikey

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/browser"
	"github.com/usetero/cli/internal/interfaces/tui/components/form"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

const fieldAPIKey form.FieldID = "api_key"

var submitBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "validate"),
)

var openWithEnterBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "open"),
)

var openBinding = key.NewBinding(
	key.WithKeys("o"),
	key.WithHelp("o", "open"),
)

// Model owns Datadog API key input UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme
	form  *form.Model
	site  integrations.DatadogSite
}

var _ screen.Model = (*Model)(nil)

// New constructs the Datadog API key model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	return &Model{
		scope: scope,
		theme: appTheme,
		form: form.New(appTheme, form.FieldSpec{
			ID:          fieldAPIKey,
			Label:       "API key: ",
			Placeholder: "Datadog API key",
		}),
	}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract. API key input currently ignores dimensions.
func (m *Model) SetSize(width, height int) { m.form.SetSize(width, height) }

// Reset clears current input state.
func (m *Model) Reset() { m.form.Reset() }

// SetSite updates the active Datadog site for browser guidance.
func (m *Model) SetSite(site integrations.DatadogSite) { m.site = site }

// Update handles local API key input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, _ := m.form.Update(msg)
	if formModel, ok := next.(*form.Model); ok {
		m.form = formModel
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, openBinding) || (key.Matches(keyMsg, submitBinding) && strings.TrimSpace(m.form.Value(fieldAPIKey)) == "") {
		return m, func() tea.Msg {
			return browser.OpenRequestedMsg{URL: integrations.DatadogAPIKeyURL(m.site)}
		}
	}
	if key.Matches(keyMsg, submitBinding) {
		apiKey, err := integrations.ParseDatadogAPIKey(strings.TrimSpace(m.form.Value(fieldAPIKey)))
		if err != nil {
			return m, nil
		}
		m.scope.Info("datadog api key submitted")
		return m, func() tea.Msg { return SubmittedMsg{APIKey: apiKey} }
	}
	return m, nil
}

// View renders the Datadog API key input screen.
func (m *Model) View() tea.View {
	return present.View(m.theme, present.Section(
		"Enter Datadog API key:",
		present.StackGap(
			1,
			present.Muted("Open Datadog, create an API key for the selected region, then paste it here."),
			present.Subtle("Press o to open the Datadog API keys page in your browser."),
			present.Raw(m.form.View().Content),
		),
	))
}

// ShortHelp returns datadog-api-key key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding{}, m.form.ShortHelp()...)
	enterBinding := submitBinding
	if strings.TrimSpace(m.form.Value(fieldAPIKey)) == "" {
		enterBinding = openWithEnterBinding
	}
	bindings = append(bindings, openBinding, enterBinding)
	return bindings
}
