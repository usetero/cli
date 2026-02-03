// Package tui2 provides the main TUI application with proper Elm architecture.
// All models use value receivers and return (Model, tea.Cmd).
package tui

import (
	"context"
	"errors"
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	psapi "github.com/usetero/cli/internal/powersync/api"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/app"
	"github.com/usetero/cli/internal/tui/cursor"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/onboarding"
	accountselect "github.com/usetero/cli/internal/tui/onboarding/account/select"
	"github.com/usetero/cli/internal/upload"
)

const (
	minWidth  = 80
	minHeight = 24
)

// Model is the root tea.Model for the TUI.
type Model struct {
	// Core
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.CLIConfig
	logger log.Logger
	theme  *styles.Theme

	// Services
	storage     sqlite.Storage
	authService auth.Auth
	syncer      powersync.Syncer
	chatClient  chat.Client

	// Runtime (created after account selection)
	db       sqlite.DB
	uploader upload.Uploader

	// Components
	onboarding onboarding.Model
	app        app.Model

	// State
	org       domain.Organization
	account   domain.Account
	workspace domain.Workspace
	inApp     bool
	err       error

	// Layout
	width  int
	height int
}

// New creates a new TUI model.
func New(cfg *config.CLIConfig, authService auth.Auth, prefs preferences.Preferences, storage sqlite.Storage, syncer powersync.Syncer, chatClient chat.Client, logger log.Logger) Model {
	if cfg == nil {
		panic("tui.New: cfg is nil")
	}
	if authService == nil {
		panic("tui.New: authService is nil")
	}
	if prefs == nil {
		panic("tui.New: prefs is nil")
	}
	if storage == nil {
		panic("tui.New: storage is nil")
	}
	if syncer == nil {
		panic("tui.New: syncer is nil")
	}
	if chatClient == nil {
		panic("tui.New: chatClient is nil")
	}
	if logger == nil {
		panic("tui.New: logger is nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	theme := styles.NewTheme(true)

	return Model{
		ctx:         ctx,
		cancel:      cancel,
		cfg:         cfg,
		logger:      logger,
		theme:       theme,
		storage:     storage,
		authService: authService,
		syncer:      syncer,
		chatClient:  chatClient,
		onboarding:  onboarding.New(ctx, theme, authService, prefs, cfg.APIEndpoint, syncer, logger),
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		setWindowTitle("Tero"),
		m.onboarding.Init(),
	)
}

// Update handles messages and returns the updated Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Quit) || key.Matches(msg, keymap.Exit) {
			m.logger.Info("user quit", "key", msg.String())
			m.shutdown()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.inApp {
			m.app, _ = m.app.Update(msg)
		} else {
			m.onboarding = m.onboarding.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case accountselect.AccountSelectedMsg:
		m.org = msg.Organization
		m.account = msg.Account
		m.logger.Info("account selected", "accountID", msg.Account.ID)

		if err := m.openDatabase(msg.Account.ID.String()); err != nil {
			m.logger.Error("failed to open database", "error", err)
			m.err = err
			return m, nil
		}

		if err := m.startSync(msg.Account.ID.String()); err != nil {
			m.logger.Error("failed to start sync", "error", err)
			m.err = err
			return m, nil
		}
	}

	// Route to current mode
	if m.inApp {
		var cmd tea.Cmd
		m.app, cmd = m.app.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		var cmd tea.Cmd
		m.onboarding, cmd = m.onboarding.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check if onboarding completed
		if m.onboarding.IsComplete() {
			m.logger.Info("onboarding complete")
			m = m.transitionToApp()
			cmds = append(cmds, m.app.Init())
		}
	}

	return m, tea.Batch(cmds...)
}

// transitionToApp creates the app and switches to app mode.
func (m Model) transitionToApp() Model {
	m.org = m.onboarding.Organization()
	m.account = m.onboarding.Account()
	m.workspace = m.onboarding.Workspace()

	m.app = app.New(m.ctx, m.theme, m.db, m.chatClient, m.account, m.workspace, m.logger)
	m.app, _ = m.app.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	m.inApp = true

	m.logger.Info("transitioned to app", "workspace", m.workspace.ID)
	return m
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
	m.logger.Info("database opened", "path", dbPath)
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
	m.logger.Info("syncer started", "accountID", accountID)

	// Set account ID on chat client
	m.chatClient.SetAccountID(accountID)

	// Create API services with account scope
	services := api.NewServices(m.cfg.APIEndpoint+"/graphql", m.authService, m.logger)
	services.SetAccountID(domain.AccountID(accountID))

	// Create PowerSync API client for write checkpoints
	psClient := psapi.NewClient(m.cfg.PowerSyncEndpoint)

	// Create and start uploader
	m.uploader = upload.New(m.db, psClient, m.authService, services.Conversations, services.Messages, m.logger)
	go func() {
		if err := m.uploader.Run(m.ctx); err != nil && !errors.Is(err, context.Canceled) {
			m.logger.Error("uploader error", "error", err)
		}
	}()
	m.logger.Info("uploader started", "accountID", accountID)

	return nil
}

// View renders the TUI.
func (m Model) View() tea.View {
	colors := m.theme.Colors

	// Check minimum size
	if m.width < minWidth || m.height < minHeight {
		return tea.View{
			BackgroundColor: colors.Page.Bg,
			AltScreen:       true,
			Content: lipgloss.NewStyle().
				Width(m.width).
				Height(m.height).
				Align(lipgloss.Center, lipgloss.Center).
				Render(
					lipgloss.NewStyle().
						Padding(1, 4).
						Foreground(colors.Page.Text).
						BorderStyle(lipgloss.RoundedBorder()).
						BorderForeground(colors.Accent).
						Render("Window too small!"),
				),
		}
	}

	// Render current mode
	var content string
	if m.inApp {
		content = m.app.View()
	} else {
		content = m.onboarding.View()
	}

	cleanView, cur := cursor.Extract(content)
	if cur != nil {
		cur.Color = colors.Accent
	}

	return tea.View{
		BackgroundColor: colors.Page.Bg,
		AltScreen:       true,
		Content:         cleanView,
		Cursor:          cur,
	}
}

// shutdown cleans up resources.
func (m *Model) shutdown() {
	m.onboarding.Close()
	if m.syncer != nil {
		m.syncer.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
	if m.cancel != nil {
		m.cancel()
	}
}

// setWindowTitle returns a command that sets the terminal window title.
func setWindowTitle(title string) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\033]0;%s\007", title)
		return nil
	}
}
