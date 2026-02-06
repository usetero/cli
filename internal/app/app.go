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
	"github.com/usetero/cli/internal/app/chat/msgs"
	"github.com/usetero/cli/internal/app/keybar"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding"
	onboardingmsg "github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/app/statusbar"
	"github.com/usetero/cli/internal/app/toast"
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

// Layout constants.
const (
	horizontalPadding = 1
	verticalPadding   = 1
	gapAfterStatusBar = 1

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
	statusBar  *statusbar.Model
	toast      *toast.Model
	keyBar     *keybar.Model
	onboarding *onboarding.Model
	chat       *chat.Model
	state      state

	// Dimensions
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
		statusBar:   statusbar.New(theme, syncer, cfg.APIEndpoint),
		toast:       toast.New(theme),
		keyBar:      keybar.New(theme, scope),
		onboarding:  onboarding.New(ctx, theme, services, prefs, authService, syncer, scope),
		state:       stateOnboarding,
	}
}

// Init initializes the app.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.statusBar.Init(),
		m.onboarding.Init(),
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Quit) || key.Matches(msg, keymap.Exit) {
			m.shutdown()
			return m, tea.Quit
		}

	case onboardingmsg.AccountSelected:
		// Open database and start syncer when account is selected
		m.scope.Info("account selected", "account_id", msg.Account.ID.String())

		if err := m.openDatabase(msg.Account.ID.String()); err != nil {
			m.scope.Error("failed to open database", "error", err)
			return m, appmsg.ErrorCmd("Failed to open database", err, true)
		}

		if err := m.startSync(msg.Account.ID.String()); err != nil {
			m.scope.Error("failed to start sync", "error", err)
			return m, appmsg.ErrorCmd("Failed to start sync", err, true)
		}

		// Start catalog status polling now that db is ready
		catalogCmd := m.statusBar.SetDB(m.db)

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
			return m, tea.Batch(catalogCmd, cmd)
		}
		return m, catalogCmd

	case onboardingmsg.OnboardingComplete:
		m.state = stateChat
		m.scope.Info("onboarding complete",
			"org", msg.Org.Name,
			"account", msg.Account.Name,
			"workspace", msg.Workspace.Name,
		)

		// Create chat model (sizing happens via updateLayout)
		m.chat = chat.New(
			msg.Account,
			msg.Workspace,
			m.theme,
			m.db,
			m.chatClient,
			m.toolRegistry,
			m.scope,
		)

		// Size the new chat component
		m.updateLayout()

		return m, m.chat.Init()

	case msgs.StreamCompleted:
		if msg.Title != "" && m.db != nil {
			m.statusBar.SetTitle(msg.Title)
			// Persist title in background
			go func() {
				ctx := context.Background()
				if err := m.db.Conversations().UpdateTitle(ctx, string(m.chat.ConversationID()), msg.Title); err != nil {
					m.scope.Error("failed to update conversation title", "error", err)
				}
			}()
		}
		// Update context window usage in statusbar
		if msg.InputTokens > 0 && msg.ContextWindow > 0 {
			pct := (msg.InputTokens*100 + msg.ContextWindow - 1) / msg.ContextWindow // round up
			m.statusBar.SetContextPercent(pct)
		}
	}

	// Forward to app-level chrome and current page
	var cmds []tea.Cmd

	// Status bar listens to all messages
	if cmd := m.statusBar.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Toast listens to all messages
	if cmd := m.toast.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			cmds = append(cmds, m.onboarding.Update(msg))
		}
	case stateChat:
		if m.chat != nil {
			cmds = append(cmds, m.chat.Update(msg))
		}
	}

	// Update keybar after page updates (bindings may have changed)
	m.updateKeyBar()

	return m, tea.Batch(cmds...)
}

// updateLayout propagates sizes to children based on current dimensions.
func (m *Model) updateLayout() {
	contentWidth, contentHeight := m.contentSize()

	// Fixed components get width, report their height
	m.statusBar.SetWidth(contentWidth)
	m.toast.SetWidth(contentWidth)
	m.keyBar.SetWidth(contentWidth)

	statusBarHeight := m.statusBar.Height()
	toastHeight := m.toast.Height()
	keyBarHeight := m.keyBar.Height()

	// Page is flexible - gets remaining height
	// Toast always reserves its line to prevent layout shifts
	pageHeight := contentHeight - statusBarHeight - gapAfterStatusBar - toastHeight - keyBarHeight

	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			m.onboarding.SetSize(contentWidth, pageHeight)
		}
	case stateChat:
		if m.chat != nil {
			m.chat.SetSize(contentWidth, pageHeight)
		}
	}
}

// contentSize returns the available space for content (after app padding).
func (m *Model) contentSize() (int, int) {
	if m.width == 0 || m.height == 0 {
		return 0, 0
	}
	contentWidth := m.width - (horizontalPadding * 2)
	contentHeight := m.height - (verticalPadding * 2)
	return contentWidth, contentHeight
}

// updateKeyBar updates the keybar with current page bindings plus global bindings.
func (m *Model) updateKeyBar() {
	var bindings []key.Binding

	switch m.state {
	case stateOnboarding:
		if m.onboarding != nil {
			bindings = m.onboarding.ShortHelp()
		}
	case stateChat:
		if m.chat != nil {
			bindings = m.chat.ShortHelp()
		}
	}

	// Always append global bindings
	bindings = append(bindings, keymap.Global...)
	m.keyBar.SetKeyBindings(bindings)
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

	// Render content
	rendered := m.renderContent()

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

// renderContent renders the main content with padding.
// Layout: statusbar | page | toast (optional) | keybar
func (m *Model) renderContent() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	contentWidth, contentHeight := m.contentSize()

	// Get fixed component views and heights
	statusBarView := m.statusBar.View()
	statusBarHeight := m.statusBar.Height()
	toastView := m.toast.View()
	toastHeight := m.toast.Height()
	keyBarView := m.keyBar.View()
	keyBarHeight := m.keyBar.Height()

	// Get page content from current state
	var pageView string
	switch m.state {
	case stateOnboarding:
		pageView = m.onboarding.View()
	case stateChat:
		if m.chat != nil {
			pageView = m.chat.View()
		}
	}

	// Calculate page height (flexible component gets remaining space)
	// Toast always reserves its line to prevent layout shifts
	pageHeight := contentHeight - statusBarHeight - gapAfterStatusBar - toastHeight - keyBarHeight

	// Style page to fill its allocated space
	styledPage := lipgloss.NewStyle().
		Width(contentWidth).
		Height(pageHeight).
		MaxHeight(pageHeight).
		Render(pageView)

	// Build vertical stack
	var sections []string
	sections = append(sections, statusBarView)
	sections = append(sections, "") // gapAfterStatusBar
	sections = append(sections, styledPage)
	sections = append(sections, toastView)
	if keyBarHeight > 0 {
		sections = append(sections, keyBarView)
	}

	innerView := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply app padding
	return lipgloss.NewStyle().
		Padding(verticalPadding, horizontalPadding).
		Render(innerView)
}

func (m *Model) shutdown() {
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
}
