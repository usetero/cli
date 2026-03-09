// Package cursor provides terminal cursor positioning helpers using an
// invisible marker embedded in rendered view output.
package cursor

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Marker is an invisible terminal marker used to recover cursor position from
// rendered strings.
const Marker = "\x1b]1337;marker\x07"

// Extract removes the first marker from view and returns a tea cursor pointing
// at that visible position. If no marker exists, cursor is nil.
func Extract(view string) (string, *tea.Cursor) {
	idx := strings.Index(view, Marker)
	if idx == -1 {
		return view, nil
	}

	before := view[:idx]
	lines := strings.Split(before, "\n")
	y := len(lines) - 1
	x := ansi.StringWidth(lines[len(lines)-1])

	clean := strings.Replace(view, Marker, "", 1)
	return clean, tea.NewCursor(x, y)
}

// Insert places the marker at the requested visible position in the rendered
// string. Coordinates are visible columns, not byte offsets.
func Insert(view string, x, y int) string {
	lines := strings.Split(view, "\n")
	if y < 0 || y >= len(lines) {
		return view
	}

	bytePos := visibleToBytePos(lines[y], x)
	if bytePos < 0 {
		return view
	}
	lines[y] = lines[y][:bytePos] + Marker + lines[y][bytePos:]
	return strings.Join(lines, "\n")
}

func visibleToBytePos(s string, visiblePos int) int {
	if visiblePos < 0 {
		return -1
	}

	visible := 0
	i := 0
	for i < len(s) {
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
