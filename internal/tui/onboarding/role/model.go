package role

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tui/keymap"
	organizationselect "github.com/usetero/cli/internal/tui/onboarding/organization/select"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const (
	Platform = "platform"
	Engineer = "engineer"
)

// KeyMap defines key bindings for role selection.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

// Model handles role selection.
type Model struct {
	ctx   context.Context
	theme *styles.Theme

	services api.APIServices
	prefs    preferences.Preferences
	auth     auth.Auth
	logger   log.Logger

	selected int
	done     bool
	role     string
	err      error
	width    int
	height   int
	keyMap   KeyMap
}

// New creates a new role selection model.
func New(
	ctx context.Context,
	theme *styles.Theme,
	services api.APIServices,
	prefs preferences.Preferences,
	authService auth.Auth,
	logger log.Logger,
) Model {
	// Load saved role and set selected to match
	savedRole := prefs.GetRole()
	var selected int
	switch savedRole {
	case Platform:
		selected = 0
	case Engineer:
		selected = 1
	default:
		selected = 0
	}

	return Model{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		auth:     authService,
		logger:   logger,
		selected: selected,
		keyMap:   DefaultKeyMap(),
		width:    80,
	}
}

// roleCheckedMsg is sent after checking for a saved role.
type roleCheckedMsg struct {
	savedRole string
}

// Init initializes the step.
func (m Model) Init() tea.Cmd {
	return m.checkSavedRole()
}

// checkSavedRole checks if there's a saved role preference.
func (m Model) checkSavedRole() tea.Cmd {
	return func() tea.Msg {
		return roleCheckedMsg{savedRole: m.prefs.GetRole()}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case roleCheckedMsg:
		if msg.savedRole == Platform || msg.savedRole == Engineer {
			m.role = msg.savedRole
			m.done = true
			m.logger.Debug("auto-selected role from preference", "role", msg.savedRole)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.selected < 1 {
				m.selected++
			}
		case key.Matches(msg, m.keyMap.Select):
			m.role = Platform
			if m.selected == 1 {
				m.role = Engineer
			}

			// Save to preferences
			if err := m.prefs.SetRole(m.role); err != nil {
				m.logger.Error("failed to save role", "error", err)
				m.err = err
				return m, nil
			}

			m.done = true
			m.err = nil
			m.logger.Info("role selected", "role", m.role)
		}
	}

	return m, nil
}

// View renders the role selection UI.
func (m Model) View() string {
	themeStyles := m.theme.Styles
	colors := m.theme.Colors

	title := themeStyles.Title.Render("What's your role?")

	options := []struct {
		name        string
		description string
	}{
		{
			name:        "Platform / Observability Team",
			description: "I'm responsible for observability across the organization",
		},
		{
			name:        "Service Owner / Engineer",
			description: "I work on specific services and own their observability",
		},
	}

	var optionViews []string
	for i, opt := range options {
		var view string
		if i == m.selected {
			nameStyle := lipgloss.NewStyle().
				Foreground(colors.Accent).
				Bold(true)
			view = nameStyle.Render("> "+opt.name) + "\n  " + themeStyles.Help.Render(opt.description)
		} else {
			view = themeStyles.Body.Render("  "+opt.name) + "\n  " + themeStyles.Help.Render(opt.description)
		}
		optionViews = append(optionViews, view)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		optionViews[0],
		"",
		optionViews[1],
	)
}

// SetSize returns a new Model with the given dimensions.
func (m Model) SetSize(width, height int) step.Step {
	m.width = width
	m.height = height
	return m
}

// IsBusy returns false.
func (m Model) IsBusy() bool {
	return false
}

// HasError returns true if there's an error.
func (m Model) HasError() bool {
	return m.err != nil
}

// Error returns the current error.
func (m Model) Error() error {
	return m.err
}

// Help returns the key bindings for this step.
func (m Model) Help() help.KeyMap {
	if m.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}
	return keymap.Simple{
		Keys: []key.Binding{
			m.keyMap.Up,
			m.keyMap.Down,
			m.keyMap.Select,
		},
	}
}

// Next returns the next step.
func (m Model) Next() (step.Step, error) {
	if m.err != nil {
		return nil, m.err
	}
	if !m.done {
		return nil, step.ErrNotReady
	}

	m.logger.Debug("role step complete", "role", m.role)

	return organizationselect.New(
		m.ctx,
		m.theme,
		m.role,
		m.services,
		m.prefs,
		m.auth,
		m.logger,
	), nil
}

// Role returns the selected role.
func (m Model) Role() string {
	return m.role
}

// Close releases resources.
func (m Model) Close() error {
	return nil
}
