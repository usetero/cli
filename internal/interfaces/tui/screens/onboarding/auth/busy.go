package auth

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

func (m *Model) Busy() *core.Busy {
	surfaceTheme := m.theme.OnSurface()

	switch m.phase {
	case phaseStarting:
		return &core.Busy{
			Label:  "Starting Secure Browser Sign-In",
			Detail: surfaceTheme.Text.Body.Render("Requesting your verification URL and confirmation code..."),
		}
	case phaseWaiting:
		busy := &core.Busy{
			Label: "Complete Sign-In In Your Browser",
		}
		if uri := m.browserURL(); uri != "" {
			lines := []string{
				surfaceTheme.Text.Body.Render(uri),
			}
			if m.flow.UserCode != "" {
				codeLine := lipgloss.JoinHorizontal(
					lipgloss.Left,
					surfaceTheme.Text.Body.Render("Confirmation code: "),
					lipgloss.NewStyle().
						Foreground(surfaceTheme.Palette.Text).
						Background(surfaceTheme.Background).
						Bold(true).
						Render(m.flow.UserCode),
				)
				lines = append(lines, "", codeLine)
			}
			lines = append(lines, "", surfaceTheme.Text.Body.Render("Tero will continue automatically when sign-in is complete."))
			busy.Detail = strings.Join(lines, "\n")
		} else {
			busy.Detail = strings.Join([]string{
				surfaceTheme.Text.Body.Render("Finish signing in from the browser window."),
				"",
				surfaceTheme.Text.Body.Render("Tero will continue automatically when sign-in is complete."),
			}, "\n")
		}
		return busy
	default:
		return nil
	}
}
