package accountcreate

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/textinput"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var submitBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "create"),
)

// Model owns account-creation UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme
	input *textinput.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the onboarding account-creation model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	input := textinput.New(appTheme)
	input.SetPlaceholder("Account name")
	return &Model{scope: scope, theme: appTheme, input: input}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract. Create currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// Reset clears current input state.
func (m *Model) Reset() { m.input.Reset() }

// Name returns current account name input.
func (m *Model) Name() string { return m.input.Value() }

// Update handles local account-creation input.
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
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			return m, nil
		}
		create := tenancy.AccountCreate{Name: name}
		m.scope.Info("account create submitted", "name", create.Name)
		return m, func() tea.Msg { return CreatedMsg{Create: create} }
	}
	return m, nil
}

// View renders the account-creation screen.
func (m *Model) View() tea.View {
	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Text.Section.Render("Create your account:"),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.theme.Input.Label.Render("Name: "),
			m.input.View().Content,
		),
	))
}

// ShortHelp returns account-create key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding{}, m.input.ShortHelp()...)
	bindings = append(bindings, submitBinding)
	return bindings
}
