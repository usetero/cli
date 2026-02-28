package role

import "charm.land/lipgloss/v2"

// View renders the role selection UI.
func (m *Model) View() string {
	s := m.theme.Styles

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
			nameStyle := lipgloss.NewStyle().Foreground(m.theme.Accent).Background(m.theme.Bg).Bold(true)
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
