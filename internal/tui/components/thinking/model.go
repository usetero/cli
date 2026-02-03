package thinking

import (
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
)

const (
	fps            = 20
	charCount      = 12
	ellipsisFrames = 4
	ellipsisSpeed  = 8 // frames per ellipsis change
)

var (
	scrambleChars = []rune("0123456789abcdefABCDEF~!@#$%^&*()+=_")
	ellipsis      = []string{"", ".", "..", "..."}
)

// TickMsg triggers the next animation frame.
type TickMsg struct {
	id int
}

// Model is an animated thinking indicator with gradient scrambled characters.
type Model struct {
	id           int
	theme        *styles.Theme
	label        string
	frame        int
	ellipsisStep int
	colorRamp    []lipgloss.Style // pre-computed gradient styles
}

var nextID int

// New creates a new thinking indicator.
func New(theme *styles.Theme, label string) Model {
	nextID++
	m := Model{
		id:    nextID,
		theme: theme,
		label: label,
	}
	m.buildColorRamp()
	return m
}

// buildColorRamp pre-computes gradient styles for performance.
func (m *Model) buildColorRamp() {
	colors := m.theme.Colors
	// Create a cycling gradient: start -> end -> start
	ramp := styles.BlendColors(
		charCount*2,
		colors.Brand.GradientStart,
		colors.Brand.GradientEnd,
		colors.Brand.GradientStart,
	)
	m.colorRamp = make([]lipgloss.Style, len(ramp))
	for i, c := range ramp {
		m.colorRamp[i] = lipgloss.NewStyle().Foreground(c)
	}
}

// Init starts the animation.
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Update handles tick messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.id != m.id {
			return m, nil
		}
		m.frame = (m.frame + 1) % len(m.colorRamp)
		m.ellipsisStep = (m.ellipsisStep + 1) % (ellipsisSpeed * ellipsisFrames)
		return m, m.tick()
	}
	return m, nil
}

// View renders the thinking indicator.
func (m Model) View() string {
	var b strings.Builder

	// Render scrambled gradient characters
	for i := range charCount {
		colorIdx := (i + m.frame) % len(m.colorRamp)
		char := scrambleChars[rand.IntN(len(scrambleChars))]
		b.WriteString(m.colorRamp[colorIdx].Render(string(char)))
	}

	// Add label with ellipsis
	if m.label != "" {
		b.WriteString(" ")
		labelStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Page.TextMuted)
		ellipsisIdx := m.ellipsisStep / ellipsisSpeed
		b.WriteString(labelStyle.Render(m.label + ellipsis[ellipsisIdx]))
	}

	return b.String()
}

// SetLabel returns a Model with the given label.
func (m Model) SetLabel(label string) Model {
	m.label = label
	return m
}

// tick returns a command that triggers the next animation frame.
func (m Model) tick() tea.Cmd {
	id := m.id
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg {
		return TickMsg{id: id}
	})
}
