package auth

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

var reopenBinding = key.NewBinding(
	key.WithKeys("o"),
	key.WithHelp("o", "open browser"),
)

func (m *Model) ShortHelp() []key.Binding {
	switch m.phase {
	case phaseWaiting, phaseFailed:
		if m.phase == phaseFailed && m.browserURL() == "" {
			return nil
		}
		return []key.Binding{reopenBinding}
	default:
		return nil
	}
}

func (m *Model) Input() *core.Input {
	input := &core.Input{
		Kind:  core.InputConfirm,
		Label: "Welcome to Tero. Please log in or create your account.",
	}

	switch m.phase {
	case phaseIdle:
		input.Action = "Get started"
	case phaseStarting:
		input.Action = "Get started"
	case phaseWaiting:
		input.Action = "Open browser again"
	case phaseAuthenticated:
		input.Label = "Authentication complete."
		if email := strings.TrimSpace(m.user.Email); email != "" {
			input.Label = "Authenticated as " + email + "."
		}
		input.Action = ""
	case phaseFailed:
		if m.browserURL() == "" {
			input.Action = "Try again"
		} else {
			input.Action = "Open browser again"
		}
	}

	return input
}
