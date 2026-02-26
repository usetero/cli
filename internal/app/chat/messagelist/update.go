package messagelist

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/usetero/cli/internal/app/chat/msgs"
	appmsg "github.com/usetero/cli/internal/app/msgs"
)

// Update handles messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		decision := reduceKeyPress(msg, m.focused)
		if decision.focusDelta < 0 {
			m.vp.FocusPrev()
		} else if decision.focusDelta > 0 {
			m.vp.FocusNext()
		}
		if decision.scrollDelta != 0 {
			m.vp.ScrollBy(decision.scrollDelta)
			m.vp.UpdateFocusFromScroll()
		}

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			viewX := msg.X - m.originX - outerBorderWidth
			viewY := msg.Y - m.originY
			blockIdx, blockY := m.vp.ItemAtY(viewY)
			m.scope.Debug("click",
				"msgX", msg.X, "msgY", msg.Y,
				"originX", m.originX, "originY", m.originY,
				"viewX", viewX, "viewY", viewY,
				"blockIdx", blockIdx, "blockY", blockY,
				"numBlocks", len(m.blocks),
				"vpHeight", m.height)
			// Log block heights for debugging
			for i, b := range m.blocks {
				rendered := m.renderBlock(b)
				renderedH := lipgloss.Height(rendered)
				reportedH := m.blockHeight(i)
				if renderedH != reportedH {
					m.scope.Warn("height mismatch",
						"blockIdx", i,
						"reported", reportedH,
						"rendered", renderedH,
						"kind", b.block.Kind())
				}
			}
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
			hl := m.hasHighlight()
			m.scope.Debug("release",
				"mouseDownBlock", m.mouseDownBlock,
				"hasHighlight", hl,
				"downX", m.mouseDownX, "downY", m.mouseDownY,
				"dragBlock", m.mouseDragBlock,
				"dragX", m.mouseDragX, "dragY", m.mouseDragY)
			if hl {
				text := m.extractHighlight()
				if text != "" {
					cmds = append(cmds, tea.SetClipboard(text))
					cmds = append(cmds, func() tea.Msg {
						_ = clipboard.WriteAll(text)
						return appmsg.Success{Message: "Copied to clipboard"}
					})
				} else {
					m.scope.Debug("release: empty highlight, treating as click")
					m.handleBlockClick(m.mouseDownBlock, m.mouseDownY)
				}
			} else {
				m.handleBlockClick(m.mouseDownBlock, m.mouseDownY)
			}
		}

	case tea.MouseWheelMsg:
		if delta := reduceMouseWheel(msg.Button); delta != 0 {
			m.vp.ScrollBy(delta)
			m.vp.UpdateFocusFromScroll()
		}

	case msgs.TurnStarted, msgs.AssistantContentUpdated, msgs.StreamCompleted, msgs.StreamFailed:
		decision := reduceLifecycle(msg, m.vp.AtBottom())
		if decision.forwardRounds {
			// Forward to rounds first so streaming state is updated
			// before rebuildBlocks reads Blocks().
			for _, r := range m.rounds {
				if cmd := r.Update(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		if decision.clearSelection {
			m.clearSelection()
		}
		if decision.rebuild {
			m.rebuildBlocks()
		}
		if decision.scrollToBottom {
			m.vp.ScrollToBottom()
		}
		if decision.focusLastAtBottom && len(m.blocks) > 0 {
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
