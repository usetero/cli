package events

// DrawerPrompt is emitted by a statusbar drawer tab to submit a
// context-specific prompt to the chat. The app model closes the
// drawer and forwards the text as a UserSubmittedInput.
type DrawerPrompt struct {
	Text string
}
