package events

import tea "charm.land/bubbletea/v2"

// DrawerPrompt is emitted by a statusbar drawer tab to submit a
// context-specific prompt to the chat. The app model closes the
// drawer and forwards the text as a UserSubmittedInput.
type DrawerPrompt struct {
	Text string
}

// DrawerPromptCmd emits a DrawerPrompt event with the provided text.
func DrawerPromptCmd(text string) tea.Cmd {
	return func() tea.Msg { return DrawerPrompt{Text: text} }
}
