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
	"github.com/usetero/cli/internal/tui/onboarding/organization"
	"github.com/usetero/cli/internal/tui/onboarding/step"
)

const (
	Platform = "platform"
	Engineer = "engineer"
)

// KeyMap defines key bindings for role selection
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}

// DefaultKeyMap returns the default key bindings
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

// SelectStep handles selecting the user's role in the organization.
type SelectStep struct {
	// Lifecycle context for cancellation
	ctx context.Context

	// Theme
	theme *styles.Theme

	// Services
	preferences preferences.Preferences
	auth        auth.Auth
	apiClient   api.Client
	logger      log.Logger

	// UI state
	selected       int // 0 = Platform Team, 1 = Service Owner
	err            error
	width          int
	keyMap         KeyMap
	globalBindings []key.Binding
}

// NewSelectStep creates a new role selection step.
func NewSelectStep(ctx context.Context, theme *styles.Theme, apiClient api.Client, prefs preferences.Preferences, authService auth.Auth, logger log.Logger, globalBindings []key.Binding) step.Step {
	if apiClient == nil {
		panic("apiClient cannot be nil")
	}
	if prefs == nil {
		panic("prefs cannot be nil")
	}
	if authService == nil {
		panic("authService cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	// Load saved role and set selected to match
	savedRole := prefs.GetRole()

	// Set selected index based on saved role
	var selected int
	switch savedRole {
	case Platform:
		selected = 0
	case Engineer:
		selected = 1
	default:
		// No saved role, default to 0 (Platform)
		selected = 0
	}

	if savedRole != "" {
		logger.Debug("role already saved", "role", savedRole)
	}

	return &SelectStep{
		ctx:            ctx,
		theme:          theme,
		preferences:    prefs,
		auth:           authService,
		apiClient:      apiClient,
		logger:         logger,
		selected:       selected,
		width:          80,
		keyMap:         DefaultKeyMap(),
		globalBindings: globalBindings,
	}
}

// Init initializes the role step.
func (s *SelectStep) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *SelectStep) Update(msg tea.Msg) (step.Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.selected > 0 {
				s.selected--
			}
		case "down", "j":
			if s.selected < 1 {
				s.selected++
			}
		case "enter":
			// Save the selected role to preferences (for next time)
			role := Platform
			if s.selected == 1 {
				role = Engineer
			}

			s.logger.Info("role selected", "role", role)

			if err := s.preferences.SetRole(role); err != nil {
				s.logger.Error("failed to save role", "error", err)
				s.err = err
				return s, nil
			}

			// Clear any previous error and mark success
			s.err = nil
			s.logger.Debug("role saved to preferences", "role", role)
			return s, nil
		case "r":
			// Retry saving if there was an error
			if s.err != nil {
				role := Platform
				if s.selected == 1 {
					role = Engineer
				}

				s.logger.Info("retrying role save", "role", role)

				if err := s.preferences.SetRole(role); err != nil {
					s.logger.Error("failed to save role", "error", err)
					s.err = err
					return s, nil
				}

				// Success - clear error
				s.err = nil
				s.logger.Debug("role saved to preferences", "role", role)
				return s, nil
			}
		}
	}

	return s, nil
}

// View renders the role selection UI.
func (s *SelectStep) View() string {
	themeStyles := s.theme.Styles
	colors := s.theme.Colors

	title := themeStyles.Title.Render("What's your role in this organization?")

	// Options
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
		if i == s.selected {
			// Selected option
			nameStyle := lipgloss.NewStyle().
				Foreground(colors.Accent).
				Bold(true)

			view = nameStyle.Render("> "+opt.name) + "\n  " + themeStyles.Help.Render(opt.description)
		} else {
			// Unselected option
			view = themeStyles.Body.Render("  "+opt.name) + "\n  " + themeStyles.Help.Render(opt.description)
		}
		optionViews = append(optionViews, view)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		optionViews[0],
		"",
		optionViews[1],
	)

	return content
}

// SetSize sets the width for rendering.
func (s *SelectStep) SetSize(width, height int) {
	s.width = width
}

// IsBusy returns false - role selection is never busy.
func (s *SelectStep) IsBusy() bool {
	return false
}

// HasError returns true if there was an error saving the role
func (s *SelectStep) HasError() bool {
	return s.err != nil
}

// Error returns the current error, or nil if no error
func (s *SelectStep) Error() error {
	return s.err
}

// Next returns the next step after role selection
func (s *SelectStep) Next() (step.Step, error) {
	if s.err != nil {
		return nil, s.err
	}

	role := s.preferences.GetRole()
	if role != Platform && role != Engineer {
		return nil, step.ErrNotReady
	}

	s.logger.Debug("role step complete", "role", role)

	// Create organization service for next step
	organizationService := api.NewOrganizationService(s.apiClient, s.logger)

	// Pass accumulated context (role), organization service, client, preferences, auth, and logger to next step
	return organization.NewSelectStep(s.ctx, s.theme, role, organizationService, s.apiClient, s.preferences, s.auth, s.logger, s.globalBindings), nil
}

// Help returns the key bindings for this step
func (s *SelectStep) Help() help.KeyMap {
	// Show retry option if there's an error
	if s.err != nil {
		return keymap.Simple{
			Keys: []key.Binding{
				key.NewBinding(
					key.WithKeys("r"),
					key.WithHelp("r", "retry"),
				),
			},
		}
	}

	// Normal state: show navigation and selection
	return keymap.Simple{
		Keys: []key.Binding{
			s.keyMap.Up,
			s.keyMap.Down,
			s.keyMap.Select,
		},
	}
}
