package app

import (
	"fmt"
	"time"
)

const (
	slowUpdateInfoThreshold = 50 * time.Millisecond
	slowUpdateWarnThreshold = 200 * time.Millisecond
	slowRenderInfoThreshold = 50 * time.Millisecond
	slowRenderWarnThreshold = 200 * time.Millisecond
)

func (m *Model) logSlowUpdate(start time.Time, msg any) {
	d := time.Since(start)
	if d < slowUpdateInfoThreshold {
		return
	}

	args := m.perfLogArgs(d)
	args = append(args, "msg_type", fmt.Sprintf("%T", msg))

	if d >= slowUpdateWarnThreshold {
		m.scope.Warn("slow app update", args...)
		return
	}
	m.scope.Info("slow app update", args...)
}

func (m *Model) logSlowRender(start time.Time) {
	d := time.Since(start)
	if d < slowRenderInfoThreshold {
		return
	}

	args := m.perfLogArgs(d)
	if d >= slowRenderWarnThreshold {
		m.scope.Warn("slow app render", args...)
		return
	}
	m.scope.Info("slow app render", args...)
}

func (m *Model) perfLogArgs(d time.Duration) []any {
	return []any{
		"duration_ms", d.Milliseconds(),
		"state", m.stateName(),
		"width", m.width,
		"height", m.height,
		"drawer_open", m.statusBar.IsDrawerOpen(),
		"drawer_tab", m.statusBar.ActiveTabLabel(),
		"palette_open", m.palette != nil,
		"quit_dialog_open", m.quitDlg != nil,
	}
}

func (m *Model) stateName() string {
	switch m.state {
	case stateOnboarding:
		return "onboarding"
	case stateChat:
		return "chat"
	default:
		return "unknown"
	}
}
