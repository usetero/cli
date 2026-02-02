package tui

import (
	"context"
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
	"github.com/usetero/cli/internal/tui/database"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/mode"
	"github.com/usetero/cli/internal/tui/onboarding"
	"github.com/usetero/cli/internal/tui/onboarding/account"
	"github.com/usetero/cli/internal/tui/onboarding/workspace"
	"github.com/usetero/cli/pkg/client"
)

const (
	minWidth  = 80
	minHeight = 24
)

var (
	DisableMinSizeCheck = false
)

// TUI is the top-level model that routes between modes (onboarding, app).
type TUI struct {
	// Config - set at construction
	logger          log.Logger
	theme           *styles.Theme
	auth            auth.Auth
	storage         sqlite.Storage
	syncClient      powersync.Syncer
	powersyncClient powersync.Client
	apiEndpoint     string
	chatEndpoint    string

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// State - set after onboarding
	org       api.Organization
	account   api.Account
	workspace api.Workspace
	database  *database.Database

	// UI
	currentMode     mode.Mode
	width           int
	height          int
	sendProgressBar bool
}

// New creates a new TUI model.
func New(cfg *config.Config, tokenStore auth.SecureStorage, oauthProvider auth.OAuthProvider, apiEndpoint, chatEndpoint, powersyncEndpoint string, logger log.Logger) tea.Model {
	ctx, cancel := context.WithCancel(context.Background())
	theme := styles.NewTheme(true)

	authService := auth.NewService(oauthProvider, tokenStore, logger)
	preferencesService := preferences.NewService(cfg, logger)
	storageService := sqlite.NewStorageService(cfg)
	syncClient := powersync.NewSync(powersyncEndpoint, authService, logger)
	powersyncClient := powersync.NewClient(powersyncEndpoint)

	onboardingMode := onboarding.New(ctx, theme, logger, authService, preferencesService, apiEndpoint, keymap.Global)

	return &TUI{
		logger:          logger,
		theme:           theme,
		ctx:             ctx,
		cancel:          cancel,
		auth:            authService,
		storage:         storageService,
		syncClient:      syncClient,
		powersyncClient: powersyncClient,
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
	if msg, ok := msg.(account.AccountSelectedMsg); ok {
		m.org = msg.Organization
		m.account = msg.Account
	}

	if msg, ok := msg.(workspace.WorkspaceSelectedMsg); ok {
		m.org = msg.Organization
		m.account = msg.Account
		m.workspace = msg.Workspace

		token, _ := m.auth.GetAccessToken(m.ctx)

		apiClient := client.New(m.apiEndpoint, token, func() (string, error) {
			return m.auth.GetAccessToken(m.ctx)
		})
		apiClient.SetAccountID(m.account.ID)
		conversations := api.NewConversationService(apiClient, m.logger)

		chatClient := chat.NewClient(m.chatEndpoint, m.logger)
		chatClient.SetToken(token)
		chatClient.SetAccountID(m.account.ID)
		messages := chat.NewMessageService(chatClient)

		m.database = database.New(m.ctx, m.storage, m.syncClient, m.powersyncClient, m.auth, conversations, messages, m.logger)
		cmds = append(cmds, m.database.Start(m.account.ID))
	}

	// Delegate to database
	if m.database != nil {
		if cmd := m.database.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
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

	// Onboarding complete -> transition to app
	if ob, ok := m.currentMode.(*onboarding.Onboarding); ok {
		m.org = ob.Organization()
		m.account = ob.Account()
		m.workspace = ob.Workspace()
		m.logger.Info("onboarding completed", "accountID", m.account.ID, "workspaceID", m.workspace.ID)
		return m.transitionToApp()
	}

	return nil
}

// transitionToApp creates the app mode.
func (m *TUI) transitionToApp() tea.Cmd {
	m.logger.Info("transitioning to app", "accountID", m.account.ID)

	// Close the previous mode (onboarding)
	if m.currentMode != nil {
		if err := m.currentMode.Close(); err != nil {
			m.logger.Error("failed to close previous mode", "error", err)
		}
	}

	m.currentMode = tuiapp.New(m.ctx, m.theme, m.database.DB(), m.org, m.account, m.workspace, m.logger, keymap.Global)

	if m.width > 0 && m.height > 0 {
		m.currentMode.SetSize(m.width, m.height)
	}

	return m.currentMode.Init()
}

func (m *TUI) shutdown() {
	if m.currentMode != nil {
		if err := m.currentMode.Close(); err != nil {
			m.logger.Error("failed to close mode", "error", err)
		}
	}
	if m.cancel != nil {
		m.cancel()
	}
	if m.database != nil {
		m.database.Close()
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
