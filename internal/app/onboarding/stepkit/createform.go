package stepkit

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/app/onboarding/errorfmt"
	"github.com/usetero/cli/internal/styles"
)

// CreateInputShortHelp returns consistent short help for input-based create steps.
func CreateInputShortHelp(creating bool, base []key.Binding) []key.Binding {
	if creating {
		return nil
	}
	bindings := append([]key.Binding{}, base...)
	bindings = append(bindings, key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "create")))
	return bindings
}

// ParseCreateSubmit returns a submitted name when enter is pressed and create is allowed.
func ParseCreateSubmit(msg tea.KeyPressMsg, creating bool, value string) (string, bool) {
	if creating || msg.String() != "enter" || value == "" {
		return "", false
	}
	return value, true
}

// RenderCreateForm renders a standard create step with title/subtitle/input and status line.
func RenderCreateForm(theme styles.Theme, title string, subtitle string, inputView string, creating bool, err error, failureFallback string) string {
	s := theme.Styles

	var status string
	if creating {
		status = s.Help.Render("Creating...")
	} else if err != nil {
		status = s.Error.Render(errorfmt.UserFacing(err, failureFallback))
	}

	parts := []string{
		s.Title.Render(title),
		s.Help.Render(subtitle),
		"",
		inputView,
	}
	if status != "" {
		parts = append(parts, "", status)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
