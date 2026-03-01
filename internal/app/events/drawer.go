package events

import tea "charm.land/bubbletea/v2"

// DrawerPromptRequested is emitted by a statusbar drawer tab to submit a
// context-specific prompt to the chat. The app model closes the
// drawer and forwards the text as a UserSubmittedInput.
type DrawerPromptRequested struct {
	Text string
}

// RequestDrawerPromptCmd emits a DrawerPromptRequested event with the provided text.
func RequestDrawerPromptCmd(text string) tea.Cmd {
	return func() tea.Msg { return DrawerPromptRequested{Text: text} }
}
