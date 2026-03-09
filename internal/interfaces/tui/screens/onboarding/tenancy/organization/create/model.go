package organizationcreate

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/components/form"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/screen"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

const fieldName form.FieldID = "name"

var submitBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "create"),
)

// Model owns organization-creation UI state.
type Model struct {
	scope logging.Scope
	theme theme.Theme
	form  *form.Model
}

var _ screen.Model = (*Model)(nil)

// New constructs the onboarding organization-creation model.
func New(scope logging.Scope, appTheme theme.Theme) *Model {
	return &Model{
		scope: scope,
		theme: appTheme,
		form: form.New(appTheme, form.FieldSpec{
			ID:          fieldName,
			Label:       "Name: ",
			Placeholder: "Organization name",
		}),
	}
}

// Init satisfies Bubble Tea model requirements.
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize is part of the screen contract. Create currently ignores dimensions.
func (m *Model) SetSize(width, height int) { m.form.SetSize(width, height) }

// Name returns the current organization name input.
func (m *Model) Name() string {
	return m.form.Value(fieldName)
}

// Reset clears current input state.
func (m *Model) Reset() {
	m.form.Reset()
}

// Update handles local organization-creation input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, _ := m.form.Update(msg)
	if formModel, ok := next.(*form.Model); ok {
		m.form = formModel
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if key.Matches(keyMsg, submitBinding) {
		create, err := (tenancy.OrganizationCreate{Name: m.form.Value(fieldName)}).Validate()
		if err != nil {
			return m, nil
		}
		m.scope.Info("organization create submitted", "name", create.Name)
		return m, func() tea.Msg { return CreatedMsg{Create: create} }
	}
	return m, nil
}

// View renders the organization-creation screen.
func (m *Model) View() tea.View {
	return present.View(m.theme, present.Section(
		"Create your organization:",
		present.Raw(m.form.View().Content),
	))
}

// ShortHelp returns organization-create key bindings.
func (m *Model) ShortHelp() []key.Binding {
	bindings := append([]key.Binding{}, m.form.ShortHelp()...)
	bindings = append(bindings, submitBinding)
	return bindings
}
