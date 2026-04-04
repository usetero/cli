package statusbar

import "charm.land/lipgloss/v2"

func (m *Model) leftCluster() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.parts.sync.Segment(),
		" ",
		m.renderBrand(),
	)
}

func (m *Model) leftContent() string {
	segments := make([]string, 0, 2)
	if scope := m.parts.scope.Segment(); scope != "" {
		segments = append(segments, scope)
	}
	if estate := m.parts.estate.Segment(); estate != "" {
		segments = append(segments, estate)
	}
	return m.joinWithSlash(segments)
}

func (m *Model) leftSection() string {
	leftCluster := m.leftCluster()
	segments := make([]string, 0, 4)
	if scope := m.parts.scope.Segment(); scope != "" {
		segments = append(segments, scope)
	}
	if estate := m.parts.estate.Segment(); estate != "" {
		segments = append(segments, estate)
	}
	if waste := m.parts.pressure.Segment(); waste != "" {
		segments = append(segments, waste)
	}
	if spikes := m.parts.alerts.Segment(); spikes != "" {
		segments = append(segments, spikes)
	}
	content := m.joinWithSlash(segments)
	if content == "" {
		return leftCluster
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftCluster,
		m.renderSeparator(),
		content,
	)
}

func (m *Model) joinWithSlash(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	joined := segments[0]
	for _, segment := range segments[1:] {
		joined = lipgloss.JoinHorizontal(lipgloss.Left, joined, m.renderSeparator(), segment)
	}
	return joined
}
