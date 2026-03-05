package auth

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
)

// View renders the authenticate UI.
func (m *AuthenticateModel) View() string {
	s := m.theme.Styles

	if m.state == stateInitializing {
		return s.Title.Render("Initializing authentication...")
	}

	if m.device == nil {
		return s.Title.Render("Loading...")
	}

	var parts []string
	parts = append(parts, s.Title.Render("Authenticate with Tero"), "")
	parts = append(parts,
		s.Body.Render("Your code: ")+s.Title.Render(m.device.UserCode),
		"",
	)
	parts = append(parts,
		s.Body.Render("Visit this URL to sign in:"),
		s.URL.Render(m.device.VerificationURIComplete),
		"",
	)
	parts = append(parts,
		s.Body.Render("Confirm the code matches, then click \"Confirm\" in your browser."),
		"",
	)

	mutedStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Background(m.theme.Bg)

	switch {
	case m.state == statePolling:
		parts = append(parts, m.spinner.View()+" "+mutedStyle.Render("Waiting for authentication..."))
	case m.browserFailed:
		parts = append(parts, s.Error.Render("Couldn't open browser. Press 'c' to copy URL"))
	case m.copiedToClipboard:
		parts = append(parts, s.Success.Render("URL copied to clipboard"))
	case m.err != nil:
		parts = append(parts,
			s.Error.Render(errorfmt.UserFacing(m.err, "Authentication failed.")),
			s.Help.Render("Press 'r' to retry"),
		)
	default:
		parts = append(parts, s.Action.Render("Press Enter to open in browser, or 'c' to copy URL"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
