package header

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/components"
	"github.com/usetero/cli/internal/tui/components/logo"
	"github.com/usetero/cli/internal/tui/styles"
)

const diag = `╱`

// Component is a full-width header component with logo
type Component struct {
	width  int
	logger log.Logger
}

// Compile-time check that Component implements components.Component
var _ components.Component = (*Component)(nil)

// New creates a new header component
func New(logger log.Logger) *Component {
	return &Component{
		logger: logger,
	}
}

// Init initializes the component
func (c *Component) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
	}
	return nil
}

// View renders the header with logo and diagonal fields on sides
func (c *Component) View() string {
	if c.width == 0 {
		return ""
	}

	theme := styles.CurrentTheme()

	// Render just the TERO wordmark
	logoView := logo.Render(logo.Opts{
		TitleColorA: theme.Primary,
		TitleColorB: theme.Secondary,
	})

	// Calculate dimensions
	fieldHeight := lipgloss.Height(logoView)
	logoWidth := lipgloss.Width(strings.Split(logoView, "\n")[0])

	// Left diagonal field (6 chars wide, like Crush)
	const leftWidth = 6
	fieldStyle := lipgloss.NewStyle().Foreground(theme.Field)
	leftFieldRow := fieldStyle.Render(strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		leftField.WriteString(leftFieldRow + "\n")
	}

	// Right diagonal field fills remaining space
	rightWidth := max(15, c.width-logoWidth-leftWidth-2) // 2 for gaps
	rightFieldRow := fieldStyle.Render(strings.Repeat(diag, rightWidth))
	rightField := new(strings.Builder)
	for range fieldHeight {
		rightField.WriteString(rightFieldRow + "\n")
	}

	// Join horizontally: left diagonals + gap + logo + gap + right diagonals
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.TrimSpace(leftField.String()),
		" ",
		logoView,
		" ",
		strings.TrimSpace(rightField.String()),
	)
}

// IsBusy returns false - header components are never busy
func (c *Component) IsBusy() bool {
	return false
}

// HasError returns false - header components don't have error states
func (c *Component) HasError() bool {
	return false
}

// Error returns nil - header components don't have errors
func (c *Component) Error() error {
	return nil
}
