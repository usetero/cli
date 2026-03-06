package datadogapikey

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/textinput"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var submitBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "validate"),
)

// Model owns Datadog API key input UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme
	input *textinput.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the Datadog API key model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	input := textinput.New(appTheme)
	input.SetPlaceholder("Datadog API key")
	return &Model{scope: scope, theme: appTheme, input: input}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract. API key input currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// Reset clears current input state.
func (m *Model) Reset() { m.input.Reset() }

// Update handles local API key input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, _ := m.input.Update(msg)
	if input, ok := next.(*textinput.Model); ok {
		m.input = input
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, submitBinding) {
		apiKey, err := integrations.ParseDatadogAPIKey(strings.TrimSpace(m.input.Value()))
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
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Text.Section.Render("Enter Datadog API key:"),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.theme.Input.Label.Render("API key: "),
			m.input.View().Content,
		),
	))
}

// ShortHelp returns datadog-api-key key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding{}, m.input.ShortHelp()...)
	bindings = append(bindings, submitBinding)
	return bindings
}
