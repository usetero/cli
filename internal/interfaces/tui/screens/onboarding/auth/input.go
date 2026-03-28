package auth

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/interfaces/tui/core"
)

var reopenBinding = key.NewBinding(
	key.WithKeys("enter", "o"),
	key.WithHelp("enter/o", "open browser"),
)

var startBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "start"),
)

var retryBinding = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "retry"),
)

func (m *Model) ShortHelp() []key.Binding {
	switch m.phase {
	case phaseIdle:
		return []key.Binding{startBinding}
	case phaseWaiting, phaseFailed:
		if m.phase == phaseFailed && m.browserURL() == "" {
			return []key.Binding{retryBinding}
		}
		return []key.Binding{reopenBinding}
	default:
		return nil
	}
}

func (m *Model) Input() *core.Input {
	input := &core.Input{
		Kind:  core.InputAction,
		Label: "Welcome to Tero. Please log in or create your account.",
	}

	switch m.phase {
	case phaseIdle:
		input.Action = "Press enter to get started."
	case phaseStarting:
		input.Action = "Press enter to get started."
	case phaseWaiting:
		input.Action = "Press enter to open the browser again."
	case phaseAuthenticated:
		input.Label = "Authentication complete."
		if email := strings.TrimSpace(m.user.Email); email != "" {
			input.Label = "Authenticated as " + email + "."
		}
		input.Action = "Continuing to organization setup..."
	case phaseFailed:
		if m.browserURL() == "" {
			input.Action = "Press enter to try again."
		} else {
			input.Action = "Press enter to open the browser again."
		}
	}

	return input
}
