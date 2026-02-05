// Package cursor provides utilities for managing terminal cursor position
// in Bubbletea views using an invisible marker.
package cursor

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Marker is a special string that components can embed in their View() output
// to indicate where the cursor should be positioned. It renders as invisible.
const Marker = "\x1b]1337;marker\x07"

// Extract finds the cursor marker in a view string, calculates its position,
// and returns the cleaned view (without marker) and the cursor position.
// Returns nil cursor if no marker is found.
func Extract(view string) (string, *tea.Cursor) {
	idx := strings.Index(view, Marker)
	if idx == -1 {
		return view, nil
	}

	beforeMarker := view[:idx]
	lines := strings.Split(beforeMarker, "\n")

	y := len(lines) - 1
	x := ansi.StringWidth(lines[len(lines)-1])

	cleanView := strings.Replace(view, Marker, "", 1)
	return cleanView, tea.NewCursor(x, y)
}

// Insert places the cursor marker at the given visible position (x, y) in a view string.
// Coordinates are in visible characters, not bytes.
func Insert(view string, x, y int) string {
	lines := strings.Split(view, "\n")
	if y < 0 || y >= len(lines) {
		return view
	}

	line := lines[y]
	bytePos := visibleToBytePos(line, x)
	if bytePos < 0 {
		return view
	}

	lines[y] = line[:bytePos] + Marker + line[bytePos:]
	return strings.Join(lines, "\n")
}

// visibleToBytePos converts a visible character position to a byte position,
// accounting for ANSI escape sequences which are invisible.
func visibleToBytePos(s string, visiblePos int) int {
	if visiblePos < 0 {
		return -1
	}

	visible := 0
	i := 0

	for i < len(s) {
		// Skip ANSI CSI sequences (e.g., colors)
		if s[i] == '\x1b' {
			remaining := s[i:]

			if ansi.HasCsiPrefix(remaining) {
				j := 2
				for j < len(remaining) && !isLetter(remaining[j]) {
					j++
				}
				if j < len(remaining) {
					j++
				}
				i += j
				continue
			}

			if ansi.HasOscPrefix(remaining) {
				j := 2
				for j < len(remaining) {
					if remaining[j] == '\x07' {
						j++
						break
					}
					if j+1 < len(remaining) && remaining[j] == '\x1b' && remaining[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i += j
				continue
			}

			// Simple two-byte escape
			if i+1 < len(s) && isLetter(s[i+1]) {
				i += 2
				continue
			}

			i++
			continue
		}

		if visible == visiblePos {
			return i
		}

		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		visible++
	}

	if visible == visiblePos {
		return len(s)
	}

	return -1
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
