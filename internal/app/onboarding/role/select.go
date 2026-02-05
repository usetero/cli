// Package role provides the role selection step for onboarding.
package role

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// savedRoleMsg is sent when checking for a saved role.
type savedRoleMsg struct {
	role string
}

// Model handles role selection.
type Model struct {
	theme    *styles.Theme
	prefs    preferences.Preferences
	logger   log.Logger
	selected int
	width    int
	height   int
}

// New creates a new role selection step.
func New(theme *styles.Theme, prefs preferences.Preferences, logger log.Logger) *Model {
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}
	return &Model{
		theme:  theme,
		prefs:  prefs,
		logger: logger,
	}
}

// Init checks for a saved role preference.
func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		return savedRoleMsg{role: m.prefs.GetRole()}
	}
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case savedRoleMsg:
		if msg.role == msgs.RolePlatform || msg.role == msgs.RoleEngineer {
			m.logger.Debug("using saved role preference", "role", msg.role)
			return func() tea.Msg {
				return msgs.RoleSelected{Role: msg.role}
			}
		}
		if msg.role == msgs.RoleEngineer {
			m.selected = 1
		}
		return nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.selected < 1 {
				m.selected++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			role := msgs.RolePlatform
			if m.selected == 1 {
				role = msgs.RoleEngineer
			}
			_ = m.prefs.SetRole(role)
			m.logger.Info("role selected", "role", role)
			return func() tea.Msg {
				return msgs.RoleSelected{Role: role}
			}
		}
	}
	return nil
}

// View renders the role selection UI.
func (m *Model) View() string {
	s := m.theme.Styles
	colors := m.theme.Colors

	title := s.Title.Render("What's your role?")

	options := []struct {
		name string
		desc string
	}{
		{"Platform / Observability Team", "I'm responsible for observability across the organization"},
		{"Service Owner / Engineer", "I work on specific services and own their observability"},
	}

	var optionViews []string
	for i, opt := range options {
		var view string
		if i == m.selected {
			nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Bold(true)
			view = nameStyle.Render("> "+opt.name) + "\n  " + s.Help.Render(opt.desc)
		} else {
			view = s.Body.Render("  "+opt.name) + "\n  " + s.Help.Render(opt.desc)
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

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}
