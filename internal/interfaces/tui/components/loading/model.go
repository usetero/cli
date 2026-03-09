package loading

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/components/spinner"
	"github.com/usetero/cli/internal/interfaces/tui/components/thinking"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// Variant selects the loading indicator style.
type Variant uint8

const (
	VariantSpinner Variant = iota
	VariantThinking
)

// Model renders a shared loading surface with a themed indicator and optional detail.
type Model struct {
	theme    theme.Theme
	variant  Variant
	label    string
	detail   string
	spinner  *spinner.Model
	thinking *thinking.Model
}

// NewSpinner constructs a spinner-backed loading model.
func NewSpinner(appTheme theme.Theme, label string) *Model {
	m := &Model{
		theme:   appTheme,
		variant: VariantSpinner,
		spinner: spinner.New(appTheme),
	}
	m.SetLabel(label)
	return m
}

// NewThinking constructs a thinking-backed loading model.
func NewThinking(appTheme theme.Theme, label string) *Model {
	return &Model{
		theme:   appTheme,
		variant: VariantThinking,
		label:   strings.TrimSpace(label),
		thinking: thinking.New(appTheme, thinking.Settings{
			Size: 12,
		}),
	}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	switch m.variant {
	case VariantThinking:
		return m.thinking.Init()
	default:
		return m.spinner.Init()
	}
}

// Update advances the active loading indicator.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.variant {
	case VariantThinking:
		next, cmd := m.thinking.Update(msg)
		if model, ok := next.(*thinking.Model); ok {
			m.thinking = model
		}
		return m, cmd
	default:
		next, cmd := m.spinner.Update(msg)
		if model, ok := next.(*spinner.Model); ok {
			m.spinner = model
		}
		return m, cmd
	}
}

// SetLabel updates the primary loading copy.
func (m *Model) SetLabel(label string) {
	m.label = strings.TrimSpace(label)
}

// SetDetail updates optional secondary copy.
func (m *Model) SetDetail(detail string) {
	m.detail = strings.TrimSpace(detail)
}

// View renders the current loading state.
func (m *Model) View() tea.View {
	lines := make([]string, 0, 2)
	switch m.variant {
	case VariantThinking:
		line := m.thinking.View().Content
		if m.label != "" {
			line = lipgloss.JoinHorizontal(lipgloss.Left, line, " ", m.theme.Text.Body.Render(m.label))
		}
		lines = append(lines, line)
	case VariantSpinner:
		line := m.spinner.View().Content
		if m.label != "" {
			line = line + " " + m.theme.Text.Body.Render(m.label)
		}
		lines = append(lines, line)
	}
	if m.detail != "" {
		lines = append(lines, m.theme.Text.Muted.Render(m.detail))
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
