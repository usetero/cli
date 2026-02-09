package msgs

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/preferences"
)

// SetTheme requests a theme change.
type SetTheme struct {
	Theme preferences.Theme
}

// Error signals an error to show to the user.
type Error struct {
	Message string // user-friendly message
	Err     error  // underlying error (for logging)
	Sticky  bool
}

func ErrorCmd(message string, err error, sticky bool) tea.Cmd {
	return func() tea.Msg { return Error{Message: message, Err: err, Sticky: sticky} }
}

// Warning signals a warning to show to the user.
type Warning struct {
	Message string
	Sticky  bool
}

func WarningCmd(message string, sticky bool) tea.Cmd {
	return func() tea.Msg { return Warning{Message: message, Sticky: sticky} }
}

// Success signals a success message to show to the user.
type Success struct {
	Message string
}

func SuccessCmd(message string) tea.Cmd {
	return func() tea.Msg { return Success{Message: message} }
}

// Info signals an info message to show to the user.
type Info struct {
	Message string
}

func InfoCmd(message string) tea.Cmd {
	return func() tea.Msg { return Info{Message: message} }
}
