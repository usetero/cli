// Package thinking provides an animated thinking indicator with scrambled gradient text.
package thinking

import (
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/styles"
)

const (
	fps           = 20
	defaultSize   = 12
	birthDuration = time.Second // time for all characters to appear
	ellipsisSpeed = 8           // frames per ellipsis change
)

var (
	scrambleChars = []rune("0123456789abcdefABCDEF~!@#$%^&*()+=_")
	ellipsis      = []string{"", ".", "..", "..."}
	nextID        atomic.Int64
)

// TickMsg triggers the next animation frame.
type TickMsg struct {
	id int
}

// Settings configures the thinking indicator.
type Settings struct {
	Size  int    // Number of scramble characters (default 12)
	Label string // Optional label that appears after scramble
}

// Model is an animated thinking indicator with gradient scrambled characters.
type Model struct {
	id           int
	theme        *styles.Theme
	size         int
	label        string
	startTime    time.Time
	birthOffsets []time.Duration // random birth time for each character
	frame        int
	ellipsisStep int
	colorRamp    []lipgloss.Style
}

// Color ramp cache.
var (
	cache   = make(map[int][]lipgloss.Style)
	cacheMu sync.RWMutex
)

// New creates a new thinking indicator.
func New(theme *styles.Theme, settings Settings) *Model {
	size := settings.Size
	if size <= 0 {
		size = defaultSize
	}

	// Total width = scramble + space + label
	labelLen := len([]rune(settings.Label))
	totalWidth := size
	if labelLen > 0 {
		totalWidth += 1 + labelLen // space + label
	}

	// Random birth offset for each character position
	births := make([]time.Duration, totalWidth)
	for i := range births {
		births[i] = time.Duration(rand.Int64N(int64(birthDuration)))
	}

	m := &Model{
		id:           int(nextID.Add(1)),
		theme:        theme,
		size:         size,
		label:        settings.Label,
		startTime:    time.Now(),
		birthOffsets: births,
	}
	m.colorRamp = m.getOrBuildColorRamp()
	return m
}

func (m *Model) getOrBuildColorRamp() []lipgloss.Style {
	cacheMu.RLock()
	ramp, ok := cache[m.size]
	cacheMu.RUnlock()
	if ok {
		return ramp
	}

	colors := m.theme.Colors
	blended := styles.BlendColors(
		m.size*2,
		colors.Brand.GradientStart,
		colors.Brand.GradientEnd,
		colors.Brand.GradientStart,
	)
	ramp = make([]lipgloss.Style, len(blended))
	for i, c := range blended {
		ramp[i] = lipgloss.NewStyle().Foreground(c)
	}

	cacheMu.Lock()
	cache[m.size] = ramp
	cacheMu.Unlock()
	return ramp
}

// Init starts the animation.
func (m *Model) Init() tea.Cmd {
	return m.tick()
}

// Update handles tick messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if tick, ok := msg.(TickMsg); ok && tick.id == m.id {
		m.frame = (m.frame + 1) % len(m.colorRamp)
		m.ellipsisStep = (m.ellipsisStep + 1) % (ellipsisSpeed * len(ellipsis))
		return m.tick()
	}
	return nil
}

// View renders the thinking indicator.
func (m *Model) View() string {
	var b strings.Builder
	elapsed := time.Since(m.startTime)
	labelRunes := []rune(m.label)
	labelStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Page.TextMuted)
	allBorn := elapsed >= birthDuration

	totalWidth := m.size
	if len(labelRunes) > 0 {
		totalWidth += 1 + len(labelRunes) // space + label
	}

	for i := range totalWidth {
		// Not yet born - show nothing
		if !allBorn && elapsed < m.birthOffsets[i] {
			b.WriteRune(' ')
			continue
		}

		if i < m.size {
			// Scramble region
			colorIdx := (i + m.frame) % len(m.colorRamp)
			char := scrambleChars[rand.IntN(len(scrambleChars))]
			b.WriteString(m.colorRamp[colorIdx].Render(string(char)))
		} else if i == m.size {
			// Space between scramble and label
			b.WriteRune(' ')
		} else {
			// Label region
			labelIdx := i - m.size - 1
			b.WriteString(labelStyle.Render(string(labelRunes[labelIdx])))
		}
	}

	// Ellipsis after label (only when all characters born)
	if allBorn && m.label != "" {
		ellipsisIdx := m.ellipsisStep / ellipsisSpeed
		b.WriteString(labelStyle.Render(ellipsis[ellipsisIdx]))
	}

	return b.String()
}

// SetLabel updates the label and resets birth offsets.
func (m *Model) SetLabel(label string) {
	if m.label == label {
		return
	}
	m.label = label
	m.startTime = time.Now()

	labelLen := len([]rune(label))
	totalWidth := m.size
	if labelLen > 0 {
		totalWidth += 1 + labelLen
	}
	m.birthOffsets = make([]time.Duration, totalWidth)
	for i := range m.birthOffsets {
		m.birthOffsets[i] = time.Duration(rand.Int64N(int64(birthDuration)))
	}
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

	labelLen := len([]rune(m.label))
	totalWidth := size
	if labelLen > 0 {
		totalWidth += 1 + labelLen
	}
	m.birthOffsets = make([]time.Duration, totalWidth)
	for i := range m.birthOffsets {
		m.birthOffsets[i] = time.Duration(rand.Int64N(int64(birthDuration)))
	}
}

func (m *Model) tick() tea.Cmd {
	id := m.id
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg {
		return TickMsg{id: id}
	})
}
