package visor

import (
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

// Model owns the assistant response surface above the command input.
type Model struct {
	theme    theme.Theme
	width    int
	title    string
	detail   string
	expanded bool
}

var _ core.Model = (*Model)(nil)

// New constructs a visor model.
func New(appTheme theme.Theme) *Model {
	return &Model{theme: appTheme}
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (m *Model) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

// Show updates the visible visor message.
func (m *Model) Show(title string, detail string, expanded bool) {
	m.title = strings.TrimSpace(title)
	m.detail = strings.TrimSpace(detail)
	m.expanded = expanded
}

// Clear removes the visor message.
func (m *Model) Clear() {
	m.title = ""
	m.detail = ""
	m.expanded = false
}

// Expand expands the visor message surface.
func (m *Model) Expand() { m.expanded = true }

// Collapse collapses the visor message surface.
func (m *Model) Collapse() { m.expanded = false }

// ApplyState updates the visor from page shell state.
func (m *Model) ApplyState(input *core.Input) {
	if input == nil {
		m.Clear()
		return
	}

	title := strings.TrimSpace(input.Title)
	detail := strings.TrimSpace(input.Detail)
	if title == "" && detail == "" {
		label := strings.TrimSpace(input.Label)
		switch parts := strings.SplitN(label, "\n\n", 2); len(parts) {
		case 2:
			title = strings.TrimSpace(parts[0])
			detail = strings.TrimSpace(parts[1])
		default:
			title = label
		}
	}
	if title == "" && detail == "" {
		m.Clear()
		return
	}
	m.Show(title, detail, false)
}

func (m *Model) PreferredHeight(width int) int {
	title := strings.TrimSpace(m.title)
	detail := strings.TrimSpace(m.detail)
	if title == "" && detail == "" {
		return 0
	}

	innerWidth := width - visorPadX*2
	if innerWidth < 1 {
		innerWidth = 1
	}

	lines := wrappedLineCount(title, innerWidth)
	if detail != "" {
		lines++
		if m.expanded {
			lines += wrappedLineCount(detail, innerWidth)
		} else {
			limit := collapsedLines - 2
			if limit < 1 {
				limit = 1
			}
			explicitLines := strings.Split(detail, "\n")
			if len(explicitLines) > limit {
				lines += limit + 1
			} else {
				lines += wrappedLineCount(detail, innerWidth)
			}
		}
	}
	return lines
}

func wrappedLineCount(value string, width int) int {
	if width < 1 {
		width = 1
	}
	total := 0
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		if line == "" {
			total++
			continue
		}
		total += int(math.Ceil(float64(len([]rune(line))) / float64(width)))
	}
	if total < 1 {
		return 1
	}
	return total
}
