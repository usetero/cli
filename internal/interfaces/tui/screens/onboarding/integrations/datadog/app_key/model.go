package datadogappkey

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

const (
	fieldName   form.FieldID = "name"
	fieldAppKey form.FieldID = "app_key"
)

var submitBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "connect"),
)

var openWithEnterBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "open"),
)

var openBinding = key.NewBinding(
	key.WithKeys("o"),
	key.WithHelp("o", "open"),
)

// Model owns Datadog account name + app key input UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme
	form  *form.Model
	site  integrations.DatadogSite
}

var _ screen.Model = (*Model)(nil)

// New constructs the Datadog app key model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	return &Model{
		scope: scope,
		theme: appTheme,
		form: form.New(
			appTheme,
			form.FieldSpec{ID: fieldName, Label: "Name: ", Placeholder: "Datadog account name"},
			form.FieldSpec{ID: fieldAppKey, Label: "App key: ", Placeholder: "Datadog application key"},
		),
	}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract. App key input currently ignores dimensions.
func (m *Model) SetSize(width, height int) { m.form.SetSize(width, height) }

// Reset clears current input state.
func (m *Model) Reset() {
	m.form.Reset()
}

// SetSite updates the active Datadog site for browser guidance.
func (m *Model) SetSite(site integrations.DatadogSite) { m.site = site }

// Update handles local name/app key input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, _ := m.form.Update(msg)
	if formModel, ok := next.(*form.Model); ok {
		m.form = formModel
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, openBinding) || (key.Matches(keyMsg, submitBinding) && strings.TrimSpace(m.form.Value(fieldAppKey)) == "") {
		return m, func() tea.Msg {
			return browser.OpenRequestedMsg{URL: integrations.DatadogAppKeyURL(m.site)}
		}
	}
	if !key.Matches(keyMsg, submitBinding) {
		return m, nil
	}

	name, err := integrations.ParseDatadogAccountName(strings.TrimSpace(m.form.Value(fieldName)))
	if err != nil {
		return m, nil
	}
	appKey, err := integrations.ParseDatadogAppKey(strings.TrimSpace(m.form.Value(fieldAppKey)))
	if err != nil {
		return m, nil
	}
	m.scope.Info("datadog app key submitted", "name", name.String())
	return m, func() tea.Msg { return SubmittedMsg{Name: name, AppKey: appKey} }
}

// View renders the Datadog app key input screen.
func (m *Model) View() tea.View {
	return present.View(m.theme, present.Section(
		"Finish Datadog setup:",
		present.StackGap(
			1,
			present.Muted("Datadog requires both keys: the API key identifies your account, and the application key authorizes read access for discovery."),
			present.Subtle("Press o to open the Datadog service accounts page and create an application key for Tero."),
			present.Raw(m.form.View().Content),
		),
	))
}

// ShortHelp returns app-key step key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding{}, m.form.ShortHelp()...)
	enterBinding := submitBinding
	if strings.TrimSpace(m.form.Value(fieldAppKey)) == "" {
		enterBinding = openWithEnterBinding
	}
	bindings = append(bindings, openBinding, enterBinding)
	return bindings
}
