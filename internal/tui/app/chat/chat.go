package chat

import (
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui/app/page"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/layouts"
	"github.com/usetero/cli/internal/tui/styles"
)

// model represents the chat page state
type model struct {
	// Identity - which org/account this chat session belongs to
	orgID     string
	accountID string

	// TODO: Add services when we know what chat needs

	// Logger
	logger log.Logger

	// Layout
	layout layouts.Layout
	ready  bool

	// Global key bindings (passed from TUI)
	globalBindings []key.Binding
}

// New creates a new chat page model.
// Takes the accumulated onboarding data (orgID, accountID) plus logger and services.
func New(orgID string, accountID string, logger log.Logger, globalBindings []key.Binding) page.Page {
	return &model{
		orgID:          orgID,
		accountID:      accountID,
		logger:         logger,
		ready:          false,
		layout:         layouts.NewSidebar(logger),
		globalBindings: globalBindings,
	}
}

// Init is called when the program starts
func (m *model) Init() tea.Cmd {
	return nil
}

// SetSize sets the width and height available for rendering
func (m *model) SetSize(width, height int) {
	m.layout.SetSize(width, height)
	m.ready = true
}

// Update handles incoming messages and updates state
func (m *model) Update(msg tea.Msg) tea.Cmd {
	// Note: WindowSizeMsg is handled by parent (tui.go), not here
	// Pages only handle their own specific messages

	// Combine page bindings + global bindings
	var bindings []key.Binding
	bindings = append(bindings, m.Help().ShortHelp()...)
	bindings = append(bindings, m.globalBindings...)
	m.layout.SetKeyBindings(bindings)

	// Pass error state to layout (always set, even if nil to clear previous errors)
	m.layout.SetError(m.Error())

	// Cascade to layout
	cmd := m.layout.Update(msg)

	switch msg.(type) {
	case tea.KeyMsg:
		// Handle key messages when needed
	}

	return cmd
}

// View renders the page content as a string (implements pages.Page interface)
func (m *model) View() string {
	if !m.ready {
		return ""
	}

	// Ask layout for available content dimensions
	contentWidth, contentHeight := m.layout.ContentSize()

	theme := styles.CurrentTheme()

	// Render chat content inline
	chatContent := lipgloss.NewStyle().
		Foreground(theme.Text).
		Render(
			lipgloss.Place(
				contentWidth,
				contentHeight,
				lipgloss.Center,
				lipgloss.Center,
				"Chat interface coming soon...",
			),
		)

	// Layout handles sidebar + content + footer composition
	return m.layout.Render(chatContent)
}

// IsBusy returns true if the chat is performing a background operation.
// Currently never busy - will be true when streaming messages in the future.
func (m *model) IsBusy() bool {
	return false
}

// HasError returns false - chat page has no error state currently
func (m *model) HasError() bool {
	return false
}

// Error returns nil - chat page has no error state currently
func (m *model) Error() error {
	return nil
}

// Help returns key bindings for the chat page
func (m *model) Help() help.KeyMap {
	// Chat page has no custom bindings yet - will add when we implement message sending
	return keymap.Simple{Keys: []key.Binding{}}
}
