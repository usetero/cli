package cursor

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Marker is a special string that components can embed in their View() output
// to indicate where the cursor should be positioned.
//
// # Why Marker Extraction Instead of Manual Offsets?
//
// We use marker extraction rather than manual coordinate calculation because:
//
//  1. **Self-correcting**: Padding, prompts, and layout changes are automatically
//     accounted for since the marker is placed in the final rendered string.
//
//  2. **No magic numbers**: Components don't need hardcoded offsets like +1, +2
//     that must be manually synchronized with padding values.
//
//  3. **Simpler**: One extraction function vs. offset calculations in every
//     component that positions children.
//
// # How It Works
//
//  1. Components with input (textarea, textinput) get cursor position from their
//     underlying Bubbles component via .Cursor()
//
//  2. Components insert Marker at that position in their View() string before
//     applying any wrapping (padding, borders, etc.)
//
//  3. TUI calls Extract() on the final composed view to find the marker and
//     calculate screen coordinates by counting lines/characters.
//
// 4. Marker is stripped before rendering to terminal.
//
// # Alternative Approach (Not Used)
//
// The alternative is interface delegation with manual offsets (used by Crush):
//   - Define Positional interface: SetPosition(x, y int)
//   - Parent calls SetPosition() to tell child its screen coordinates
//   - Child implements Cursor() to add stored x,y to its input cursor
//   - Requires hardcoded offsets for padding that must stay in sync
//
// We tried this approach and found it brittle - needed +2 for Y offset but
// couldn't easily determine why (padding? separator? mystery line?). The magic
// numbers had to be empirically determined and would break if layout changed.
//
// The marker uses a special ANSI escape sequence that lipgloss ignores when
// calculating widths. This prevents layout corruption when the marker is
// embedded in views that go through lipgloss composition (JoinHorizontal, etc).
const Marker = "\x1b]1337;marker\x07"

// Extract finds the cursor marker in a view string, calculates its position,
// and returns the cleaned view (without marker) and the cursor position.
// Returns nil cursor if no marker is found.
// The cursor position is calculated in visible characters, skipping ANSI escape codes.
func Extract(view string) (cleanView string, cursor *tea.Cursor) {
	idx := strings.Index(view, Marker)
	if idx == -1 {
		// No marker found
		return view, nil
	}

	// Calculate cursor position from marker location
	beforeMarker := view[:idx]
	lines := strings.Split(beforeMarker, "\n")

	y := len(lines) - 1

	// Count visible characters in the last line, skipping ANSI codes
	x := countVisibleChars(lines[len(lines)-1])

	// Strip the marker from the view
	cleanView = strings.Replace(view, Marker, "", 1)

	return cleanView, tea.NewCursor(x, y)
}

// Insert inserts the cursor marker at the given position (x, y) in a view string.
// The position is in visible characters, so ANSI escape codes are skipped.
// Returns the view with the marker inserted, or the original view if position is invalid.
func Insert(view string, x, y int) string {
	lines := strings.Split(view, "\n")
	if y < 0 || y >= len(lines) {
		return view
	}

	line := lines[y]

	// Find the byte position that corresponds to the x-th visible character
	bytePos := findVisibleCharPos(line, x)
	if bytePos < 0 {
		return view
	}

	// Insert marker at byte position
	lines[y] = line[:bytePos] + Marker + line[bytePos:]
	return strings.Join(lines, "\n")
}

// countVisibleChars counts visible characters in a string, skipping ANSI escape sequences.
func countVisibleChars(s string) int {
	return ansi.StringWidth(s)
}

// findVisibleCharPos finds the byte position of the n-th visible character,
// skipping ANSI escape sequences. Uses the ansi library to properly handle
// all ANSI sequence types (CSI, OSC, SGR, etc) and UTF-8.
func findVisibleCharPos(s string, n int) int {
	if n < 0 {
		return -1
	}

	visibleCount := 0
	i := 0

	for i < len(s) {
		// Check if we're at an ANSI escape sequence
		if s[i] == '\x1b' {
			// Try to parse it as various ANSI sequence types
			remaining := s[i:]

			// CSI sequence: ESC [ ... letter
			if ansi.HasCsiPrefix(remaining) {
				// Find the end (a letter A-Z or a-z)
				j := 2 // Skip ESC [
				for j < len(remaining) && !isLetter(remaining[j]) {
					j++
				}
				if j < len(remaining) {
					j++ // Include the final letter
				}
				i += j
				continue
			}

			// OSC sequence: ESC ] ... (BEL or ST)
			if ansi.HasOscPrefix(remaining) {
				// Find BEL (\x07) or ST (ESC \)
				j := 2 // Skip ESC ]
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

			// Simple ESC sequence: ESC letter
			if i+1 < len(s) && isLetter(s[i+1]) {
				i += 2
				continue
			}

			// Unknown escape, skip just the ESC
			i++
			continue
		}

		// We're at a visible character
		if visibleCount == n {
			return i
		}

		// Move to next UTF-8 character
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		visibleCount++
	}

	// If we've counted all visible chars and n == visibleCount, return end position
	if visibleCount == n {
		return len(s)
	}

	return -1 // Position out of bounds
}

// isLetter checks if a byte is an ASCII letter
func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
