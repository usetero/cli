package thinking

import (
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

const (
	fps           = 20
	defaultSize   = 12
	birthDuration = time.Second
	ellipsisSpeed = 8
)

var (
	scrambleChars = []rune("0123456789abcdefABCDEF~!@#$%^&*()+=_")
	ellipsis      = []string{"", ".", "..", "..."}
	nextID        atomic.Int64
	rampCache     = make(map[int][]lipgloss.Style)
	rampCacheMu   sync.RWMutex
)

// TickMsg advances one animation frame.
type TickMsg struct {
	id int
}

// Settings configures the thinking indicator.
type Settings struct {
	Size  int
	Label string
}

// Model renders an animated thinking indicator.
type Model struct {
	id           int
	theme        theme.Theme
	size         int
	label        string
	startTime    time.Time
	birthOffsets []time.Duration
	frame        int
	ellipsisStep int
	colorRamp    []lipgloss.Style
}

// New constructs a thinking indicator.
func New(appTheme theme.Theme, settings Settings) *Model {
	size := settings.Size
	if size <= 0 {
		size = defaultSize
	}
	m := &Model{
		id:        int(nextID.Add(1)),
		theme:     appTheme,
		size:      size,
		label:     settings.Label,
		startTime: time.Now(),
	}
	m.resetBirthOffsets()
	m.colorRamp = m.getOrBuildColorRamp()
	return m
}

// Init starts the animation.
func (m *Model) Init() tea.Cmd { return m.tick() }

// Update advances the animation.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	tick, ok := msg.(TickMsg)
	if !ok || tick.id != m.id {
		return m, nil
	}
	m.frame = (m.frame + 1) % len(m.colorRamp)
	m.ellipsisStep = (m.ellipsisStep + 1) % (ellipsisSpeed * len(ellipsis))
	return m, m.tick()
}

// View renders the current animation frame.
func (m *Model) View() tea.View {
	var b strings.Builder
	elapsed := time.Since(m.startTime)
	labelRunes := []rune(m.label)
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Background)
	allBorn := elapsed >= birthDuration

	totalWidth := m.size
	if len(labelRunes) > 0 {
		totalWidth += 1 + len(labelRunes)
	}

	for i := range totalWidth {
		if !allBorn && elapsed < m.birthOffsets[i] {
			b.WriteString(labelStyle.Render(" "))
			continue
		}
		if i < m.size {
			colorIdx := (i + m.frame) % len(m.colorRamp)
			char := scrambleChars[rand.IntN(len(scrambleChars))]
			b.WriteString(m.colorRamp[colorIdx].Background(m.theme.Background).Render(string(char)))
			continue
		}
		if i == m.size {
			b.WriteString(labelStyle.Render(" "))
			continue
		}
		labelIdx := i - m.size - 1
		b.WriteString(labelStyle.Render(string(labelRunes[labelIdx])))
	}

	if allBorn && m.label != "" {
		ellipsisIdx := m.ellipsisStep / ellipsisSpeed
		b.WriteString(labelStyle.Render(ellipsis[ellipsisIdx]))
	}

	return tea.NewView(b.String())
}

// SetLabel updates the label and restarts the birth animation.
func (m *Model) SetLabel(label string) {
	if m.label == label {
		return
	}
	m.label = label
	m.startTime = time.Now()
	m.resetBirthOffsets()
}

// SetSize updates the scramble size.
func (m *Model) SetSize(size int) {
	if size <= 0 {
		size = defaultSize
	}
	if m.size == size {
		return
	}
	m.size = size
	m.colorRamp = m.getOrBuildColorRamp()
	m.resetBirthOffsets()
}

func (m *Model) tick() tea.Cmd {
	id := m.id
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg {
		return TickMsg{id: id}
	})
}

func (m *Model) resetBirthOffsets() {
	labelLen := len([]rune(m.label))
	totalWidth := m.size
	if labelLen > 0 {
		totalWidth += 1 + labelLen
	}
	m.birthOffsets = make([]time.Duration, totalWidth)
	for i := range m.birthOffsets {
		m.birthOffsets[i] = time.Duration(rand.Int64N(int64(birthDuration)))
	}
}

func (m *Model) getOrBuildColorRamp() []lipgloss.Style {
	rampCacheMu.RLock()
	ramp, ok := rampCache[m.size]
	rampCacheMu.RUnlock()
	if ok {
		return ramp
	}

	colors := m.theme.Gradients.Brand
	r := colors.Ramp(m.size * 2)
	styles := make([]lipgloss.Style, len(r))
	for i, c := range r {
		styles[i] = lipgloss.NewStyle().Foreground(c)
	}

	rampCacheMu.Lock()
	rampCache[m.size] = styles
	rampCacheMu.Unlock()
	return styles
}
