// Package viewport provides item-aware scrolling over a list of
// variable-height items separated by gaps.
//
// It is pure math — no rendering, no external dependencies.
// The consumer provides item heights and gap sizes as int slices,
// and the viewport handles scroll position, focus tracking,
// hit testing, and bottom detection.
package viewport

// Model is an item-aware viewport that tracks scroll position,
// focus, and layout over a list of variable-height items.
type Model struct {
	heights []int // height of each item in lines
	gaps    []int // gap before each item: gaps[0]=0, gaps[i]=lines between item i-1 and i

	height         int // viewport height in lines
	trailingHeight int // extra content below last item (e.g. divider)

	offsetIdx  int // index of first visible item
	offsetLine int // lines scrolled into that item (>= 0)

	focusIdx int // focused item (-1 = none)

	// Cached bottom position
	bottomIdx   int
	bottomLine  int
	bottomValid bool
}

// New returns a viewport with no items.
func New() Model {
	return Model{focusIdx: -1}
}

// --- Consumer sets these ---

// SetItems replaces the item heights and gaps.
// gaps[i] is the number of lines between item i-1 and item i.
// gaps[0] must be 0 (no gap before the first item).
// Both slices must have the same length.
// Invalidates the bottom cache and clamps scroll/focus state.
func (m *Model) SetItems(heights, gaps []int) {
	m.heights = heights
	m.gaps = gaps
	m.bottomValid = false

	n := len(heights)
	if n == 0 {
		m.offsetIdx = 0
		m.offsetLine = 0
		m.focusIdx = -1
		return
	}

	if m.offsetIdx >= n {
		m.offsetIdx = n - 1
		m.offsetLine = 0
	}
	if m.focusIdx >= n {
		m.focusIdx = n - 1
	}
}

// SetHeight sets the viewport height in lines.
func (m *Model) SetHeight(h int) {
	m.height = h
	m.bottomValid = false
}

// SetTrailingHeight sets extra content height below the last item
// (e.g. a divider or status line). The viewport accounts for this
// when computing the bottom scroll position.
func (m *Model) SetTrailingHeight(h int) {
	m.trailingHeight = h
	m.bottomValid = false
}

// Len returns the number of items.
func (m *Model) Len() int {
	return len(m.heights)
}

// --- Scroll ---

// ScrollBy scrolls by the given number of lines (positive = down, negative = up).
func (m *Model) ScrollBy(lines int) {
	if len(m.heights) == 0 || lines == 0 {
		return
	}
	if lines > 0 {
		m.scrollDown(lines)
	} else {
		m.scrollUp(-lines)
	}
}

func (m *Model) scrollDown(lines int) {
	m.offsetLine += lines

	for m.offsetIdx < len(m.heights) {
		bh := m.itemHeight(m.offsetIdx)
		gap := m.gapAfter(m.offsetIdx)
		total := bh + gap

		if m.offsetLine < total {
			break
		}

		m.offsetLine -= total
		m.offsetIdx++

		if m.offsetIdx >= len(m.heights) {
			m.ScrollToBottom()
			return
		}
	}

	// Clamp to bottom
	lastIdx, lastLine := m.bottom()
	if m.offsetIdx > lastIdx || (m.offsetIdx == lastIdx && m.offsetLine > lastLine) {
		m.offsetIdx = lastIdx
		m.offsetLine = lastLine
	}
}

func (m *Model) scrollUp(lines int) {
	m.offsetLine -= lines

	for m.offsetLine < 0 {
		if m.offsetIdx <= 0 {
			m.ScrollToTop()
			return
		}
		m.offsetIdx--
		bh := m.itemHeight(m.offsetIdx)
		gap := m.gapAfter(m.offsetIdx)
		m.offsetLine += bh + gap
	}
}

// ScrollToTop scrolls to the first item.
func (m *Model) ScrollToTop() {
	m.offsetIdx = 0
	m.offsetLine = 0
}

// ScrollToBottom scrolls so the last item is at the bottom of the viewport.
func (m *Model) ScrollToBottom() {
	if len(m.heights) == 0 {
		return
	}
	m.offsetIdx, m.offsetLine = m.bottom()
}

// AtBottom returns whether the viewport is showing the bottom.
func (m *Model) AtBottom() bool {
	if len(m.heights) == 0 {
		return true
	}
	idx, line := m.bottom()
	return m.offsetIdx > idx || (m.offsetIdx == idx && m.offsetLine >= line)
}

// Offset returns the current scroll position: the index of the first
// visible item and the number of lines scrolled into it.
func (m *Model) Offset() (idx, line int) {
	return m.offsetIdx, m.offsetLine
}

// --- Focus ---

// FocusIdx returns the focused item index (-1 if none).
func (m *Model) FocusIdx() int {
	return m.focusIdx
}

// SetFocusIdx sets the focused item index.
func (m *Model) SetFocusIdx(idx int) {
	m.focusIdx = idx
}

// FocusPrev moves focus to the previous item and scrolls to keep it visible.
func (m *Model) FocusPrev() {
	if len(m.heights) == 0 {
		return
	}
	if m.focusIdx <= 0 {
		m.focusIdx = 0
	} else {
		m.focusIdx--
	}
	m.scrollToFocused()
}

// FocusNext moves focus to the next item and scrolls to keep it visible.
func (m *Model) FocusNext() {
	if len(m.heights) == 0 {
		return
	}
	if m.focusIdx < 0 {
		m.focusIdx = 0
	} else if m.focusIdx < len(m.heights)-1 {
		m.focusIdx++
	}
	m.scrollToFocused()
}

// UpdateFocusFromScroll sets focus to the first fully visible item.
func (m *Model) UpdateFocusFromScroll() {
	if len(m.heights) == 0 {
		m.focusIdx = -1
		return
	}
	// If offsetLine > 0, the first item is partially hidden — focus the next one.
	if m.offsetLine > 0 && m.offsetIdx+1 < len(m.heights) {
		m.focusIdx = m.offsetIdx + 1
	} else {
		m.focusIdx = m.offsetIdx
	}
}

// scrollToFocused ensures the focused item is visible in the viewport.
func (m *Model) scrollToFocused() {
	if m.focusIdx < 0 || m.focusIdx >= len(m.heights) {
		return
	}

	// Focused item is before the visible area — scroll up to it
	if m.focusIdx < m.offsetIdx {
		m.offsetIdx = m.focusIdx
		m.offsetLine = 0
		return
	}

	// Focused item is at the offset but partially hidden — show it fully
	if m.focusIdx == m.offsetIdx && m.offsetLine > 0 {
		m.offsetLine = 0
		return
	}

	// Check if focused item is below the visible area.
	y := -m.offsetLine
	for idx := m.offsetIdx; idx <= m.focusIdx && idx < len(m.heights); idx++ {
		if idx > m.offsetIdx {
			y += m.gaps[idx]
		}
		bh := m.itemHeight(idx)
		if idx == m.focusIdx {
			if y+bh > m.height {
				m.offsetIdx = m.focusIdx
				m.offsetLine = 0
			}
			return
		}
		y += bh
	}
}

// --- Hit testing ---

// ItemAtY maps a viewport-relative Y coordinate to an item index and
// the Y offset within that item. Returns (-1, -1) if the position is
// in a gap, above, or below all items.
func (m *Model) ItemAtY(y int) (idx, localY int) {
	if y < 0 || y >= m.height || len(m.heights) == 0 {
		return -1, -1
	}

	currentLine := -m.offsetLine
	for i := m.offsetIdx; i < len(m.heights); i++ {
		// Gap before this item
		if i > m.offsetIdx {
			gap := m.gaps[i]
			if y >= currentLine && y < currentLine+gap {
				return -1, -1 // in a gap
			}
			currentLine += gap
		}

		bh := m.itemHeight(i)
		if y >= currentLine && y < currentLine+bh {
			return i, y - currentLine
		}

		currentLine += bh
		if currentLine >= m.height {
			break
		}
	}

	return -1, -1
}

// --- Internal helpers ---

// itemHeight returns the height of item at idx, with a minimum of 1.
func (m *Model) itemHeight(idx int) int {
	h := m.heights[idx]
	if h < 1 {
		return 1
	}
	return h
}

// gapAfter returns the gap that follows item at idx.
// This is gaps[idx+1] if idx+1 is in range, otherwise 0.
func (m *Model) gapAfter(idx int) int {
	next := idx + 1
	if next >= len(m.gaps) {
		return 0
	}
	return m.gaps[next]
}

// bottom returns the cached bottom scroll position.
func (m *Model) bottom() (int, int) {
	if !m.bottomValid {
		m.bottomIdx, m.bottomLine = m.computeBottom()
		m.bottomValid = true
	}
	return m.bottomIdx, m.bottomLine
}

// computeBottom calculates the scroll position that shows the bottom of content.
// Walks backward from the last item, accumulating heights using the same
// gap-after attribution that scrollDown uses.
func (m *Model) computeBottom() (int, int) {
	remaining := m.height - m.trailingHeight

	for idx := len(m.heights) - 1; idx >= 0; idx-- {
		bh := m.itemHeight(idx)
		gap := m.gapAfter(idx)

		remaining -= bh + gap
		if remaining <= 0 {
			return idx, -remaining
		}
	}

	// All content fits in viewport
	return 0, 0
}
