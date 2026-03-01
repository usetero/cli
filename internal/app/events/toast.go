package events

import (
	tea "charm.land/bubbletea/v2"

	"github.com/usetero/cli/internal/preferences"
)

// ThemeChangeRequested requests a theme change.
type ThemeChangeRequested struct {
	Theme preferences.Theme
}

// ErrorToastPublished signals an error toast to show to the user.
type ErrorToastPublished struct {
	Message string // user-friendly message
	Err     error  // underlying error (for logging)
	Sticky  bool
}

func PublishErrorToastCmd(message string, err error, sticky bool) tea.Cmd {
	return func() tea.Msg { return ErrorToastPublished{Message: message, Err: err, Sticky: sticky} }
}

// WarningToastPublished signals a warning toast to show to the user.
type WarningToastPublished struct {
	Message string
	Sticky  bool
}

func PublishWarningToastCmd(message string, sticky bool) tea.Cmd {
	return func() tea.Msg { return WarningToastPublished{Message: message, Sticky: sticky} }
}

// SuccessToastPublished signals a success toast to show to the user.
type SuccessToastPublished struct {
	Message string
}

func PublishSuccessToastCmd(message string) tea.Cmd {
	return func() tea.Msg { return SuccessToastPublished{Message: message} }
}

// InfoToastPublished signals an info toast to show to the user.
type InfoToastPublished struct {
	Message string
}

func PublishInfoToastCmd(message string) tea.Cmd {
	return func() tea.Msg { return InfoToastPublished{Message: message} }
}
