package messagelist

import tea "charm.land/bubbletea/v2"

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) {
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
}
