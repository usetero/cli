// Package app provides the main TUI application.
package app

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/api"
	chatclient "github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/app/chat"
	"github.com/usetero/cli/internal/app/chat/usecase"
	chattools "github.com/usetero/cli/internal/app/chattools"
	"github.com/usetero/cli/internal/app/keybar"
	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/app/onboarding"
	"github.com/usetero/cli/internal/app/palette"
	"github.com/usetero/cli/internal/app/statusbar"
	"github.com/usetero/cli/internal/app/toast"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/update"
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
	gapAfterStatusBar = 2
	gapBeforeKeyBar   = 1

	minWidth  = 50
	minHeight = 25
)

// Model is the root application model.
type Model struct {
	ctx     context.Context
	theme   styles.Theme
	scope   log.Scope
	version string

	// Dependencies
	cfg         *config.CLIConfig
	storage     sqlite.Storage
	authService auth.Auth
	syncer      powersync.Syncer
	services    api.APIServices
	userPrefs   preferences.UserPreferences
	orgPrefs    preferences.OrgPreferences

	// Runtime (created after account selection / onboarding)
	sessionCancel context.CancelFunc
	db            sqlite.DB
	uploader      upload.Uploader
	chatClient    chatclient.Client
	runtimeDeps   usecase.RuntimeDeps
	toolRegistry  *chattools.Registry
	user          *auth.User
	account       domain.Account
	workspace     domain.Workspace

	// Components
	statusBar  *statusbar.Model
	toast      *toast.Model
	keyBar     *keybar.Model
	onboarding *onboarding.Model
	chat       *chat.Model
	quitDlg    *quitDialog
	palette    *palette.Model
	state      state

	// Dimensions
	width  int
	height int

	// Terminal window title (set from conversation title)
	windowTitle string
}

// New creates a new app model.
func New(
	ctx context.Context,
	cfg *config.CLIConfig,
	theme styles.Theme,
	version string,
	services api.APIServices,
	authService auth.Auth,
	userPrefs preferences.UserPreferences,
	orgPrefs preferences.OrgPreferences,
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
	if authService == nil {
		panic("authService is nil")
	}
	if userPrefs == nil {
		panic("userPrefs is nil")
	}
	if orgPrefs == nil {
		panic("orgPrefs is nil")
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
		version:     version,
		cfg:         cfg,
		storage:     storage,
		authService: authService,
		syncer:      syncer,
		services:    services,
		userPrefs:   userPrefs,
		orgPrefs:    orgPrefs,
		statusBar:   statusbar.New(theme, scope, syncer, cfg.APIEndpoint, cfg.Env),
		toast:       toast.New(theme),
		keyBar:      keybar.New(theme, scope),
		onboarding:  onboarding.New(ctx, theme, services, userPrefs, orgPrefs, authService, syncer, scope),
		state:       stateOnboarding,
	}
}

// Init initializes the app.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.statusBar.Init(),
		m.onboarding.Init(),
		m.checkForUpdate(),
	)
}

// checkForUpdate returns a command that checks GitHub for a newer release.
// Skips the check for dev builds. Errors are logged, never shown to the user.
func (m *Model) checkForUpdate() tea.Cmd {
	if m.version == "" || m.version == "dev" {
		return nil
	}
	version := m.version
	ctx := m.ctx
	scope := m.scope
	return func() tea.Msg {
		result, err := update.Check(ctx, version, "")
		if err != nil {
			scope.Debug("update check failed", "error", err)
			return nil
		}
		if result == nil {
			return nil
		}
		return appmsg.Info{Message: fmt.Sprintf("Tero CLI update available: %s → %s", result.Current, result.Latest)}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	start := time.Now()
	defer m.logSlowUpdate(start, msg)

	if cmd, handled := m.handleGlobalMessage(msg); handled {
		return m, cmd
	}

	if cmd, handled := m.handleInteractionMessage(msg); handled {
		return m, cmd
	}

	if cmd, handled := m.handleOnboardingMessage(msg); handled {
		return m, cmd
	}
	m.handleStreamCompleted(msg)
	return m, m.updateChildren(msg)
}

// newChat creates a fresh chat model with current dependencies.
func (m *Model) newChat() *chat.Model {
	return chat.New(
		m.user,
		m.account,
		m.workspace,
		m.theme,
		m.db,
		m.runtimeDeps,
		m.toolRegistry,
		m.scope,
	)
}

// openPalette creates and opens the command palette.
func (m *Model) openPalette() tea.Cmd {
	commands := m.paletteCommands()
	if len(commands) == 0 {
		return nil
	}

	contentWidth, _ := m.contentSize()
	paletteWidth := min(contentWidth, 50)

	m.palette = palette.New(m.theme, commands)
	m.palette.SetWidth(paletteWidth)
	return m.palette.Init()
}
