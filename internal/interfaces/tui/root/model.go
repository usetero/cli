package root

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/tui/browser"
	"github.com/usetero/cli/internal/interfaces/tui/chrome"
	"github.com/usetero/cli/internal/interfaces/tui/components/helpbar"
	"github.com/usetero/cli/internal/interfaces/tui/cursor"
	onboardingscreen "github.com/usetero/cli/internal/interfaces/tui/screens/onboarding"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

type keyMsg = tea.KeyPressMsg

var immediateQuitBinding = key.NewBinding(
	key.WithKeys("ctrl+c"),
	key.WithHelp("ctrl+c", "quit"),
)

var confirmQuitBinding = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "quit"),
)

const windowTitle = "Tero"
const (
	defaultWidth  = 80
	defaultHeight = 24
)

// HeaderModel is the root header contract.
type HeaderModel interface {
	View() string
	SetWidth(width int)
}

type bodyModel interface {
	View() tea.View
	ShortHelp() []key.Binding
	SetSize(width, height int)
	Layout() chrome.BodyLayout
}

// Model is the app shell. It owns top-level lifecycle and composes root screens.
type Model struct {
	scope      logging.Scope
	theme      theme.Theme
	quit       bool
	quitDialog *quitDialog
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
		help:       helpbar.New(appTheme),
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
		if key.Matches(msg, immediateQuitBinding) {
			m.scope.Info("quit requested", "mode", "immediate")
			m.quit = true
			return m, tea.Quit
		}
		if m.quitDialog != nil {
			switch m.quitDialog.Update(msg) {
			case quitDialogConfirm:
				m.scope.Info("quit requested", "mode", "confirmed")
				m.quit = true
				return m, tea.Quit
			case quitDialogCancel:
				m.quitDialog = nil
				m.syncLayout()
				return m, nil
			default:
				return m, nil
			}
		}
		if key.Matches(msg, confirmQuitBinding) {
			m.quitDialog = newQuitDialog(m.theme)
			m.syncLayout()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case browser.OpenRequestedMsg:
		return m, browser.OpenURLCmd(msg.URL)
	case browser.OpenCompletedMsg:
		if msg.Err != nil {
			m.scope.Error("open browser failed", "url", msg.URL, "error", msg.Err)
			return m, nil
		}
		m.scope.Info("opened browser", "url", msg.URL)
		return m, nil
	}

	next, cmd := m.onboarding.Update(msg)
	if model, ok := next.(*onboardingscreen.Model); ok {
		m.onboarding = model
	}
	m.syncLayout()
	return m, cmd
}

// View renders the active root screen in the shared app frame.
func (m *Model) View() tea.View {
	if m.quit {
		return tea.NewView("")
	}
	body := chrome.BodySlot{
		Content: m.onboarding.View().Content,
		Layout:  m.onboarding.Layout(),
	}
	bindings := m.currentBindings()
	if m.quitDialog != nil {
		body = chrome.BodySlot{
			Content: m.quitDialog.View(),
			Layout: chrome.BodyLayout{
				WidthMode:     chrome.WidthIntrinsic,
				HeightMode:    chrome.HeightIntrinsic,
				VerticalAlign: chrome.AlignCenter,
				MaxWidth:      48,
			},
		}
	}
	width := m.width
	if width <= 0 {
		width = defaultWidth
	}
	height := m.height
	if height <= 0 {
		height = defaultHeight
	}
	view := chrome.Render(
		m.theme,
		chrome.Slots{
			Header: m.statusbar.View(),
			Body:   body,
			Footer: m.help.Short(bindings),
		},
		chrome.Viewport{Width: width, Height: height},
	)
	clean, cur := cursor.Extract(view.Content)
	view.Content = clean
	if cur != nil {
		cur.Color = m.theme.Accent
	}
	if m.quitDialog != nil {
		cur = nil
	}
	view.AltScreen = true
	view.Cursor = cur
	view.WindowTitle = windowTitle
	view.MouseMode = tea.MouseModeNone
	view.BackgroundColor = m.theme.Background
	return view
}

func (m *Model) syncLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	screen := bodyModel(m.onboarding)
	contentWidth := m.width - m.theme.Shell.Outer.GetHorizontalFrameSize()
	if contentWidth < 0 {
		contentWidth = 0
	}
	m.statusbar.SetWidth(contentWidth)

	bindings := m.currentBindings()
	m.help.SetWidth(contentWidth)

	metrics := chrome.Measure(
		m.theme,
		chrome.Slots{
			Header: m.statusbar.View(),
			Body: chrome.BodySlot{
				Layout: screen.Layout(),
			},
			Footer: m.help.Short(bindings),
		},
		chrome.Viewport{Width: m.width, Height: m.height},
	)
	screen.SetSize(metrics.BodyContentWidth, metrics.BodyContentHeight)
}

func (m *Model) currentBindings() []key.Binding {
	if m.quitDialog != nil {
		return append([]key.Binding{}, m.quitDialog.ShortHelp()...)
	}
	bindings := append([]key.Binding{}, m.onboarding.ShortHelp()...)
	bindings = append(bindings, confirmQuitBinding, immediateQuitBinding)
	return bindings
}
