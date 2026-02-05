package datadog

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
)

// RegionModel handles Datadog region selection.
type RegionModel struct {
	theme    *styles.Theme
	scope    log.Scope
	selected int
	width    int
	height   int
}

// NewRegion creates a new region selection step.
func NewRegion(theme *styles.Theme, scope log.Scope) *RegionModel {
	if theme == nil {
		panic("theme is nil")
	}

	return &RegionModel{
		theme: theme,
		scope: scope,
	}
}

// Init returns nil.
func (m *RegionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *RegionModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.selected < len(domain.DatadogRegions)-1 {
				m.selected++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			site := domain.DatadogRegions[m.selected].Site
			m.scope.Info("datadog region selected", "site", site)
			return func() tea.Msg {
				return msgs.DatadogRegionSelected{Site: site}
			}
		}
	}
	return nil
}

// View renders the region selection UI.
func (m *RegionModel) View() string {
	s := m.theme.Styles
	colors := m.theme.Colors

	title := s.Title.Render("Select your Datadog region")
	subtitle := s.Help.Render("Choose the region where your Datadog account is hosted")

	var optionViews []string
	for i, r := range domain.DatadogRegions {
		var view string
		if i == m.selected {
			nameStyle := lipgloss.NewStyle().Foreground(colors.Accent).Bold(true)
			view = nameStyle.Render("> " + r.Name)
		} else {
			view = s.Body.Render("  " + r.Name)
		}
		optionViews = append(optionViews, view)
	}

	parts := []string{title, subtitle, ""}
	parts = append(parts, optionViews...)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize updates dimensions.
func (m *RegionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
