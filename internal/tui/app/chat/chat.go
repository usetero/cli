package chat

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app/page"
)

// DatabaseReadyMsg is sent when the database is ready for queries.
type DatabaseReadyMsg struct {
	DB sqlite.Database
}

// model represents the chat page state
type model struct {
	// Identity - which org/account this chat session belongs to
	orgID     string
	accountID string

	// Data layer
	db sqlite.Database

	// Sync status display
	serviceCount int64
	syncError    error

	// Logger
	logger log.Logger

	// Dimensions (set by app)
	width  int
	height int
	ready  bool
}

// New creates a new chat page model.
func New(orgID string, accountID string, logger log.Logger) page.Page {
	return &model{
		orgID:     orgID,
		accountID: accountID,
		logger:    logger,
	}
}

// Init is called when the page is first loaded
func (m *model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m *model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case DatabaseReadyMsg:
		m.db = msg.DB
		m.logger.Info("chat received database")
		return m.queryServiceCount()

	case serviceCountMsg:
		if msg.err != nil {
			m.syncError = msg.err
			m.logger.Error("failed to query service count", "error", msg.err)
		} else {
			m.serviceCount = msg.count
			m.logger.Info("queried service count", "count", msg.count)
		}
		return nil
	}
	return nil
}

// queryServiceCount queries the number of services from the local database.
func (m *model) queryServiceCount() tea.Cmd {
	return func() tea.Msg {
		if m.db == nil {
			return serviceCountMsg{err: fmt.Errorf("database not ready")}
		}
		count, err := m.db.Count("services")
		return serviceCountMsg{count: count, err: err}
	}
}

// serviceCountMsg is the result of querying the service count.
type serviceCountMsg struct {
	count int64
	err   error
}

// View renders just the page content (no chrome)
func (m *model) View() string {
	if !m.ready {
		return ""
	}

	theme := styles.CurrentTheme()

	var status string
	if m.db == nil {
		status = "Connecting to PowerSync..."
	} else if m.syncError != nil {
		status = fmt.Sprintf("Sync error: %v", m.syncError)
	} else {
		status = fmt.Sprintf("Connected to PowerSync. Synced %d services.", m.serviceCount)
	}

	content := lipgloss.NewStyle().
		Foreground(theme.Page.Text).
		Render(
			lipgloss.Place(
				m.width,
				m.height,
				lipgloss.Center,
				lipgloss.Center,
				status,
			),
		)

	return content
}

// SetSize sets the dimensions available for content
func (m *model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ready = true
}

// Title returns the page title
func (m *model) Title() string {
	return "Chat"
}

// Metadata returns context to display in sidebar/header
func (m *model) Metadata() []page.Metadata {
	return []page.Metadata{
		{Label: "Organization", Value: m.orgID, Priority: 1},
		{Label: "Account", Value: m.accountID, Priority: 2},
	}
}

// AcceptsNaturalLanguage returns true - chat accepts free-form input
func (m *model) AcceptsNaturalLanguage() bool {
	return true
}

// Commands returns available slash commands
func (m *model) Commands() []page.Command {
	return []page.Command{
		{Name: "services", Description: "View and manage services"},
		{Name: "policies", Description: "View and manage policies"},
		{Name: "help", Description: "Show available commands"},
	}
}

// KeyBindings returns keyboard shortcuts for the footer
func (m *model) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear"),
		),
	}
}

// IsBusy returns true if chat is streaming a response
func (m *model) IsBusy() bool {
	return m.db == nil // Busy while waiting for sync
}

// HasError returns true if chat is in an error state
func (m *model) HasError() bool {
	return m.syncError != nil
}

// Error returns the current error
func (m *model) Error() error {
	return m.syncError
}
