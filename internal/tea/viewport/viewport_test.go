package viewport

import "testing"

// Helper to create a viewport with uniform items and gaps.
// n items, each of height h, with gap g between them.
func uniform(n, h, g, vpHeight int) *Model {
	heights := make([]int, n)
	gaps := make([]int, n)
	for i := range n {
		heights[i] = h
		if i > 0 {
			gaps[i] = g
		}
	}
	m := New()
	m.SetHeight(vpHeight)
	m.SetItems(heights, gaps)
	return &m
}

// Helper to create a viewport with custom item heights, uniform gap.
func custom(heights []int, g, vpHeight int) *Model {
	gaps := make([]int, len(heights))
	for i := range heights {
		if i > 0 {
			gaps[i] = g
		}
	}
	m := New()
	m.SetHeight(vpHeight)
	m.SetItems(heights, gaps)
	return &m
}

func TestNew(t *testing.T) {
	t.Parallel()

	m := New()
	if m.Len() != 0 {
		t.Errorf("Len = %d, want 0", m.Len())
	}
	if m.FocusIdx() != -1 {
		t.Errorf("FocusIdx = %d, want -1", m.FocusIdx())
	}
	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0)", idx, line)
	}
}

func TestScrollBy_Empty(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetHeight(10)
	m.ScrollBy(5)
	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0)", idx, line)
	}
}

func TestAtBottom_Empty(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetHeight(10)
	if !m.AtBottom() {
		t.Error("AtBottom = false, want true for empty viewport")
	}
}

// --- ScrollDown ---

func TestScrollDown_WithinFirstItem(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 10
	m := uniform(3, 5, 1, 10)
	m.ScrollBy(2)

	idx, line := m.Offset()
	if idx != 0 || line != 2 {
		t.Errorf("Offset = (%d, %d), want (0, 2)", idx, line)
	}
}

func TestScrollDown_AcrossItemBoundary(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 10
	// Item 0: 5 lines + 1 gap = 6 total. Scroll 7 → into item 1, line 1.
	m := uniform(3, 5, 1, 10)
	m.ScrollBy(7)

	idx, line := m.Offset()
	if idx != 1 || line != 1 {
		t.Errorf("Offset = (%d, %d), want (1, 1)", idx, line)
	}
}

func TestScrollDown_AcrossMultipleItems(t *testing.T) {
	t.Parallel()

	// 5 items of height 3, gap 1, viewport 10
	// Item unit = 3 + 1 = 4. Scroll 9 → skip 2 items (8 lines), into item 2 line 1.
	m := uniform(5, 3, 1, 10)
	m.ScrollBy(9)

	idx, line := m.Offset()
	if idx != 2 || line != 1 {
		t.Errorf("Offset = (%d, %d), want (2, 1)", idx, line)
	}
}

func TestScrollDown_ClampsToBottom(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 10
	// Total content = 5+1+5+1+5 = 17. Scroll way past.
	m := uniform(3, 5, 1, 10)
	m.ScrollBy(100)

	if !m.AtBottom() {
		t.Error("should be at bottom after overshooting")
	}
}

func TestScrollDown_ExactBottom(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 10
	m := uniform(3, 5, 1, 10)
	m.ScrollToBottom()
	bottomIdx, bottomLine := m.Offset()

	m.ScrollToTop()
	m.ScrollBy(100)

	idx, line := m.Offset()
	if idx != bottomIdx || line != bottomLine {
		t.Errorf("ScrollBy(100) = (%d, %d), ScrollToBottom = (%d, %d)", idx, line, bottomIdx, bottomLine)
	}
}

// --- ScrollUp ---

func TestScrollUp_WithinFirstItem(t *testing.T) {
	t.Parallel()

	m := uniform(3, 5, 1, 10)
	m.ScrollBy(3)
	m.ScrollBy(-1)

	idx, line := m.Offset()
	if idx != 0 || line != 2 {
		t.Errorf("Offset = (%d, %d), want (0, 2)", idx, line)
	}
}

func TestScrollUp_AcrossItemBoundary(t *testing.T) {
	t.Parallel()

	// Start at item 1, line 1. Scroll up 3 → should end up in item 0.
	// Item 0: height 5 + gap 1 = 6. offsetLine = 1 - 3 = -2 → go to item 0, offsetLine = 6-2 = 4.
	m := uniform(3, 5, 1, 10)
	m.ScrollBy(7) // item 1, line 1
	m.ScrollBy(-3)

	idx, line := m.Offset()
	if idx != 0 || line != 4 {
		t.Errorf("Offset = (%d, %d), want (0, 4)", idx, line)
	}
}

func TestScrollUp_ClampsToTop(t *testing.T) {
	t.Parallel()

	m := uniform(3, 5, 1, 10)
	m.ScrollBy(5)
	m.ScrollBy(-100)

	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0)", idx, line)
	}
}

func TestScrollRoundTrip(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)
	m.ScrollBy(7)
	idx1, line1 := m.Offset()
	m.ScrollBy(-7)

	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("after roundtrip: Offset = (%d, %d), want (0, 0); intermediate was (%d, %d)", idx, line, idx1, line1)
	}
}

// --- ScrollToBottom / AtBottom ---

func TestScrollToBottom(t *testing.T) {
	t.Parallel()

	m := uniform(3, 5, 1, 10)
	m.ScrollToBottom()

	if !m.AtBottom() {
		t.Error("AtBottom = false after ScrollToBottom")
	}
}

func TestAtBottom_InitiallyFitsInViewport(t *testing.T) {
	t.Parallel()

	// 2 items of height 3, gap 1, viewport 20 → total content 7, fits easily
	m := uniform(2, 3, 1, 20)
	if !m.AtBottom() {
		t.Error("AtBottom = false but content fits in viewport")
	}
}

func TestAtBottom_FalseWhenScrolledUp(t *testing.T) {
	t.Parallel()

	// Make content taller than viewport
	m := uniform(10, 5, 1, 10)
	if m.AtBottom() {
		t.Error("AtBottom = true at top of long content")
	}
}

func TestScrollToBottom_WithTrailingHeight(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 10, trailing 2
	m := uniform(3, 5, 1, 10)
	m.SetTrailingHeight(2)
	m.ScrollToBottom()

	if !m.AtBottom() {
		t.Error("AtBottom = false after ScrollToBottom with trailing height")
	}

	// Should be scrolled further than without trailing
	idx1, line1 := m.Offset()

	m.SetTrailingHeight(0)
	m.ScrollToBottom()
	idx2, line2 := m.Offset()

	if idx1 < idx2 || (idx1 == idx2 && line1 <= line2) {
		t.Errorf("trailing height should scroll further: with=(%d,%d) without=(%d,%d)", idx1, line1, idx2, line2)
	}
}

// --- Bottom computation ---

func TestBottom_SingleItemFitsExactly(t *testing.T) {
	t.Parallel()

	// 1 item of height 10, viewport 10 → bottom is (0, 0)
	m := uniform(1, 10, 0, 10)
	m.ScrollToBottom()

	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0)", idx, line)
	}
}

func TestBottom_SingleItemTallerThanViewport(t *testing.T) {
	t.Parallel()

	// 1 item of height 20, viewport 10 → bottom is (0, 10)
	m := uniform(1, 20, 0, 10)
	m.ScrollToBottom()

	idx, line := m.Offset()
	if idx != 0 || line != 10 {
		t.Errorf("Offset = (%d, %d), want (0, 10)", idx, line)
	}
}

func TestBottom_VariableHeights(t *testing.T) {
	t.Parallel()

	// Items: [3, 7, 2], gap 1, viewport 10
	// Total from end: item 2 (2) + gap (1) + item 1 (7) = 10. Exactly fills viewport starting at item 1.
	m := custom([]int{3, 7, 2}, 1, 10)
	m.ScrollToBottom()

	idx, line := m.Offset()
	if idx != 1 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (1, 0)", idx, line)
	}
}

// --- SetItems clamping ---

func TestSetItems_ClampsScrollPosition(t *testing.T) {
	t.Parallel()

	m := uniform(10, 5, 1, 10)
	m.ScrollToBottom()

	// Shrink to 2 items — scroll must clamp
	m.SetItems([]int{5, 5}, []int{0, 1})
	idx, _ := m.Offset()
	if idx >= 2 {
		t.Errorf("offsetIdx = %d, should be < 2 after shrinking to 2 items", idx)
	}
}

func TestSetItems_ClampsFocusIdx(t *testing.T) {
	t.Parallel()

	m := uniform(10, 5, 1, 10)
	m.SetFocusIdx(9)

	m.SetItems([]int{5, 5}, []int{0, 1})
	if m.FocusIdx() >= 2 {
		t.Errorf("focusIdx = %d, should be < 2 after shrinking", m.FocusIdx())
	}
}

func TestSetItems_Empty(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)
	m.ScrollBy(5)
	m.SetFocusIdx(3)

	m.SetItems(nil, nil)
	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0) after clearing items", idx, line)
	}
	if m.FocusIdx() != -1 {
		t.Errorf("FocusIdx = %d, want -1 after clearing items", m.FocusIdx())
	}
}

// --- Focus ---

func TestFocusPrev(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 20)
	m.SetFocusIdx(3)
	m.FocusPrev()

	if m.FocusIdx() != 2 {
		t.Errorf("FocusIdx = %d, want 2", m.FocusIdx())
	}
}

func TestFocusPrev_ClampsAtZero(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 20)
	m.SetFocusIdx(0)
	m.FocusPrev()

	if m.FocusIdx() != 0 {
		t.Errorf("FocusIdx = %d, want 0", m.FocusIdx())
	}
}

func TestFocusNext(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 20)
	m.SetFocusIdx(1)
	m.FocusNext()

	if m.FocusIdx() != 2 {
		t.Errorf("FocusIdx = %d, want 2", m.FocusIdx())
	}
}

func TestFocusNext_ClampsAtEnd(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 20)
	m.SetFocusIdx(4)
	m.FocusNext()

	if m.FocusIdx() != 4 {
		t.Errorf("FocusIdx = %d, want 4", m.FocusIdx())
	}
}

func TestFocusNext_FromNegative(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 20)
	// focusIdx starts at -1 from New()
	m.FocusNext()

	if m.FocusIdx() != 0 {
		t.Errorf("FocusIdx = %d, want 0", m.FocusIdx())
	}
}

func TestFocusPrev_ScrollsUp(t *testing.T) {
	t.Parallel()

	// 10 items of height 5, gap 1, viewport 10
	// Scroll down far, set focus to item 5, then FocusPrev.
	m := uniform(10, 5, 1, 10)
	m.ScrollBy(30) // scroll deep
	m.SetFocusIdx(5)
	m.FocusPrev() // focus moves to 4, should scroll to show it

	idx, _ := m.Offset()
	if idx > 4 {
		t.Errorf("offsetIdx = %d, should be <= 4 to show focused item", idx)
	}
}

func TestFocusNext_ScrollsDown(t *testing.T) {
	t.Parallel()

	// 10 items of height 5, gap 1, viewport 10
	// Focus item 1, then FocusNext until item is below viewport — should auto-scroll.
	m := uniform(10, 5, 1, 10)
	m.SetFocusIdx(0)

	// Item 0: 5 lines, item 1: gap(1) + 5 = visible at line 6, fits in viewport of 10.
	// Item 2: would start at line 12, doesn't fit. FocusNext should scroll.
	m.FocusNext() // focus = 1
	m.FocusNext() // focus = 2, should scroll

	idx, _ := m.Offset()
	if m.FocusIdx() != 2 {
		t.Errorf("FocusIdx = %d, want 2", m.FocusIdx())
	}
	// Item 2 should be visible — offsetIdx should be <= 2
	if idx > 2 {
		t.Errorf("offsetIdx = %d, should be <= 2", idx)
	}
}

func TestUpdateFocusFromScroll_FullyVisible(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)
	m.UpdateFocusFromScroll()

	// Items: 0(3)+gap(1)+1(3)+gap(1)+2(3)... viewport=10.
	// Item 0: y=0..2, item 1: y=4..6, item 2: y=8..10 (doesn't fit, 11>10).
	// Last fully visible = 1.
	if m.FocusIdx() != 1 {
		t.Errorf("FocusIdx = %d, want 1 (last fully visible)", m.FocusIdx())
	}
}

func TestUpdateFocusFromScroll_PartiallyHidden(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)
	m.ScrollBy(1) // partially into item 0

	m.UpdateFocusFromScroll()

	// After scrolling 1 line: item 0 partially hidden, items 1,2 fully visible. Last = 2.
	if m.FocusIdx() != 2 {
		t.Errorf("FocusIdx = %d, want 2 (last fully visible)", m.FocusIdx())
	}
}

func TestUpdateFocusFromScroll_Empty(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetHeight(10)
	m.UpdateFocusFromScroll()

	if m.FocusIdx() != -1 {
		t.Errorf("FocusIdx = %d, want -1", m.FocusIdx())
	}
}

// --- Hit testing ---

func TestItemAtY_FirstItem(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 20
	m := uniform(3, 5, 1, 20)

	idx, localY := m.ItemAtY(0)
	if idx != 0 || localY != 0 {
		t.Errorf("ItemAtY(0) = (%d, %d), want (0, 0)", idx, localY)
	}

	idx, localY = m.ItemAtY(4)
	if idx != 0 || localY != 4 {
		t.Errorf("ItemAtY(4) = (%d, %d), want (0, 4)", idx, localY)
	}
}

func TestItemAtY_Gap(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 20
	// Gap between item 0 and 1 is at Y=5 (after 5 lines of item 0)
	m := uniform(3, 5, 1, 20)

	idx, _ := m.ItemAtY(5)
	if idx != -1 {
		t.Errorf("ItemAtY(5) = %d, want -1 (gap)", idx)
	}
}

func TestItemAtY_SecondItem(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 1, viewport 20
	// Item 1 starts at Y=6 (5 lines of item 0 + 1 gap)
	m := uniform(3, 5, 1, 20)

	idx, localY := m.ItemAtY(6)
	if idx != 1 || localY != 0 {
		t.Errorf("ItemAtY(6) = (%d, %d), want (1, 0)", idx, localY)
	}
}

func TestItemAtY_WithOffset(t *testing.T) {
	t.Parallel()

	// 5 items of height 5, gap 1, viewport 10
	// Scroll down 2 lines into item 0
	m := uniform(5, 5, 1, 10)
	m.ScrollBy(2)

	// Y=0 should be item 0, local line 2 (since we scrolled 2 into it)
	idx, localY := m.ItemAtY(0)
	if idx != 0 || localY != 2 {
		t.Errorf("ItemAtY(0) = (%d, %d), want (0, 2)", idx, localY)
	}
}

func TestItemAtY_OutOfBounds(t *testing.T) {
	t.Parallel()

	m := uniform(3, 5, 1, 20)

	idx, _ := m.ItemAtY(-1)
	if idx != -1 {
		t.Errorf("ItemAtY(-1) = %d, want -1", idx)
	}

	idx, _ = m.ItemAtY(20)
	if idx != -1 {
		t.Errorf("ItemAtY(20) = %d, want -1", idx)
	}
}

func TestItemAtY_Empty(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetHeight(10)

	idx, _ := m.ItemAtY(0)
	if idx != -1 {
		t.Errorf("ItemAtY(0) = %d, want -1 for empty viewport", idx)
	}
}

func TestItemAtY_NoGap(t *testing.T) {
	t.Parallel()

	// 3 items of height 5, gap 0, viewport 20
	m := uniform(3, 5, 0, 20)

	// Item 1 starts immediately at Y=5
	idx, localY := m.ItemAtY(5)
	if idx != 1 || localY != 0 {
		t.Errorf("ItemAtY(5) = (%d, %d), want (1, 0)", idx, localY)
	}
}

// --- Zero-height items ---

func TestZeroHeightItem_TreatedAsOne(t *testing.T) {
	t.Parallel()

	m := custom([]int{0, 5}, 1, 20)

	// Item 0 has height 0 but should be treated as 1
	idx, localY := m.ItemAtY(0)
	if idx != 0 || localY != 0 {
		t.Errorf("ItemAtY(0) = (%d, %d), want (0, 0)", idx, localY)
	}

	// Gap at Y=1, item 1 at Y=2
	idx, _ = m.ItemAtY(1)
	if idx != -1 {
		t.Errorf("ItemAtY(1) = %d, want -1 (gap)", idx)
	}

	idx, localY = m.ItemAtY(2)
	if idx != 1 || localY != 0 {
		t.Errorf("ItemAtY(2) = (%d, %d), want (1, 0)", idx, localY)
	}
}

// --- Variable gaps ---

func TestVariableGaps(t *testing.T) {
	t.Parallel()

	// 3 items of height 3, gaps: [0, 1, 4], viewport 20
	m := New()
	m.SetHeight(20)
	m.SetItems([]int{3, 3, 3}, []int{0, 1, 4})

	// Item 0: Y=0..2
	// Gap: Y=3 (size 1)
	// Item 1: Y=4..6
	// Gap: Y=7..10 (size 4)
	// Item 2: Y=11..13

	idx, _ := m.ItemAtY(3)
	if idx != -1 {
		t.Errorf("ItemAtY(3) = %d, want -1 (gap)", idx)
	}

	idx, localY := m.ItemAtY(4)
	if idx != 1 || localY != 0 {
		t.Errorf("ItemAtY(4) = (%d, %d), want (1, 0)", idx, localY)
	}

	idx, _ = m.ItemAtY(9)
	if idx != -1 {
		t.Errorf("ItemAtY(9) = %d, want -1 (in 4-line gap)", idx)
	}

	idx, localY = m.ItemAtY(11)
	if idx != 2 || localY != 0 {
		t.Errorf("ItemAtY(11) = (%d, %d), want (2, 0)", idx, localY)
	}
}

// --- Scroll consistency ---

func TestScrollToBottom_ThenScrollToTop_Roundtrip(t *testing.T) {
	t.Parallel()

	m := uniform(20, 5, 1, 10)
	m.ScrollToBottom()
	m.ScrollToTop()

	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0)", idx, line)
	}
}

func TestScrollDown_EqualsScrollToBottom_WhenOvershooting(t *testing.T) {
	t.Parallel()

	m1 := uniform(10, 5, 1, 10)
	m1.ScrollToBottom()
	idx1, line1 := m1.Offset()

	m2 := uniform(10, 5, 1, 10)
	m2.ScrollBy(10000)
	idx2, line2 := m2.Offset()

	if idx1 != idx2 || line1 != line2 {
		t.Errorf("ScrollToBottom=(%d,%d) != ScrollBy(10000)=(%d,%d)", idx1, line1, idx2, line2)
	}
}

func TestScrollDown_OneLine_Repeatedly_ReachesBottom(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)

	for range 200 {
		if m.AtBottom() {
			break
		}
		m.ScrollBy(1)
	}

	if !m.AtBottom() {
		idx, line := m.Offset()
		t.Errorf("never reached bottom after 200 single-line scrolls, at (%d, %d)", idx, line)
	}
}

func TestScrollUp_OneLine_Repeatedly_ReachesTop(t *testing.T) {
	t.Parallel()

	m := uniform(5, 3, 1, 10)
	m.ScrollToBottom()

	for range 200 {
		idx, line := m.Offset()
		if idx == 0 && line == 0 {
			break
		}
		m.ScrollBy(-1)
	}

	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("never reached top after 200 single-line scrolls, at (%d, %d)", idx, line)
	}
}

// --- Edge: content smaller than viewport ---

func TestContentSmallerThanViewport(t *testing.T) {
	t.Parallel()

	// 1 item of height 3, viewport 20
	m := uniform(1, 3, 0, 20)

	if !m.AtBottom() {
		t.Error("AtBottom = false when content fits in viewport")
	}

	m.ScrollBy(100)
	idx, line := m.Offset()
	if idx != 0 || line != 0 {
		t.Errorf("Offset = (%d, %d), want (0, 0) — can't scroll when content fits", idx, line)
	}
}

// --- Edge: single item taller than viewport ---

func TestSingleTallItem_ScrollsLineByLine(t *testing.T) {
	t.Parallel()

	m := uniform(1, 50, 0, 10)
	m.ScrollBy(5)

	idx, line := m.Offset()
	if idx != 0 || line != 5 {
		t.Errorf("Offset = (%d, %d), want (0, 5)", idx, line)
	}

	m.ScrollToBottom()
	idx, line = m.Offset()
	if idx != 0 || line != 40 {
		t.Errorf("Offset = (%d, %d), want (0, 40)", idx, line)
	}
}
