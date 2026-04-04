package input

func (m *Model) CapturingInput() bool {
	return !m.Empty()
}
