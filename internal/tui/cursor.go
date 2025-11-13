package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// CursorMarker is a special string that pages can embed in their View() output
// to indicate where the cursor should be positioned.
// The marker will be automatically detected and removed before rendering.
const CursorMarker = "\x00CURSOR\x00"

// ExtractCursor finds the cursor marker in a view string, calculates its position,
// and returns the cleaned view (without marker) and the cursor position.
// Returns nil cursor if no marker is found.
func ExtractCursor(view string) (cleanView string, cursor *tea.Cursor) {
	idx := strings.Index(view, CursorMarker)
	if idx == -1 {
		// No marker found
		return view, nil
	}

	// Calculate cursor position from marker location
	beforeMarker := view[:idx]
	lines := strings.Split(beforeMarker, "\n")

	y := len(lines) - 1
	x := len(lines[len(lines)-1])

	// Strip the marker from the view
	cleanView = strings.Replace(view, CursorMarker, "", 1)

	return cleanView, tea.NewCursor(x, y)
}
