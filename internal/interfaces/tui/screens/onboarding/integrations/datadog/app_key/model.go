package datadogappkey

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

const (
	fieldName = iota
	fieldAppKey
)

var (
	tabBinding = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch field"),
	)
	submitBinding = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	)
)

// Model owns Datadog account name + app key input UI state.
type Model struct {
	scope  logging.Scope
	theme  theme.Theme
	active int
	name   *textinput.Model
	appKey *textinput.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the Datadog app key model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	nameInput := textinput.New(appTheme)
	nameInput.SetPlaceholder("Datadog account name")
	appKeyInput := textinput.New(appTheme)
	appKeyInput.SetPlaceholder("Datadog application key")

	return &Model{
		scope:  scope,
		theme:  appTheme,
		name:   nameInput,
		appKey: appKeyInput,
	}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd { return nil }

// SetSize is part of the screen contract. App key input currently ignores dimensions.
func (m *Model) SetSize(_, _ int) {}

// Reset clears current input state.
func (m *Model) Reset() {
	m.active = fieldName
	m.name.Reset()
	m.appKey.Reset()
}

// Update handles local name/app key input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, tabBinding):
		if m.active == fieldName {
			m.active = fieldAppKey
		} else {
			m.active = fieldName
		}
		return m, nil
	case key.Matches(keyMsg, submitBinding):
		name, err := integrations.ParseDatadogAccountName(strings.TrimSpace(m.name.Value()))
		if err != nil {
			return m, nil
		}
		appKey, err := integrations.ParseDatadogAppKey(strings.TrimSpace(m.appKey.Value()))
		if err != nil {
			return m, nil
		}
		m.scope.Info("datadog app key submitted", "name", name.String())
		return m, func() tea.Msg { return SubmittedMsg{Name: name, AppKey: appKey} }
	default:
		if m.active == fieldName {
			next, _ := m.name.Update(msg)
			if model, ok := next.(*textinput.Model); ok {
				m.name = model
			}
		} else {
			next, _ := m.appKey.Update(msg)
			if model, ok := next.(*textinput.Model); ok {
				m.appKey = model
			}
		}
	}
	return m, nil
}

// View renders the Datadog app key input screen.
func (m *Model) View() tea.View {
	namePrefix := m.theme.Input.Inactive.Render("  ")
	appPrefix := m.theme.Input.Inactive.Render("  ")
	if m.active == fieldName {
		namePrefix = m.theme.Input.Active.Render("> ")
	} else {
		appPrefix = m.theme.Input.Active.Render("> ")
	}

	nameLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		namePrefix,
		m.theme.Input.Label.Render("Name: "),
		m.name.View().Content,
	)
	appKeyLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		appPrefix,
		m.theme.Input.Label.Render("App key: "),
		m.appKey.View().Content,
	)

	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Text.Section.Render("Finish Datadog setup:"),
		"",
		nameLine,
		appKeyLine,
	))
}

// ShortHelp returns app-key step key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		tabBinding,
	}
	if m.active == fieldName {
		bindings = append(bindings, m.name.ShortHelp()...)
	} else {
		bindings = append(bindings, m.appKey.ShortHelp()...)
	}
	bindings = append(bindings, submitBinding)
	return bindings
}
