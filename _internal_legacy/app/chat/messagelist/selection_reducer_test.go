package messagelist

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestReduceSelectionClick(t *testing.T) {
	t.Parallel()

	t.Run("non-left click ignored", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDownBlock: -1, mouseDragBlock: -1}
		next, d := reduceSelectionClick(start, tea.MouseRight, selectionPoint{block: 2, x: 4, y: 5}, true)
		if next != start {
			t.Fatalf("state changed unexpectedly: %+v -> %+v", start, next)
		}
		if d.setFocusIdx {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("miss click ignored", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDownBlock: -1, mouseDragBlock: -1}
		next, d := reduceSelectionClick(start, tea.MouseLeft, selectionPoint{block: -1, x: 4, y: 5}, false)
		if next != start {
			t.Fatalf("state changed unexpectedly: %+v -> %+v", start, next)
		}
		if d.setFocusIdx {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("left click sets anchor and focus", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDownBlock: -1, mouseDragBlock: -1}
		next, d := reduceSelectionClick(start, tea.MouseLeft, selectionPoint{block: 3, x: 7, y: 9}, true)
		if !next.mouseDown || next.mouseDownBlock != 3 || next.mouseDownX != 7 || next.mouseDownY != 9 {
			t.Fatalf("unexpected down state: %+v", next)
		}
		if next.mouseDragBlock != 3 || next.mouseDragX != 7 || next.mouseDragY != 9 {
			t.Fatalf("unexpected drag state: %+v", next)
		}
		if !d.setFocusIdx || d.focusIdx != 3 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})
}

func TestReduceSelectionMotion(t *testing.T) {
	t.Parallel()

	t.Run("ignored when not dragging", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDown: false, mouseDragBlock: 1}
		next := reduceSelectionMotion(start, tea.MouseLeft, selectionPoint{block: 2, x: 6, y: 8}, true)
		if next != start {
			t.Fatalf("state changed unexpectedly: %+v -> %+v", start, next)
		}
	})

	t.Run("ignored when no hit", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDown: true, mouseDragBlock: 1}
		next := reduceSelectionMotion(start, tea.MouseLeft, selectionPoint{block: -1, x: 6, y: 8}, false)
		if next != start {
			t.Fatalf("state changed unexpectedly: %+v -> %+v", start, next)
		}
	})

	t.Run("updates drag cursor", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDown: true, mouseDragBlock: 1, mouseDragX: 2, mouseDragY: 3}
		next := reduceSelectionMotion(start, tea.MouseLeft, selectionPoint{block: 4, x: 10, y: 11}, true)
		if next.mouseDragBlock != 4 || next.mouseDragX != 10 || next.mouseDragY != 11 {
			t.Fatalf("unexpected drag state: %+v", next)
		}
	})
}

func TestReduceSelectionRelease(t *testing.T) {
	t.Parallel()

	t.Run("ignored when not dragging", func(t *testing.T) {
		t.Parallel()
		start := selectionState{mouseDown: false}
		next, d := reduceSelectionRelease(start)
		if next != start {
			t.Fatalf("state changed unexpectedly: %+v -> %+v", start, next)
		}
		if d.handle {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})

	t.Run("release clears mouseDown and preserves click anchor", func(t *testing.T) {
		t.Parallel()
		start := selectionState{
			mouseDown:      true,
			mouseDownBlock: 5,
			mouseDownY:     12,
			mouseDragBlock: 7,
		}
		next, d := reduceSelectionRelease(start)
		if next.mouseDown {
			t.Fatalf("mouseDown should be false: %+v", next)
		}
		if !d.handle || d.clickBlock != 5 || d.clickY != 12 {
			t.Fatalf("unexpected decision: %+v", d)
		}
	})
}

func TestReduceReleaseAction(t *testing.T) {
	t.Parallel()

	t.Run("no highlight becomes click", func(t *testing.T) {
		t.Parallel()
		got := reduceReleaseAction(false, "")
		if got != releaseActionClick {
			t.Fatalf("action=%v, want %v", got, releaseActionClick)
		}
	})

	t.Run("empty extracted highlight becomes click", func(t *testing.T) {
		t.Parallel()
		got := reduceReleaseAction(true, "")
		if got != releaseActionClick {
			t.Fatalf("action=%v, want %v", got, releaseActionClick)
		}
	})

	t.Run("non-empty highlight becomes copy", func(t *testing.T) {
		t.Parallel()
		got := reduceReleaseAction(true, "hello")
		if got != releaseActionCopy {
			t.Fatalf("action=%v, want %v", got, releaseActionCopy)
		}
	})
}
