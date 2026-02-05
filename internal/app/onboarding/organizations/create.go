package organizations

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/input"
)

// orgCreatedMsg is sent when org creation completes.
type orgCreatedMsg struct {
	result *api.OrganizationBootstrapResult
	err    error
}

// CreateModel handles organization creation.
type CreateModel struct {
	ctx      context.Context
	theme    *styles.Theme
	services api.APIServices
	prefs    preferences.Preferences
	logger   log.Logger

	input    *input.Model
	creating bool
	err      error
	width    int
	height   int
}

// NewCreate creates a new organization creation step.
func NewCreate(
	ctx context.Context,
	theme *styles.Theme,
	services api.APIServices,
	prefs preferences.Preferences,
	logger log.Logger,
) *CreateModel {
	if ctx == nil {
		panic("ctx is nil")
	}
	if theme == nil {
		panic("theme is nil")
	}
	if prefs == nil {
		panic("prefs is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	inp := input.New(theme)
	inp.SetPlaceholder("Organization name")
	inp.SetCharLimit(100)

	return &CreateModel{
		ctx:      ctx,
		theme:    theme,
		services: services,
		prefs:    prefs,
		logger:   logger,
		input:    inp,
	}
}

// Init focuses the input.
func (m *CreateModel) Init() tea.Cmd {
	return m.input.Focus()
}

// Update handles messages.
func (m *CreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orgCreatedMsg:
		m.creating = false
		if msg.err != nil {
			m.logger.Error("failed to create organization", "error", msg.err)
			m.err = msg.err
			return nil
		}
		_ = m.prefs.SetDefaultOrgID(msg.result.Organization.ID)
		org := *msg.result.Organization
		m.logger.Info("organization created", "id", org.ID, "name", org.Name)
		return func() tea.Msg { return msgs.OrgCreated{Org: org} }

	case tea.KeyPressMsg:
		if m.creating {
			return nil
		}
		switch msg.String() {
		case "enter":
			name := m.input.Value()
			if name == "" {
				return nil
			}
			m.creating = true
			m.err = nil
			m.logger.Info("creating organization", "name", name)
			return m.createOrg(name)
		}
	}

	return m.input.Update(msg)
}

func (m *CreateModel) createOrg(name string) tea.Cmd {
	return func() tea.Msg {
		id := uuid.New()
		result, err := m.services.Organizations.Create(m.ctx, id, name)
		if err != nil {
			return orgCreatedMsg{err: err}
		}
		return orgCreatedMsg{result: result}
	}
}

// View renders the organization creation UI.
func (m *CreateModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Create an organization")
	subtitle := s.Help.Render("Organizations contain accounts and team members")

	var status string
	if m.creating {
		status = s.Help.Render("Creating...")
	} else if m.err != nil {
		status = s.Error.Render("Failed to create organization. Try again.")
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *CreateModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.SetWidth(width)
}
