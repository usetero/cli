package tui

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	tuiapp "github.com/usetero/cli/internal/tui/app"
	"github.com/usetero/cli/internal/tui/mode"
	"github.com/usetero/cli/internal/tui/onboarding"
	"github.com/usetero/cli/internal/tui/styles"
	"github.com/usetero/cli/internal/workos"
)

const (
	// Minimum window dimensions
	minWidth  = 80
	minHeight = 24

	// WorkOS authentication configuration
	workosBaseURL = "https://api.workos.com"
)

var (
	// DisableMinSizeCheck disables the minimum window size check for testing
	DisableMinSizeCheck = false

	// Global key bindings (immutable after initialization)
	globalQuitBinding = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	globalExitBinding = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("", ""), // Works but not shown in help (avoid redundancy)
	)
	// Only show ctrl+c in help to avoid cluttering the footer
	globalBindings = []key.Binding{globalQuitBinding}
)

// KeyMap defines the global key bindings for the TUI
type KeyMap struct {
	Quit key.Binding
	Exit key.Binding
}

// DefaultKeyMap returns the default global key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: globalQuitBinding,
		Exit: globalExitBinding,
	}
}

// TUI is the top-level model that routes between modes (onboarding, app).
type TUI struct {
	config             *config.Config
	logger             log.Logger
	preferencesService *preferences.Service

	// Current mode (onboarding or app)
	currentMode mode.Mode

	// Global key bindings
	keyMap KeyMap

	// Window dimensions
	width  int
	height int

	// sendProgressBar instructs the TUI to send progress bar updates to the
	// terminal. Only enabled for supported terminals (Windows Terminal, Ghostty).
	sendProgressBar bool
}

// New creates a new TUI model
func New(cfg *config.Config, apiEndpoint string, workosClientID string, logger log.Logger) tea.Model {
	// Create WorkOS client for authentication
	workosClient := workos.NewClient(workosBaseURL, workosClientID)

	// Create keyring for secure token storage
	tokenStore := keyring.New()

	// Create domain services
	authService := auth.NewService(workosClient, tokenStore, logger)
	preferencesService := preferences.NewService(cfg, logger)

	// Start with onboarding mode
	onboardingMode := onboarding.New(logger, authService, preferencesService, apiEndpoint, globalBindings)

	return &TUI{
		config:             cfg,
		logger:             logger,
		preferencesService: preferencesService,
		currentMode:        onboardingMode,
		keyMap:             DefaultKeyMap(),
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Global key bindings (checked first)
		if key.Matches(msg, m.keyMap.Quit) || key.Matches(msg, m.keyMap.Exit) {
			m.logger.Info("user quit", "key", msg.String())
			return m, tea.Quit
		}
	case tea.EnvMsg:
		// Detect Windows Terminal
		if !m.sendProgressBar {
			m.sendProgressBar = slices.Contains(msg, "WT_SESSION")
			if m.sendProgressBar {
				m.logger.Info("enabled progress bar", "terminal", "Windows Terminal")
			}
		}
	case tea.TerminalVersionMsg:
		// Detect Ghostty
		termVersion := strings.ToLower(string(msg))
		m.logger.Debug("received terminal version", "version", termVersion)
		if !m.sendProgressBar {
			m.sendProgressBar = strings.Contains(termVersion, "ghostty")
			if m.sendProgressBar {
				m.logger.Info("enabled progress bar", "terminal", "Ghostty")
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		modeType := fmt.Sprintf("%T", m.currentMode)
		m.logger.Info("window resized",
			"terminalWidth", msg.Width,
			"terminalHeight", msg.Height,
			"mode", modeType)

		// Modes get full terminal dimensions (layouts handle padding)
		m.currentMode.SetSize(msg.Width, msg.Height)

		return m, nil
	}

	// Route message to current mode
	cmd := m.currentMode.Update(msg)

	// Check if mode completed and transition to next mode
	if m.currentMode.IsComplete() {
		switch mode := m.currentMode.(type) {
		case *onboarding.Onboarding:
			// Onboarding complete - extract final state
			orgID := mode.OrganizationID()
			accountID := mode.AccountID()

			m.logger.Info("onboarding completed, transitioning to app",
				"orgID", orgID,
				"accountID", accountID)

			// Create app mode (chat page will be created when we know what services it needs)
			m.currentMode = tuiapp.New(orgID, accountID, m.logger, globalBindings)

			// Set size on new mode before initializing
			if m.width > 0 && m.height > 0 {
				m.currentMode.SetSize(m.width, m.height)
			}

			return m, m.currentMode.Init()
		}
	}

	return m, cmd
}

// isBusy returns true if the TUI is currently performing a background operation
// and should show the progress bar animation
func (m *TUI) isBusy() bool {
	return m.currentMode.IsBusy()
}

// View renders the application
func (m *TUI) View() tea.View {
	theme := styles.CurrentTheme()

	// Check minimum window size
	if !DisableMinSizeCheck && (m.width < minWidth || m.height < minHeight) {
		view := tea.View{
			BackgroundColor: theme.Background,
			AltScreen:       true,
		}

		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(
				lipgloss.NewStyle().
					Width(m.width).
					Height(m.height).
					Align(lipgloss.Center, lipgloss.Center).
					Render(
						lipgloss.NewStyle().
							Padding(1, 4).
							Foreground(theme.Text).
							BorderStyle(lipgloss.RoundedBorder()).
							BorderForeground(theme.Primary).
							Render("Window too small!"),
					),
			),
		)

		return view
	}

	// Get mode view (modes handle all layout via their chosen layout)
	modeView := m.currentMode.View()

	// Extract cursor before creating layers
	finalView, cursor := ExtractCursor(modeView)

	// Create layers (base layer with page content)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(finalView),
	}

	// Future: Add dialog/overlay layers here like Crush does
	// if m.dialog.HasDialogs() {
	//     layers = append(layers, m.dialog.GetLayers()...)
	// }

	// Create canvas from layers
	canvas := lipgloss.NewCanvas(layers...)

	// Build final view
	view := tea.View{
		BackgroundColor: theme.Background,
		AltScreen:       true,
	}
	view.Layer = canvas
	view.Cursor = cursor
	view.MouseMode = tea.MouseModeCellMotion

	// Show progress bar if supported terminal and we're busy
	if m.sendProgressBar && m.isBusy() {
		// HACK: use a random percentage to prevent ghostty from hiding it
		// after a timeout.
		view.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, rand.IntN(100))
	}

	return view
}
