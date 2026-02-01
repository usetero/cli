package tui

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	tuiapp "github.com/usetero/cli/internal/tui/app"
	"github.com/usetero/cli/internal/tui/cursor"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/mode"
	"github.com/usetero/cli/internal/tui/onboarding"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/upload"
	"github.com/usetero/cli/pkg/client"
)

const (
	minWidth  = 80
	minHeight = 24
)

var (
	DisableMinSizeCheck = false
)

// databaseOpenedMsg is sent when the database is opened.
type databaseOpenedMsg struct {
	db sqlite.Database
}

// firstSyncCompleteMsg is sent when PowerSync completes its first sync.
type firstSyncCompleteMsg struct {
	sync *powersync.Sync
}

// uploadEventMsg wraps an upload event for the bubbletea message loop.
type uploadEventMsg struct {
	event upload.Event
}

// uploadDoneMsg is sent when the uploader goroutine exits.
type uploadDoneMsg struct{}

// TUI is the top-level model that routes between modes (onboarding, app).
type TUI struct {
	// Config - set at construction
	logger          log.Logger
	theme           *styles.Theme
	auth            auth.Auth
	powersyncConfig *powersync.Config
	apiEndpoint     string
	chatEndpoint    string

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// State - set after onboarding
	org           api.Organization
	account       api.Account
	db            sqlite.Database
	sync          powersync.Syncer
	conversations api.Conversations
	messages      chat.Messages
	uploader      *upload.Uploader
	uploaderDone  chan struct{}
	uploadStatus  upload.Status

	// UI
	currentMode     mode.Mode
	width           int
	height          int
	sendProgressBar bool
}

// New creates a new TUI model.
func New(cfg *config.Config, tokenStore auth.SecureStorage, oauthProvider auth.OAuthProvider, apiEndpoint, chatEndpoint string, powersyncConfig *powersync.Config, logger log.Logger) tea.Model {
	ctx, cancel := context.WithCancel(context.Background())
	theme := styles.NewTheme(true)

	authService := auth.NewService(oauthProvider, tokenStore, logger)
	preferencesService := preferences.NewService(cfg, logger)

	onboardingMode := onboarding.New(ctx, theme, logger, authService, preferencesService, apiEndpoint, keymap.Global)

	return &TUI{
		logger:          logger,
		theme:           theme,
		ctx:             ctx,
		cancel:          cancel,
		auth:            authService,
		powersyncConfig: powersyncConfig,
		apiEndpoint:     apiEndpoint,
		chatEndpoint:    chatEndpoint,
		currentMode:     onboardingMode,
	}
}

func (m *TUI) Init() tea.Cmd {
	return tea.Batch(
		setWindowTitle("Tero"),
		tea.RequestTerminalVersion,
		m.currentMode.Init(),
	)
}

func setWindowTitle(title string) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\033]0;%s\007", title)
		return nil
	}
}

func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global concerns
	if cmd, handled := m.handleGlobal(msg); handled {
		return m, cmd
	}

	// Handle lifecycle messages
	switch msg := msg.(type) {
	case account.AccountSelectedMsg:
		m.org = msg.Organization
		m.account = msg.Account

		token, _ := m.auth.GetAccessToken(m.ctx)

		apiClient := client.New(m.apiEndpoint, token, func() (string, error) {
			return m.auth.GetAccessToken(m.ctx)
		})
		apiClient.SetAccountID(m.account.ID)
		m.conversations = api.NewConversationService(apiClient, m.logger)

		chatClient := chat.NewClient(m.chatEndpoint, m.logger)
		chatClient.SetToken(token)
		chatClient.SetAccountID(m.account.ID)
		m.messages = chat.NewMessageService(chatClient)

		cmds = append(cmds, m.openDatabase())

	case databaseOpenedMsg:
		m.db = msg.db
		cmds = append(cmds, m.startSync())

	case firstSyncCompleteMsg:
		m.sync = msg.sync
		cmds = append(cmds, m.startUploader())
		cmds = append(cmds, m.transitionToApp())

	case uploadEventMsg:
		m.uploadStatus = msg.event.Status
		// TODO: surface status in UI
		cmds = append(cmds, m.listenUploadEvents())

	case uploadDoneMsg:
		m.uploader = nil
	}

	// Delegate to current mode
	if cmd := m.currentMode.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Check for mode transitions (onboarding complete)
	if cmd := m.checkModeTransition(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *TUI) handleGlobal(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keymap.Quit) || key.Matches(msg, keymap.Exit) {
			m.logger.Info("user quit", "key", msg.String())
			m.shutdown()
			return tea.Quit, true
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.currentMode.SetSize(msg.Width, msg.Height)
		return nil, true

	case tea.EnvMsg:
		if !m.sendProgressBar {
			m.sendProgressBar = slices.Contains(msg, "WT_SESSION")
		}

	case tea.TerminalVersionMsg:
		if !m.sendProgressBar {
			m.sendProgressBar = strings.Contains(strings.ToLower(msg.Name), "ghostty")
		}
	}

	return nil, false
}

func (m *TUI) checkModeTransition() tea.Cmd {
	if !m.currentMode.IsComplete() {
		return nil
	}

	// Onboarding complete -> wait for sync (show loading)
	if ob, ok := m.currentMode.(*onboarding.Onboarding); ok {
		m.org = ob.Organization()
		m.account = ob.Account()
		m.logger.Info("onboarding completed", "accountID", m.account.ID)
		// Mode stays as onboarding but shows "loading" state
		// Sync was already started when AccountSelectedMsg was received
	}

	return nil
}

// openDatabase opens the SQLite database for the account.
func (m *TUI) openDatabase() tea.Cmd {
	return func() tea.Msg {
		dbPath, err := m.powersyncConfig.DatabasePath(m.account.ID)
		if err != nil {
			m.logger.Error("failed to get database path", "error", err)
			return nil
		}

		db, err := sqlite.Open(m.ctx, dbPath)
		if err != nil {
			m.logger.Error("failed to open database", "error", err)
			return nil
		}

		m.logger.Info("database opened", "path", dbPath)
		return databaseOpenedMsg{db: db}
	}
}

// startSync starts PowerSync. Blocks until first sync completes.
func (m *TUI) startSync() tea.Cmd {
	return func() tea.Msg {
		done := make(chan struct{})

		sync := powersync.NewSync(m.powersyncConfig, m.auth, m.logger)
		err := sync.Start(m.ctx, m.db, m.account.ID, func() {
			close(done)
		})
		if err != nil {
			m.logger.Error("failed to start sync", "error", err)
			return nil
		}

		m.logger.Info("sync started, waiting for first sync", "accountID", m.account.ID)
		<-done
		m.logger.Info("first sync complete", "accountID", m.account.ID)

		return firstSyncCompleteMsg{sync: sync}
	}
}

// startUploader starts the upload loop and returns a command to listen for events.
func (m *TUI) startUploader() tea.Cmd {
	m.uploader = upload.New(m.db, m.conversations, m.messages, m.logger)
	m.uploaderDone = make(chan struct{})
	go func() {
		defer close(m.uploaderDone)
		if err := m.uploader.Run(m.ctx); err != nil && !errors.Is(err, context.Canceled) {
			m.logger.Error("upload loop error", "error", err)
		}
	}()
	return m.listenUploadEvents()
}

// listenUploadEvents returns a command that waits for the next upload event.
func (m *TUI) listenUploadEvents() tea.Cmd {
	if m.uploader == nil {
		return nil
	}
	events := m.uploader.Events()
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return uploadDoneMsg{}
		}
		return uploadEventMsg{event: event}
	}
}

// transitionToApp creates the app mode.
func (m *TUI) transitionToApp() tea.Cmd {
	m.logger.Info("transitioning to app", "accountID", m.account.ID)

	m.currentMode = tuiapp.New(m.ctx, m.theme, m.db, m.org, m.account, m.logger, keymap.Global)

	if m.width > 0 && m.height > 0 {
		m.currentMode.SetSize(m.width, m.height)
	}

	return m.currentMode.Init()
}

func (m *TUI) shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.uploaderDone != nil {
		<-m.uploaderDone
	}
	if m.sync != nil {
		m.sync.Stop()
	}
	if m.db != nil {
		m.db.Close()
	}
}

func (m *TUI) isBusy() bool {
	return m.currentMode.IsBusy()
}

func (m *TUI) View() tea.View {
	colors := m.theme.Colors

	if !DisableMinSizeCheck && (m.width < minWidth || m.height < minHeight) {
		view := tea.View{
			BackgroundColor: colors.Page.Bg,
			AltScreen:       true,
		}
		view.Content = lipgloss.NewStyle().
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
			)
		return view
	}

	modeView := m.currentMode.View()

	cleanView, cur := cursor.Extract(modeView)
	if cur != nil {
		cur.Color = colors.Accent
	}

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(cleanView),
	}

	comp := lipgloss.NewCompositor(layers...)

	view := tea.View{
		BackgroundColor: colors.Page.Bg,
		AltScreen:       true,
	}
	view.Content = comp.Render()
	view.Cursor = cur
	view.MouseMode = tea.MouseModeCellMotion

	if m.sendProgressBar && m.isBusy() {
		view.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, rand.IntN(100))
	}

	return view
}
