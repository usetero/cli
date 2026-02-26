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
			target := resolveMouseTarget(msg.X, msg.Y, m.originX, m.originY, m.vp.ItemAtY)
			m.scope.Debug("click",
				"msgX", msg.X, "msgY", msg.Y,
				"originX", m.originX, "originY", m.originY,
				"viewX", target.viewX, "viewY", target.viewY,
				"blockIdx", target.blockIdx, "blockY", target.blockY,
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
			state, decision := reduceSelectionClick(
				m.selectionState(),
				msg.Button,
				selectionPoint{block: target.blockIdx, x: target.viewX, y: target.blockY},
				target.hit,
			)
			m.setSelectionState(state)
			if decision.setFocusIdx {
				m.vp.SetFocusIdx(decision.focusIdx)
			}
		}

	case tea.MouseMotionMsg:
		target := resolveMouseTarget(msg.X, msg.Y, m.originX, m.originY, m.vp.ItemAtY)
		state := reduceSelectionMotion(
			m.selectionState(),
			msg.Button,
			selectionPoint{block: target.blockIdx, x: target.viewX, y: target.blockY},
			target.hit,
		)
		m.setSelectionState(state)

	case tea.MouseReleaseMsg:
		state, decision := reduceSelectionRelease(m.selectionState())
		m.setSelectionState(state)
		if decision.handle {
			hl := m.hasHighlight()
			text := ""
			if hl {
				text = m.extractHighlight()
			}
			action := reduceReleaseAction(hl, text)
			m.scope.Debug("release",
				"mouseDownBlock", m.mouseDownBlock,
				"hasHighlight", hl,
				"action", action,
				"downX", m.mouseDownX, "downY", m.mouseDownY,
				"dragBlock", m.mouseDragBlock,
				"dragX", m.mouseDragX, "dragY", m.mouseDragY)
			switch action {
			case releaseActionCopy:
				cmds = append(cmds, tea.SetClipboard(text))
				cmds = append(cmds, func() tea.Msg {
					_ = clipboard.WriteAll(text)
					return appmsg.Success{Message: "Copied to clipboard"}
				})
			case releaseActionClick:
				m.handleBlockClick(decision.clickBlock, decision.clickY)
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
