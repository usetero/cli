package organizations

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
)

// View renders the organization creation UI.
func (m *CreateModel) View() string {
	s := m.theme.Styles

	title := s.Title.Render("Create an organization")
	subtitle := s.Help.Render("Organizations contain accounts and team members")

	var status string
	if m.creating {
		status = s.Help.Render("Creating...")
	} else if m.err != nil {
		status = s.Error.Render(errorfmt.UserFacing(m.err, "Failed to create organization. Try again."))
	}

	parts := []string{title, subtitle, "", m.input.View()}
	if status != "" {
		parts = append(parts, "", status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
