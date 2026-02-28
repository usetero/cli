package onboarding

import tea "charm.land/bubbletea/v2"

func (m *Model) runGate(gate Gate) tea.Cmd {
	def := m.definitionForGate(gate)
	return m.setStep(def.newStep(m))
}
