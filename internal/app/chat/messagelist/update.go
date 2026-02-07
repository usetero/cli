package messagelist

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.focused {
			switch {
			case key.Matches(msg, focusPrevKey):
				m.vp.FocusPrev()
				m.userScrolled = true
			case key.Matches(msg, focusNextKey):
				m.vp.FocusNext()
				if m.vp.AtBottom() {
					m.userScrolled = false
				}
			case key.Matches(msg, scrollUpKey):
				m.vp.ScrollBy(-1)
				m.userScrolled = true
				m.vp.UpdateFocusFromScroll()
			case key.Matches(msg, scrollDownKey):
				m.vp.ScrollBy(1)
				if m.vp.AtBottom() {
					m.userScrolled = false
				}
				m.vp.UpdateFocusFromScroll()
			}
		}

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			viewX := msg.X - m.originX - outerBorderWidth
			viewY := msg.Y - m.originY
			blockIdx, blockY := m.vp.ItemAtY(viewY)
			if blockIdx >= 0 {
				m.vp.SetFocusIdx(blockIdx)
				m.mouseDown = true
				m.mouseDownBlock = blockIdx
				m.mouseDownX = viewX
				m.mouseDownY = blockY
				m.mouseDragBlock = blockIdx
				m.mouseDragX = viewX
				m.mouseDragY = blockY
			}
		}

	case tea.MouseMotionMsg:
		if m.mouseDown && msg.Button == tea.MouseLeft {
			viewX := msg.X - m.originX - outerBorderWidth
			viewY := msg.Y - m.originY
			blockIdx, blockY := m.vp.ItemAtY(viewY)
			if blockIdx >= 0 {
				m.mouseDragBlock = blockIdx
				m.mouseDragX = viewX
				m.mouseDragY = blockY
			}
		}

	case tea.MouseReleaseMsg:
		if m.mouseDown {
			m.mouseDown = false
			if m.hasHighlight() {
				text := m.extractHighlight()
				if text != "" {
					cmds = append(cmds, tea.SetClipboard(text))
				}
			} else {
				m.handleBlockClick(m.mouseDownBlock)
			}
		}

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.vp.ScrollBy(-5)
			m.userScrolled = true
			m.vp.UpdateFocusFromScroll()
		case tea.MouseWheelDown:
			m.vp.ScrollBy(5)
			if m.vp.AtBottom() {
				m.userScrolled = false
			}
			m.vp.UpdateFocusFromScroll()
		}

	case msgs.TurnStarted:
		m.userScrolled = false
		m.clearSelection()
		m.rebuildBlocks()
		m.vp.ScrollToBottom()
		m.vp.SetFocusIdx(len(m.blocks) - 1)

	case msgs.AssistantContentUpdated, msgs.StreamCompleted:
		// Forward to rounds first so streaming state is updated
		// before rebuildBlocks reads Blocks().
		for _, r := range m.rounds {
			if cmd := r.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		m.rebuildBlocks()
		if !m.userScrolled {
			m.vp.ScrollToBottom()
			m.vp.SetFocusIdx(len(m.blocks) - 1)
		}
		return tea.Batch(cmds...)
	}

	// Forward to all rounds
	for _, r := range m.rounds {
		if cmd := r.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}
