package messages

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Mouse selection and text highlighting methods for Model.
// This file encapsulates all mouse-based text selection behavior.

// handleMouseDown handles mouse button press events.
// Returns the updated model and true if the event was handled.
func (m Model) handleMouseDown(x, y int) (Model, bool) {
	// Ignore if outside our bounds
	if y < 0 || y >= m.height {
		return m, false
	}

	if len(m.items) == 0 {
		return m, false
	}

	// Find which item was clicked
	itemIdx, itemY := m.itemIndexAtPosition(y)
	if itemIdx < 0 {
		return m, false
	}

	// Check if item is highlightable
	if _, ok := m.items[itemIdx].(Highlightable); !ok {
		return m, false
	}

	// Clear any existing highlight
	m = m.ClearHighlight()

	// Start tracking mouse selection
	m.mouseDown = true
	m.mouseDownIdx = itemIdx
	m.mouseDownX = x
	m.mouseDownY = itemY
	m.mouseDragIdx = itemIdx
	m.mouseDragX = x
	m.mouseDragY = itemY

	return m, true
}

// handleMouseDrag handles mouse motion events during a drag.
func (m Model) handleMouseDrag(x, y int) Model {
	if !m.mouseDown || len(m.items) == 0 {
		return m
	}

	// Find item at current position
	itemIdx, itemY := m.itemIndexAtPosition(y)
	if itemIdx < 0 {
		return m
	}

	// Update drag position
	m.mouseDragIdx = itemIdx
	m.mouseDragX = x
	m.mouseDragY = itemY

	// Update highlight on affected items
	m = m.updateHighlight()

	return m
}

// handleMouseUp handles mouse button release events.
// Returns the updated model and a command to copy highlighted content.
func (m Model) handleMouseUp(x, y int) (Model, tea.Cmd) {
	if !m.mouseDown {
		return m, nil
	}

	m.mouseDown = false

	// If we have a highlight, copy to clipboard
	if m.HasHighlight() {
		content := m.HighlightContent()
		if content != "" {
			// Keep the highlight visible, user can press Esc to clear
			return m, tea.SetClipboard(content)
		}
	}

	return m, nil
}

// itemIndexAtPosition finds the item at the given viewport-relative y coordinate.
// Returns the item index and the y offset within that item, or -1, -1 if not found.
func (m Model) itemIndexAtPosition(y int) (itemIdx int, itemY int) {
	if y < 0 || y >= m.height || len(m.items) == 0 {
		return -1, -1
	}

	// Walk through visible items to find which one contains this y
	currentIdx := m.offsetIdx
	currentLine := -m.offsetLine // Negative because offsetLine is lines hidden above

	for currentIdx < len(m.items) && currentLine < m.height {
		item := m.items[currentIdx]
		itemHeight := item.Height(m.width)
		itemEndLine := currentLine + itemHeight

		// Check if y is within this item's visible range
		if y >= currentLine && y < itemEndLine {
			// Found the item, calculate itemY (offset within the item)
			itemY = y - currentLine
			return currentIdx, itemY
		}

		// Move to next item (including gap)
		currentLine = itemEndLine
		if m.gap > 0 && currentIdx < len(m.items)-1 {
			currentLine += m.gap
		}
		currentIdx++
	}

	return -1, -1
}

// updateHighlight updates the highlight on items based on current mouse selection.
func (m Model) updateHighlight() Model {
	if m.mouseDownIdx < 0 {
		return m
	}

	// Determine selection direction
	startIdx, startLine, startCol := m.mouseDownIdx, m.mouseDownY, m.mouseDownX
	endIdx, endLine, endCol := m.mouseDragIdx, m.mouseDragY, m.mouseDragX

	// Normalize: ensure start is before end
	if endIdx < startIdx || (endIdx == startIdx && (endLine < startLine || (endLine == startLine && endCol < startCol))) {
		startIdx, endIdx = endIdx, startIdx
		startLine, endLine = endLine, startLine
		startCol, endCol = endCol, startCol
	}

	// Clear highlights on all items first
	for i := range m.items {
		if h, ok := m.items[i].(Highlightable); ok {
			h.ClearHighlight()
		}
	}

	// Apply highlight to items in the selection range
	for i := startIdx; i <= endIdx && i < len(m.items); i++ {
		h, ok := m.items[i].(Highlightable)
		if !ok {
			continue
		}

		var sl, sc, el, ec int

		if i == startIdx && i == endIdx {
			// Single item selection
			sl, sc = startLine, startCol
			el, ec = endLine, endCol
		} else if i == startIdx {
			// First item in multi-item selection
			sl, sc = startLine, startCol
			el, ec = -1, -1 // To end of item
		} else if i == endIdx {
			// Last item in multi-item selection
			sl, sc = 0, 0
			el, ec = endLine, endCol
		} else {
			// Middle item - select all
			sl, sc = 0, 0
			el, ec = -1, -1
		}

		h.SetHighlight(sl, sc, el, ec)
	}

	return m
}

// HasHighlight returns true if any item has a highlight.
func (m Model) HasHighlight() bool {
	for _, item := range m.items {
		if h, ok := item.(Highlightable); ok && h.IsHighlighted() {
			return true
		}
	}
	return false
}

// HighlightContent returns the combined highlighted content from all items.
func (m Model) HighlightContent() string {
	var parts []string
	for _, item := range m.items {
		if h, ok := item.(Highlightable); ok && h.IsHighlighted() {
			content := h.HighlightedContent()
			if content != "" {
				parts = append(parts, content)
			}
		}
	}
	return strings.Join(parts, "")
}

// ClearHighlight removes highlights from all items.
func (m Model) ClearHighlight() Model {
	for _, item := range m.items {
		if h, ok := item.(Highlightable); ok {
			h.ClearHighlight()
		}
	}
	m.mouseDown = false
	m.mouseDownIdx = -1
	return m
}
