// Package thinker provides a thinking animation with random characters.
// Inspired by Crush's spinner animation.
package thinker

import (
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/components"
)

const (
	fps      = 20         // 20 FPS = 50ms per frame
	numChars = 10         // Number of random characters
	label    = "Thinking" // Label to show
	labelGap = " "        // Gap between label and animation
)

var (
	// Random characters for the animation
	availableRunes = []rune("0123456789abcdefABCDEF~!@#$£€%^&*()+=_")

	// Ellipsis frames
	ellipsisFrames = []string{"", ".", "..", "..."}
)

// TickMsg is sent to advance the animation frame.
type TickMsg struct {
	id int
}

var lastID int

func nextID() int {
	lastID++
	return lastID
}

// Component is an animated thinking spinner.
type Component struct {
	theme         *styles.Theme
	id            int
	frame         int
	ellipsisFrame int
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new animation component.
func New(theme *styles.Theme) *Component {
	return &Component{
		theme:         theme,
		id:            nextID(),
		frame:         0,
		ellipsisFrame: 0,
	}
}

// Init initializes the component with default animation
func (a *Component) Init() tea.Cmd {
	return a.Start()
}

// Start begins the animation loop.
func (a *Component) Start() tea.Cmd {
	return a.tick()
}

// Update handles tick messages to advance the animation.
func (a *Component) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.id != a.id {
			return nil
		}

		a.frame++

		// Advance ellipsis every 8 frames (400ms at 20 FPS)
		if a.frame%8 == 0 {
			a.ellipsisFrame = (a.ellipsisFrame + 1) % len(ellipsisFrames)
		}

		return a.tick()
	}

	return nil
}

// tick returns a Cmd that sends the next tick after 50ms.
func (a *Component) tick() tea.Cmd {
	return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
		return TickMsg{id: a.id}
	})
}

// View renders the animation.
func (a *Component) View() string {
	colors := a.theme.Colors

	// Generate random characters for this frame
	var chars strings.Builder
	for i := 0; i < numChars; i++ {
		chars.WriteRune(availableRunes[rand.Intn(len(availableRunes))])
	}

	// Style the components
	labelStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted)

	charsStyle := lipgloss.NewStyle().
		Foreground(colors.Accent)

	ellipsisStyle := lipgloss.NewStyle().
		Foreground(colors.Page.TextMuted)

	// Compose: "Thinking [random chars]..."
	return labelStyle.Render(label) +
		labelGap +
		charsStyle.Render(chars.String()) +
		ellipsisStyle.Render(ellipsisFrames[a.ellipsisFrame])
}

// IsBusy returns false - thinker components are never busy
func (a *Component) IsBusy() bool {
	return false
}

// HasError returns false - thinker components have no error state
func (a *Component) HasError() bool {
	return false
}

// Error returns nil - thinker components have no error state
func (a *Component) Error() error {
	return nil
}
