package messagelist

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	appevents "github.com/usetero/cli/internal/app/events"
)

func (m *Model) handleMouseClick(msg tea.MouseClickMsg) {
	if msg.Button != tea.MouseLeft {
		return
	}

	target := resolveMouseTarget(msg.X, msg.Y, m.originX, m.originY, m.vp.ItemAtY)
	m.scope.Debug("click",
		"msgX", msg.X, "msgY", msg.Y,
		"originX", m.originX, "originY", m.originY,
		"viewX", target.viewX, "viewY", target.viewY,
		"blockIdx", target.blockIdx, "blockY", target.blockY,
		"numBlocks", len(m.blocks),
		"vpHeight", m.height)

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

func (m *Model) handleMouseMotion(msg tea.MouseMotionMsg) {
	target := resolveMouseTarget(msg.X, msg.Y, m.originX, m.originY, m.vp.ItemAtY)
	state := reduceSelectionMotion(
		m.selectionState(),
		msg.Button,
		selectionPoint{block: target.blockIdx, x: target.viewX, y: target.blockY},
		target.hit,
	)
	m.setSelectionState(state)
}

func (m *Model) handleMouseRelease(_ tea.MouseReleaseMsg) []tea.Cmd {
	state, decision := reduceSelectionRelease(m.selectionState())
	m.setSelectionState(state)
	if !decision.handle {
		return nil
	}

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
	case releaseActionNoop:
		return nil
	case releaseActionCopy:
		return []tea.Cmd{
			tea.SetClipboard(text),
			func() tea.Msg {
				_ = clipboard.WriteAll(text)
				return appevents.SuccessToastPublished{Message: "Copied to clipboard"}
			},
		}
	case releaseActionClick:
		m.handleBlockClick(decision.clickBlock, decision.clickY)
	}
	return nil
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) {
	if delta := reduceMouseWheel(msg.Button); delta != 0 {
		m.vp.ScrollBy(delta)
		m.vp.UpdateFocusFromScroll()
	}
}
