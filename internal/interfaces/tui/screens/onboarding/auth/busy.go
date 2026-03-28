package auth

import "github.com/usetero/cli/internal/interfaces/tui/core"

func (m *Model) Busy() *core.Busy {
	switch m.phase {
	case phaseStarting:
		return &core.Busy{
			Label:  "Starting Secure Browser Sign-In",
			Detail: "Requesting your verification URL and confirmation code...",
		}
	case phaseWaiting:
		busy := &core.Busy{
			Label: "Finish Authentication In Your Browser",
		}
		if uri := m.browserURL(); uri != "" {
			busy.Detail = uri
			if m.flow.UserCode != "" {
				busy.Detail += "\n\nConfirmation code: " + m.flow.UserCode
			}
		} else {
			busy.Detail = "Finish signing in from the browser window."
		}
		return busy
	default:
		return nil
	}
}
