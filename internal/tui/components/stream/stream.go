// Package stream provides a text streaming component with thinking animation.
// Follows Crush's pattern: thinking animation → text streams word-by-word → done.
package stream

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/components/thinker"
)

// TickMsg is sent to advance to the next word
type TickMsg struct {
	id int
}

// Component streams text word-by-word with a thinking animation.
type Component struct {
	id int // unique ID for this component

	// Content
	text  string   // Full text to display
	words []string // Split by words
	shown int      // Number of words currently shown

	// Layout
	width int // Width for text wrapping

	// State
	thinking bool // Show thinking animation vs text
	done     bool // All text shown

	// Animation
	thinker  *thinker.Component // Thinking animation
	tickRate time.Duration      // How fast to stream words
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

var lastID int

func nextID() int {
	lastID++
	return lastID
}

// New creates a new streaming text component.
func New(text string) *Component {
	// Split text by whitespace into words
	words := strings.Fields(text)

	return &Component{
		id:       nextID(),
		text:     text,
		words:    words,
		shown:    0,
		width:    0, // Will be set by parent via SetWidth
		thinking: true,
		done:     false,
		thinker:  thinker.New(),
		tickRate: 50 * time.Millisecond, // Default: 50ms per word
	}
}

// SetWidth sets the width for text wrapping.
func (c *Component) SetWidth(width int) {
	c.width = width
}

// Init initializes the component with a default thinking duration
func (c *Component) Init() tea.Cmd {
	return c.Start(1 * time.Second) // Default 1 second of thinking
}

// Start begins the streaming animation.
// Returns a Cmd that starts the thinking animation for the specified duration.
func (c *Component) Start(thinkingDuration time.Duration) tea.Cmd {
	return tea.Batch(
		c.thinker.Start(),
		tea.Tick(thinkingDuration, func(t time.Time) tea.Msg {
			// After thinking duration, transition to streaming
			return TickMsg{id: c.id}
		}),
	)
}

// Update handles messages to advance the streaming.
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.id != c.id {
			return nil
		}

		// First tick: transition from thinking to streaming
		if c.thinking {
			c.thinking = false
			c.shown = 0
			// Start streaming immediately
			return c.tick()
		}

		// Subsequent ticks: show next word
		if c.shown < len(c.words) {
			c.shown++
			if c.shown < len(c.words) {
				// More words to show
				return c.tick()
			}
			// Last word shown
			c.done = true
			return nil
		}

		return nil

	case thinker.TickMsg:
		// Forward animation ticks to the animation component
		if c.thinking {
			cmd := c.thinker.Update(msg)
			return cmd
		}
		return nil
	}

	return nil
}

// tick returns a Cmd that sends the next TickMsg after tickRate duration.
func (c *Component) tick() tea.Cmd {
	return tea.Tick(c.tickRate, func(t time.Time) tea.Msg {
		return TickMsg{id: c.id}
	})
}

// View renders the component.
func (c *Component) View() string {
	// Show thinking animation
	if c.thinking {
		return c.thinker.View()
	}

	// Show streamed text (partial or complete)
	if c.shown == 0 {
		return ""
	}

	// Join the words shown so far
	shownText := strings.Join(c.words[:c.shown], " ")

	// Style with text color and wrapping
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")) // White text for readability

	// Only set width if parent provided it (> 0)
	if c.width > 0 {
		style = style.Width(c.width)
	}

	return style.Render(shownText)
}

// IsDone returns true if all text has been shown.
func (c *Component) IsDone() bool {
	return c.done
}

// IsBusy returns false - stream components are never busy
func (c *Component) IsBusy() bool {
	return false
}

// HasError returns false - stream components have no error state
func (c *Component) HasError() bool {
	return false
}

// Error returns nil - stream components have no error state
func (c *Component) Error() error {
	return nil
}
