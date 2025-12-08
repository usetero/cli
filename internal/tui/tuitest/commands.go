package tuitest

import tea "github.com/charmbracelet/bubbletea/v2"

// DrainCmds executes all commands from a tea.Cmd, handling batches recursively.
// Returns all resulting messages flattened into a single slice.
func DrainCmds(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	if msg == nil {
		return nil
	}

	// tea.Batch returns a tea.BatchMsg which is []tea.Cmd
	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, innerCmd := range batch {
			msgs = append(msgs, DrainCmds(innerCmd)...)
		}
		return msgs
	}

	return []tea.Msg{msg}
}
