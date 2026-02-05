package blocks

import tea "charm.land/bubbletea/v2"

// Block is the interface for all content blocks in an assistant message.
type Block interface {
	Index() int
	Update(tea.Msg) tea.Cmd
	View() string
	SetWidth(int)
}
