package root

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/chrome"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	onboardingscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

var quitBinding = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("q", "quit"),
)

const windowTitle = "Tero"

// HeaderModel is the root header contract.
type HeaderModel interface {
	View() string
	SetWidth(width int)
}

// Model is the app shell. It owns top-level lifecycle and composes root screens.
type Model struct {
	scope      logging.Scope
	theme      theme.Theme
	quit       bool
	help       *helpbar.Model
	statusbar  HeaderModel
	onboarding *onboardingscreen.Model
	width      int
	height     int
}

// New constructs the root app shell.
func New(scope logging.Scope, onboarding *onboardingscreen.Model, statusbarModel HeaderModel, appTheme theme.Theme) *Model {
	if onboarding == nil {
		panic("root onboarding model is required")
	}
	if statusbarModel == nil {
		panic("root status bar model is required")
	}
	return &Model{
		scope:      scope,
		theme:      appTheme,
		help:       helpbar.New(),
		statusbar:  statusbarModel,
		onboarding: onboarding,
	}
}

// Init initializes the app shell.
func (m *Model) Init() tea.Cmd {
	m.scope.Info("tui initialized")
	return m.onboarding.Init()
}

// Update applies global app behavior then delegates to the active root screen.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, quitBinding) {
			m.scope.Info("quit requested")
			m.quit = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentWidth := msg.Width - m.theme.Shell.Outer.GetHorizontalFrameSize()
		if contentWidth < 0 {
			contentWidth = 0
		}
		m.statusbar.SetWidth(contentWidth)
		m.help.SetWidth(contentWidth)
	}

	next, cmd := m.onboarding.Update(msg)
	if model, ok := next.(*onboardingscreen.Model); ok {
		m.onboarding = model
	}
	return m, cmd
}

// View renders the active root screen in the shared app frame.
func (m *Model) View() tea.View {
	if m.quit {
		return tea.NewView("")
	}
	bindings := append([]key.Binding{}, m.onboarding.ShortHelp()...)
	bindings = append(bindings, quitBinding)
	view := chrome.Render(
		m.theme,
		chrome.Slots{
			Header: m.statusbar.View(),
			Body:   m.onboarding.View().Content,
			Footer: m.help.Short(bindings),
		},
		chrome.Viewport{Width: m.width, Height: m.height},
	)
	view.AltScreen = true
	view.WindowTitle = windowTitle
	view.MouseMode = tea.MouseModeNone
	return view
}
