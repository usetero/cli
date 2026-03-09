package form

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/textinput"
	"github.com/usetero/cli/internal/interfaces/tui/present"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var (
	nextFieldBinding = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	)
	prevFieldBinding = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	)
)

// FieldID identifies a form field.
type FieldID string

// FieldSpec configures one form field.
type FieldSpec struct {
	ID          FieldID
	Label       string
	Placeholder string
}

type field struct {
	id    FieldID
	label string
	input *textinput.Model
}

// Model owns shared multi-field form interaction over text inputs.
type Model struct {
	theme  theme.Theme
	fields []field
	active int
	width  int
}

// New constructs a form from field specs.
func New(appTheme theme.Theme, specs ...FieldSpec) *Model {
	if len(specs) == 0 {
		panic("form requires at least one field")
	}

	fields := make([]field, 0, len(specs))
	for _, spec := range specs {
		if spec.ID == "" {
			panic("form field id is required")
		}
		input := textinput.New(appTheme)
		input.SetPlaceholder(spec.Placeholder)
		fields = append(fields, field{
			id:    spec.ID,
			label: spec.Label,
			input: input,
		})
	}

	m := &Model{
		theme:  appTheme,
		fields: fields,
	}
	m.applyFocus()
	return m
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update handles field focus and forwards input to the active field.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, nextFieldBinding):
			if len(m.fields) > 1 {
				m.active = (m.active + 1) % len(m.fields)
				m.applyFocus()
			}
			return m, nil
		case key.Matches(keyMsg, prevFieldBinding):
			if len(m.fields) > 1 {
				m.active = (m.active - 1 + len(m.fields)) % len(m.fields)
				m.applyFocus()
			}
			return m, nil
		}
	}

	active := &m.fields[m.active]
	next, cmd := active.input.Update(msg)
	if input, ok := next.(*textinput.Model); ok {
		active.input = input
	}
	return m, cmd
}

// View renders the full form.
func (m *Model) View() tea.View {
	rows := make([]present.Node, 0, len(m.fields))
	labelWidth := 0
	for i := range m.fields {
		if w := lipgloss.Width(m.fields[i].label); w > labelWidth {
			labelWidth = w
		}
	}
	for i := range m.fields {
		prefix := m.theme.Input.Inactive.Render("  ")
		if i == m.active {
			prefix = m.theme.Input.Active.Render("> ")
		}
		label := m.theme.Text.Body.Width(labelWidth).Render(m.fields[i].label)
		rows = append(rows, present.Raw(present.Render(
			m.theme,
			present.Field(prefix, label, present.Raw(m.fields[i].input.View().Content)),
		)))
	}
	return present.View(m.theme, present.Stack(rows...))
}

// SetSize updates all field widths for the available body width.
func (m *Model) SetSize(width, _ int) {
	m.width = width
	labelWidth := 0
	for i := range m.fields {
		if w := lipgloss.Width(m.fields[i].label); w > labelWidth {
			labelWidth = w
		}
	}
	inputWidth := width - labelWidth - 2
	if inputWidth < 12 {
		inputWidth = 12
	}
	for i := range m.fields {
		m.fields[i].input.SetWidth(inputWidth)
	}
}

// ShortHelp returns shared form navigation help plus active input help.
func (m *Model) ShortHelp() []key.Binding {
	var bindings []key.Binding
	if len(m.fields) > 1 {
		bindings = append(bindings, nextFieldBinding, prevFieldBinding)
	}
	bindings = append(bindings, m.fields[m.active].input.ShortHelp()...)
	return bindings
}

// Value returns the current value for a field.
func (m *Model) Value(id FieldID) string {
	if field := m.find(id); field != nil {
		return field.input.Value()
	}
	return ""
}

// SetValue sets the value for a field.
func (m *Model) SetValue(id FieldID, value string) {
	if field := m.find(id); field != nil {
		field.input.SetValue(value)
	}
}

// Reset clears all inputs and focuses the first field.
func (m *Model) Reset() {
	for i := range m.fields {
		m.fields[i].input.Reset()
	}
	m.active = 0
	m.applyFocus()
}

// SetActive focuses a specific field by id.
func (m *Model) SetActive(id FieldID) bool {
	index := slices.IndexFunc(m.fields, func(f field) bool { return f.id == id })
	if index < 0 {
		return false
	}
	m.active = index
	m.applyFocus()
	return true
}

// ActiveField returns the active field id.
func (m *Model) ActiveField() FieldID {
	return m.fields[m.active].id
}

func (m *Model) applyFocus() {
	for i := range m.fields {
		if i == m.active {
			m.fields[i].input.Focus()
			continue
		}
		m.fields[i].input.Blur()
	}
}

func (m *Model) find(id FieldID) *field {
	for i := range m.fields {
		if m.fields[i].id == id {
			return &m.fields[i]
		}
	}
	return nil
}
