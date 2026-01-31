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
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	tuiapp "github.com/usetero/cli/internal/tui/app"
	"github.com/usetero/cli/internal/tui/cursor"
	"github.com/usetero/cli/internal/tui/keymap"
	"github.com/usetero/cli/internal/tui/loading"
	"github.com/usetero/cli/internal/tui/mode"
	"github.com/usetero/cli/internal/tui/onboarding"
	"github.com/usetero/cli/internal/tui/sync"
)

const (
	// Minimum window dimensions
	minWidth  = 80
	minHeight = 24
)

var (
	// DisableMinSizeCheck disables the minimum window size check for testing
	DisableMinSizeCheck = false
)

// TUI is the top-level model that routes between modes (onboarding, app).
type TUI struct {
	logger log.Logger
	theme  *styles.Theme

	// Lifecycle context - cancelled on quit for clean shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// PowerSync - handles sync lifecycle, started when account is selected
	syncManager *sync.Manager

	// Current mode (onboarding or app)
	currentMode mode.Mode

	// Window dimensions
	width  int
	height int

	// sendProgressBar instructs the TUI to send progress bar updates to the
	// terminal. Only enabled for supported terminals (Windows Terminal, Ghostty).
	sendProgressBar bool
}

// New creates a new TUI model
func New(cfg *config.Config, tokenStore auth.SecureStorage, oauthProvider auth.OAuthProvider, apiEndpoint string, powersyncConfig *powersync.Config, logger log.Logger) tea.Model {
	// Create lifecycle context - cancelled on quit
	ctx, cancel := context.WithCancel(context.Background())

	// Create theme first - it's used by everything
	theme := styles.NewTheme(true) // dark mode

	// Create domain services
	authService := auth.NewService(oauthProvider, tokenStore, logger)
	preferencesService := preferences.NewService(cfg, logger)

	// Create sync manager (not started until account is selected)
	syncManager := sync.NewManager(ctx, powersyncConfig, authService, logger)

	// Start with onboarding mode
	onboardingMode := onboarding.New(ctx, theme, logger, authService, preferencesService, apiEndpoint, keymap.Global)

	return &TUI{
		logger:      logger,
		theme:       theme,
		ctx:         ctx,
		cancel:      cancel,
		syncManager: syncManager,
		currentMode: onboardingMode,
	}
}

// Init initializes the application
func (m *TUI) Init() tea.Cmd {
	return tea.Batch(
		setWindowTitle("Tero"),
		tea.RequestTerminalVersion,
		m.currentMode.Init(),
	)
}

// setWindowTitle returns a command that sets the terminal window title
func setWindowTitle(title string) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("\033]0;%s\007", title)
		return nil
	}
}

// Update handles messages
func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// 1. Handle global concerns (quit, window size, terminal detection)
	if cmd, handled := m.handleGlobal(msg); handled {
		return m, cmd
	}

	// 2. Delegate to sync manager
	if cmd := m.syncManager.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// 3. Delegate to current mode
	if cmd := m.currentMode.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// 4. Check for mode transitions
	if cmd := m.checkModeTransition(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleGlobal handles global concerns like quit, window size, and terminal detection.
// Returns (cmd, true) if the message was fully handled and should not be delegated further.
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
		m.logger.Info("window resized",
			"terminalWidth", msg.Width,
			"terminalHeight", msg.Height,
			"mode", fmt.Sprintf("%T", m.currentMode))
		m.currentMode.SetSize(msg.Width, msg.Height)
		return nil, true

	case tea.EnvMsg:
		if !m.sendProgressBar {
			m.sendProgressBar = slices.Contains(msg, "WT_SESSION")
			if m.sendProgressBar {
				m.logger.Info("enabled progress bar", "terminal", "Windows Terminal")
			}
		}
		// Don't return handled - let it propagate

	case tea.TerminalVersionMsg:
		termVersion := strings.ToLower(msg.Name)
		m.logger.Debug("received terminal version", "version", termVersion)
		if !m.sendProgressBar {
			m.sendProgressBar = strings.Contains(termVersion, "ghostty")
			if m.sendProgressBar {
				m.logger.Info("enabled progress bar", "terminal", "Ghostty")
			}
		}
		// Don't return handled - let it propagate
	}

	return nil, false
}

// checkModeTransition checks if the current mode completed and transitions to the next.
func (m *TUI) checkModeTransition() tea.Cmd {
	if !m.currentMode.IsComplete() {
		return nil
	}

	switch mode := m.currentMode.(type) {
	case *onboarding.Onboarding:
		org := mode.Organization()
		account := mode.Account()

		m.logger.Info("onboarding completed, transitioning to loading",
			"orgID", org.ID,
			"accountID", account.ID)

		// Transition to Loading mode with sync manager reference
		// Loading checks sync state directly rather than relying on messages
		m.currentMode = loading.New(m.theme, org, account, m.syncManager)

		if m.width > 0 && m.height > 0 {
			m.currentMode.SetSize(m.width, m.height)
		}

		return m.currentMode.Init()

	case *loading.Loading:
		db := mode.DB()
		org := mode.Organization()
		account := mode.Account()

		m.logger.Info("loading completed, transitioning to app",
			"orgID", org.ID,
			"accountID", account.ID)

		// Transition to App with database
		m.currentMode = tuiapp.New(m.ctx, m.theme, db, org, account, m.logger, keymap.Global)

		if m.width > 0 && m.height > 0 {
			m.currentMode.SetSize(m.width, m.height)
		}

		return m.currentMode.Init()
	}

	return nil
}

// shutdown cleans up resources on quit
func (m *TUI) shutdown() {
	// Cancel context to stop all background operations
	if m.cancel != nil {
		m.cancel()
	}

	// Stop sync and close database
	if m.syncManager != nil {
		m.syncManager.Shutdown()
	}
}

// isBusy returns true if the TUI is currently performing a background operation
// and should show the progress bar animation
func (m *TUI) isBusy() bool {
	return m.currentMode.IsBusy()
}

// View renders the application
func (m *TUI) View() tea.View {
	colors := m.theme.Colors

	// Check minimum window size
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

	// Get mode view (modes handle all layout via their chosen layout)
	modeView := m.currentMode.View()

	// Extract cursor marker from the view and get cursor position
	markerIdx := strings.Index(modeView, cursor.Marker)
	m.logger.Debug("extracting cursor",
		"hasMarker", markerIdx >= 0,
		"markerCount", strings.Count(modeView, cursor.Marker),
		"markerIndex", markerIdx,
		"surroundingText", func() string {
			if markerIdx >= 0 {
				start := markerIdx - 20
				if start < 0 {
					start = 0
				}
				end := markerIdx + len(cursor.Marker) + 20
				if end > len(modeView) {
					end = len(modeView)
				}
				return fmt.Sprintf("%q", modeView[start:end])
			}
			return "NO_MARKER"
		}())
	cleanView, cur := cursor.Extract(modeView)
	m.logger.Debug("extracted cursor",
		"found", cur != nil,
		"stillHasMarker", strings.Contains(cleanView, cursor.Marker),
		"cursorPos", func() string {
			if cur != nil {
				return fmt.Sprintf("(%d,%d)", cur.X, cur.Y)
			}
			return "nil"
		}())

	// Create layers (base layer with page content)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(cleanView),
	}

	// Future: Add dialog/overlay layers here like Crush does
	// if m.dialog.HasDialogs() {
	//     layers = append(layers, m.dialog.GetLayers()...)
	// }

	// Create compositor from layers
	comp := lipgloss.NewCompositor(layers...)

	// Build final view
	view := tea.View{
		BackgroundColor: colors.Page.Bg,
		AltScreen:       true,
	}
	view.Content = comp.Render()
	view.Cursor = cur

	m.logger.Debug("final view",
		"hasMarker", strings.Contains(view.Content, cursor.Marker),
		"cursorX", func() int { if cur != nil { return cur.X }; return -1 }(),
		"cursorY", func() int { if cur != nil { return cur.Y }; return -1 }())
	view.MouseMode = tea.MouseModeCellMotion

	// Show progress bar if supported terminal and we're busy
	if m.sendProgressBar && m.isBusy() {
		// HACK: use a random percentage to prevent ghostty from hiding it
		// after a timeout.
		view.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, rand.IntN(100))
	}

	return view
}
