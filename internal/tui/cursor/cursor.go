package cursor

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Marker is a special string that components can embed in their View() output
// to indicate where the cursor should be positioned.
const Marker = "\x1b]1337;marker\x07"

// Extract finds the cursor marker in a view string, calculates its position,
// and returns the cleaned view (without marker) and the cursor position.
// Returns nil cursor if no marker is found.
func Extract(view string) (cleanView string, cursor *tea.Cursor) {
	idx := strings.Index(view, Marker)
	if idx == -1 {
		return view, nil
	}

	beforeMarker := view[:idx]
	lines := strings.Split(beforeMarker, "\n")

	y := len(lines) - 1
	x := countVisibleChars(lines[len(lines)-1])

	cleanView = strings.Replace(view, Marker, "", 1)

	return cleanView, tea.NewCursor(x, y)
}

// Insert inserts the cursor marker at the given position (x, y) in a view string.
func Insert(view string, x, y int) string {
	lines := strings.Split(view, "\n")
	if y < 0 || y >= len(lines) {
		return view
	}

	line := lines[y]
	bytePos := findVisibleCharPos(line, x)
	if bytePos < 0 {
		return view
	}

	lines[y] = line[:bytePos] + Marker + line[bytePos:]
	return strings.Join(lines, "\n")
}

func countVisibleChars(s string) int {
	return ansi.StringWidth(s)
}

func findVisibleCharPos(s string, n int) int {
	if n < 0 {
		return -1
	}

	visibleCount := 0
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

		if visibleCount == n {
			return i
		}

		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		visibleCount++
	}

	if visibleCount == n {
		return len(s)
	}

	return -1
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
