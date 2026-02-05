// Package app provides the main TUI application.
package app

import (
	"context"
	"errors"
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/chat"
	"github.com/usetero/cli/internal/app/onboarding"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/auth"
	chatclient "github.com/usetero/cli/internal/chat"
	chattools "github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	psapi "github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/footer"
	"github.com/usetero/cli/internal/tea/cursor"
	"github.com/usetero/cli/internal/tea/keymap"
	"github.com/usetero/cli/internal/upload"
)

// state represents the app state.
type state int

const (
	stateOnboarding state = iota
	stateChat
)

const (
	horizontalPadding = 1
	verticalPadding   = 1
	footerSpacing     = 1

	minWidth  = 50
	minHeight = 25
)

// Model is the root application model.
type Model struct {
	ctx   context.Context
	theme *styles.Theme
	scope log.Scope

	// Dependencies
	cfg         *config.CLIConfig
	storage     sqlite.Storage
	authService auth.Auth
	syncer      powersync.Syncer
	services    api.APIServices

	// Runtime (created after account selection)
	db           sqlite.DB
	uploader     upload.Uploader
	chatClient   chatclient.Client
	toolRegistry *chattools.Registry

	// Components
	footer     *footer.Model
	onboarding *onboarding.Model
	chat       *chat.Model
	state      state

	width  int
	height int
}

// New creates a new app model.
func New(
	ctx context.Context,
	cfg *config.CLIConfig,
	theme *styles.Theme,
	services api.APIServices,
	authService auth.Auth,
	prefs preferences.Preferences,
	storage sqlite.Storage,
	syncer powersync.Syncer,
	scope log.Scope,
) *Model {
	if ctx == nil {
		panic("ctx is nil")
	}
	if cfg == nil {
		panic("cfg is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if authService == nil {
		panic("authService is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if storage == nil {
		panic("storage is nil")
	}
	if syncer == nil {
		panic("syncer is nil")
	}

	scope = scope.Child("app")

	return &Model{
		ctx:         ctx,
		theme:       theme,
		scope:       scope,
		cfg:         cfg,
		storage:     storage,
		authService: authService,
		syncer:      syncer,
		services:    services,
		footer:      footer.New(theme),
		onboarding:  onboarding.New(ctx, theme, services, prefs, authService, syncer, scope),
		state:       stateOnboarding,
	}
}

// Init initializes the app.
func (m *Model) Init() tea.Cmd {
	return m.onboarding.Init()
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.footer.SetWidth(msg.Width - (horizontalPadding * 2))
		contentWidth, contentHeight := m.contentSize()
		if m.state == stateOnboarding && m.onboarding != nil {
			m.onboarding.SetSize(contentWidth, contentHeight)
		}
		if m.state == stateChat && m.chat != nil {
			m.chat.SetSize(contentWidth, contentHeight)
		}
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Quit) || key.Matches(msg, keymap.Exit) {
			m.shutdown()
			return m, tea.Quit
		}

	case msgs.AccountSelected:
		// Open database and start syncer when account is selected
		m.scope.Info("account selected", "account_id", msg.Account.ID.String())

		if err := m.openDatabase(msg.Account.ID.String()); err != nil {
			m.scope.Error("failed to open database", "error", err)
			// TODO: show error to user
			return m, nil
		}

		if err := m.startSync(msg.Account.ID.String()); err != nil {
			m.scope.Error("failed to start sync", "error", err)
			// TODO: show error to user
			return m, nil
		}

		// Create tool registry with executors
		m.toolRegistry = &chattools.Registry{
			Query:        chattools.NewQueryTool(m.db),
			StartJourney: chattools.NewStartJourneyTool(),
			EndJourney:   chattools.NewEndJourneyTool(),
		}

		// Create chat client with tool definitions
		m.chatClient = chatclient.NewClient(m.cfg.ChatEndpoint, m.authService, m.scope, m.toolRegistry.Definitions())
		m.chatClient.SetAccountID(msg.Account.ID)

		// Forward to onboarding so it can continue
		if m.onboarding != nil {
			cmd := m.onboarding.Update(msg)
			return m, cmd
		}
		return m, nil

	case msgs.OnboardingComplete:
		m.state = stateChat
		m.scope.Info("onboarding complete",
			"org", msg.Org.Name,
			"account", msg.Account.Name,
			"workspace", msg.Workspace.Name,
		)

		// Create chat model
		contentWidth, contentHeight := m.contentSize()
		m.chat = chat.New(
			msg.Account,
			msg.Workspace,
			contentWidth,
			contentHeight,
			m.theme,
			m.db,
			m.chatClient,
			m.toolRegistry,
			m.scope,
		)

		// Set keybindings from chat
		m.footer.SetKeyBindings(m.chat.KeyBindings())

		return m, m.chat.Init()
	}

	// Forward to current state
	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			cmd := m.onboarding.Update(msg)
			return m, cmd
		}
	case stateChat:
		if m.chat != nil {
			cmd := m.chat.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// openDatabase opens the SQLite database for the given account.
func (m *Model) openDatabase(accountID string) error {
	dbPath, err := m.storage.DatabasePath(accountID)
	if err != nil {
		return err
	}

	db, err := sqlite.Open(m.ctx, dbPath)
	if err != nil {
		return err
	}

	m.db = db
	m.scope.Info("database opened", "path", dbPath)
	return nil
}

// startSync starts the syncer and uploader with the open database.
func (m *Model) startSync(accountID string) error {
	if m.db == nil {
		return fmt.Errorf("database not open")
	}

	if err := m.syncer.Start(m.ctx, m.db, accountID, nil); err != nil {
		return err
	}
	m.scope.Info("syncer started", "account_id", accountID)

	// Set account ID on services
	m.services.SetAccountID(domain.AccountID(accountID))

	// Create PowerSync API client for write checkpoints
	psClient := psapi.NewClient(m.cfg.PowerSyncEndpoint)

	// Create and start uploader
	m.uploader = upload.New(m.db, psClient, m.authService, m.services.Conversations, m.services.Messages, m.scope)
	go func() {
		if err := m.uploader.Run(m.ctx); err != nil && !errors.Is(err, context.Canceled) {
			m.scope.Error("uploader error", "error", err)
		}
	}()
	m.scope.Info("uploader started", "account_id", accountID)

	return nil
}

// View renders the app.
func (m *Model) View() tea.View {
	colors := m.theme.Colors

	// Show message if window is too small
	if m.width < minWidth || m.height < minHeight {
		content := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(
				lipgloss.NewStyle().
					Padding(0, 2).
					Foreground(colors.Page.Text).
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(colors.Accent).
					Render("Window too small"),
			)
		return tea.View{
			Content:         content,
			BackgroundColor: colors.Page.Bg,
			AltScreen:       true,
		}
	}

	// Get content from current state
	var content string
	switch m.state {
	case stateOnboarding:
		content = m.onboarding.View()
	case stateChat:
		if m.chat != nil {
			content = m.chat.View()
		}
	}

	// Render with padding and footer
	rendered := m.renderWithChrome(content)

	// Extract cursor marker and set cursor position
	cleanView, cur := cursor.Extract(rendered)
	if cur != nil {
		cur.Color = colors.Accent
	}

	return tea.View{
		Content:         cleanView,
		BackgroundColor: colors.Page.Bg,
		AltScreen:       true,
		Cursor:          cur,
		MouseMode:       tea.MouseModeCellMotion,
	}
}

// renderWithChrome wraps content with padding and footer.
func (m *Model) renderWithChrome(content string) string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	innerWidth := m.width - (horizontalPadding * 2)
	footerView := m.footer.View()
	footerHeight := lipgloss.Height(footerView)
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing

	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(contentHeight)
	styledContent := contentStyle.Render(content)

	innerView := lipgloss.JoinVertical(
		lipgloss.Top,
		styledContent,
		"",
		footerView,
	)

	return lipgloss.NewStyle().
		Padding(verticalPadding, horizontalPadding).
		Render(innerView)
}

// contentSize returns the available space for content.
func (m *Model) contentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}
	contentWidth := m.width - (horizontalPadding * 2)
	footerHeight := m.footer.Height()
	contentHeight := m.height - (verticalPadding * 2) - footerHeight - footerSpacing
	return contentWidth, contentHeight
}

func (m *Model) shutdown() {
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
}
