package messagelist

import (
	"image"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/tea/highlight"
)

func (m *Model) selectionState() selectionState {
	return selectionState{
		mouseDown:      m.mouseDown,
		mouseDownBlock: m.mouseDownBlock,
		mouseDownX:     m.mouseDownX,
		mouseDownY:     m.mouseDownY,
		mouseDragBlock: m.mouseDragBlock,
		mouseDragX:     m.mouseDragX,
		mouseDragY:     m.mouseDragY,
	}
}

func (m *Model) setSelectionState(state selectionState) {
	m.mouseDown = state.mouseDown
	m.mouseDownBlock = state.mouseDownBlock
	m.mouseDownX = state.mouseDownX
	m.mouseDownY = state.mouseDownY
	m.mouseDragBlock = state.mouseDragBlock
	m.mouseDragX = state.mouseDragX
	m.mouseDragY = state.mouseDragY
}

// --- Click handling ---

// handleBlockClick handles a click on a specific block.
// If the block implements Toggleable, it toggles and we rebuild
// the viewport items (heights may have changed).
func (m *Model) handleBlockClick(idx, y int) {
	if idx < 0 || idx >= len(m.blocks) {
		return
	}
	if t, ok := m.blocks[idx].block.(block.Toggleable); ok {
		t.Toggle(y)
		m.syncViewportItems()
	}
}

// --- Highlight / selection ---

// hasHighlight returns whether there is a current text selection.
func (m *Model) hasHighlight() bool {
	state := m.selectionState()
	if state.mouseDownBlock < 0 || state.mouseDragBlock < 0 {
		return false
	}
	return state.mouseDownBlock != state.mouseDragBlock ||
		state.mouseDownY != state.mouseDragY ||
		state.mouseDownX != state.mouseDragX
}

// getHighlightRange returns the normalized selection range (start <= end).
func (m *Model) getHighlightRange() (startBlock, startLine, startCol, endBlock, endLine, endCol int) {
	if m.mouseDownBlock < 0 {
		return -1, -1, -1, -1, -1, -1
	}

	draggingForward := m.mouseDragBlock > m.mouseDownBlock ||
		(m.mouseDragBlock == m.mouseDownBlock && m.mouseDragY > m.mouseDownY) ||
		(m.mouseDragBlock == m.mouseDownBlock && m.mouseDragY == m.mouseDownY && m.mouseDragX >= m.mouseDownX)

	if draggingForward {
		return m.mouseDownBlock, m.mouseDownY, m.mouseDownX,
			m.mouseDragBlock, m.mouseDragY, m.mouseDragX
	}
	return m.mouseDragBlock, m.mouseDragY, m.mouseDragX,
		m.mouseDownBlock, m.mouseDownY, m.mouseDownX
}

// blockHighlightRange returns the highlight coordinates for a single block,
// given its index and the overall selection range. Returns (-1,-1,-1,-1) if
// the block is not in the selection.
func blockHighlightRange(idx, startBlock, startLine, startCol, endBlock, endLine, endCol int) (sLine, sCol, eLine, eCol int) {
	if idx < startBlock || idx > endBlock {
		return -1, -1, -1, -1
	}
	if idx == startBlock && idx == endBlock {
		return startLine, startCol, endLine, endCol
	}
	if idx == startBlock {
		return startLine, startCol, -1, -1 // to end of block
	}
	if idx == endBlock {
		return 0, 0, endLine, endCol
	}
	return 0, 0, -1, -1 // fully highlighted
}

// extractHighlight returns the plain text of the current selection.
func (m *Model) extractHighlight() string {
	startBlock, startLine, startCol, endBlock, endLine, endCol := m.getHighlightRange()
	if startBlock < 0 {
		return ""
	}

	var sb strings.Builder
	for i := startBlock; i <= endBlock && i < len(m.blocks); i++ {
		sLine, sCol, eLine, eCol := blockHighlightRange(i, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sLine < 0 {
			continue
		}

		// Extract from the block's own View() — the content before
		// renderBlock wraps it with border/padding decorations.
		content := m.blocks[i].block.View()
		w := lipgloss.Width(content)
		h := lipgloss.Height(content)
		area := image.Rect(0, 0, w, h)

		text := highlight.Extract(content, area, sLine, sCol, eLine, eCol)
		if text != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(strings.TrimRight(text, "\n"))
		}
	}

	return sb.String()
}

// clearSelection resets mouse selection state.
func (m *Model) clearSelection() {
	m.mouseDown = false
	m.mouseDownBlock = -1
	m.mouseDragBlock = -1
}
