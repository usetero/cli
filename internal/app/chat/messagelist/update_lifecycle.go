package messagelist

import tea "charm.land/bubbletea/v2"

func (m *Model) handleLifecycle(msg tea.Msg) ([]tea.Cmd, bool) {
	decision := reduceLifecycle(msg, m.vp.AtBottom())
	if !decision.handle {
		return nil, false
	}

	var cmds []tea.Cmd
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
	return cmds, true
}
