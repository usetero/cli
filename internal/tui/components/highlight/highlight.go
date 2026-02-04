package highlight

import (
	"image"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// Highlighter represents a function that defines how to highlight text.
type Highlighter func(x, y int, c *uv.Cell) *uv.Cell

// DefaultHighlighter applies inverse style (reverse video) to cells.
var DefaultHighlighter Highlighter = func(x, y int, c *uv.Cell) *uv.Cell {
	if c == nil {
		return c
	}
	c.Style.Attrs |= uv.AttrReverse
	return c
}

// SelectionHighlighter creates a highlighter using selection colors.
func SelectionHighlighter(bg, fg color.Color) Highlighter {
	return func(x, y int, c *uv.Cell) *uv.Cell {
		if c == nil {
			return c
		}
		c.Style.Bg = bg
		c.Style.Fg = fg
		return c
	}
}

// Highlight applies highlighting to a region of rendered ANSI content.
// Returns the content with the specified region highlighted.
func Highlight(content string, area image.Rectangle, startLine, startCol, endLine, endCol int, highlighter Highlighter) string {
	buf := HighlightBuffer(content, area, startLine, startCol, endLine, endCol, highlighter)
	if buf == nil {
		return content
	}
	return buf.Render()
}

// HighlightBuffer applies highlighting to a region and returns the screen buffer.
func HighlightBuffer(content string, area image.Rectangle, startLine, startCol, endLine, endCol int, highlighter Highlighter) *uv.ScreenBuffer {
	content = normalizeSpace(content)

	if startLine < 0 || startCol < 0 {
		return nil
	}

	if highlighter == nil {
		highlighter = DefaultHighlighter
	}

	width, height := area.Dx(), area.Dy()
	buf := uv.NewScreenBuffer(width, height)
	styled := uv.NewStyledString(content)
	styled.Draw(&buf, area)

	// Treat -1 as "end of content"
	if endLine < 0 {
		endLine = height - 1
	}
	if endCol < 0 {
		endCol = width
	}

	for y := startLine; y <= endLine && y < height; y++ {
		if y >= buf.Height() {
			break
		}

		line := buf.Line(y)

		// Determine column range for this line
		colStart := 0
		if y == startLine {
			colStart = min(startCol, len(line))
		}

		colEnd := len(line)
		if y == endLine {
			colEnd = min(endCol, len(line))
		}

		// Track last non-empty position
		lastContentX := -1

		// Find last non-empty position in range
		for x := colStart; x < colEnd; x++ {
			cell := line.At(x)
			if cell == nil {
				continue
			}
			if cell.Content != "" && cell.Content != " " {
				lastContentX = x
			}
		}

		// Only highlight up to last content position
		highlightEnd := colEnd
		if lastContentX >= 0 {
			highlightEnd = lastContentX + 1
		} else if lastContentX == -1 {
			highlightEnd = colStart // No content on this line
		}

		// Apply highlight style
		for x := colStart; x < highlightEnd; x++ {
			if !image.Pt(x, y).In(area) {
				continue
			}
			cell := line.At(x)
			if cell != nil {
				line.Set(x, highlighter(x, y, cell))
			}
		}
	}

	return &buf
}

// Content extracts plain text from a highlighted region.
func Content(content string, area image.Rectangle, startLine, startCol, endLine, endCol int) string {
	var sb strings.Builder
	pos := image.Pt(-1, -1)

	HighlightBuffer(content, area, startLine, startCol, endLine, endCol, func(x, y int, c *uv.Cell) *uv.Cell {
		pos.X = x
		if pos.Y == -1 {
			pos.Y = y
		} else if y > pos.Y {
			sb.WriteString(strings.Repeat("\n", y-pos.Y))
			pos.Y = y
		}
		sb.WriteString(c.Content)
		return c
	})

	if sb.Len() > 0 {
		sb.WriteString("\n")
	}
	return sb.String()
}

// ToHighlighter converts a lipgloss.Style to a Highlighter.
func ToHighlighter(lgStyle lipgloss.Style) Highlighter {
	return func(_ int, _ int, c *uv.Cell) *uv.Cell {
		if c != nil {
			c.Style = toUVStyle(lgStyle)
		}
		return c
	}
}

// toUVStyle converts a lipgloss.Style to an ultraviolet Style.
func toUVStyle(lgStyle lipgloss.Style) uv.Style {
	var uvStyle uv.Style

	uvStyle.Fg = lgStyle.GetForeground()
	uvStyle.Bg = lgStyle.GetBackground()

	var attrs uint8
	if lgStyle.GetBold() {
		attrs |= uv.AttrBold
	}
	if lgStyle.GetItalic() {
		attrs |= uv.AttrItalic
	}
	if lgStyle.GetUnderline() {
		uvStyle.Underline = uv.UnderlineSingle
	}
	if lgStyle.GetStrikethrough() {
		attrs |= uv.AttrStrikethrough
	}
	if lgStyle.GetFaint() {
		attrs |= uv.AttrFaint
	}
	if lgStyle.GetBlink() {
		attrs |= uv.AttrBlink
	}
	if lgStyle.GetReverse() {
		attrs |= uv.AttrReverse
	}
	uvStyle.Attrs = attrs

	return uvStyle
}

// normalizeSpace normalizes whitespace in a string (replaces tabs, etc).
func normalizeSpace(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

// Item is an embeddable struct for items that support highlighting.
// Embed this in your message item types to get highlighting support.
type Item struct {
	startLine   int
	startCol    int
	endLine     int
	endCol      int
	highlighter Highlighter
}

// NewItem creates a new highlight Item with default settings.
func NewItem() Item {
	return Item{
		startLine:   -1,
		startCol:    -1,
		endLine:     -1,
		endCol:      -1,
		highlighter: DefaultHighlighter,
	}
}

// SetHighlight sets the highlight range.
func (h *Item) SetHighlight(startLine, startCol, endLine, endCol int) {
	h.startLine = startLine
	h.startCol = startCol
	h.endLine = endLine
	h.endCol = endCol
}

// SetHighlightWithOffset sets the highlight range, adjusting for a left offset
// (e.g., border + padding width).
func (h *Item) SetHighlightWithOffset(startLine, startCol, endLine, endCol, offset int) {
	h.startLine = startLine
	h.startCol = max(0, startCol-offset)
	h.endLine = endLine
	if endCol >= 0 {
		h.endCol = max(0, endCol-offset)
	} else {
		h.endCol = endCol
	}
}

// Highlight returns the current highlight range.
func (h *Item) Highlight() (startLine, startCol, endLine, endCol int) {
	return h.startLine, h.startCol, h.endLine, h.endCol
}

// ClearHighlight removes any highlight.
func (h *Item) ClearHighlight() {
	h.startLine = -1
	h.startCol = -1
	h.endLine = -1
	h.endCol = -1
}

// IsHighlighted returns true if a highlight range is set.
func (h *Item) IsHighlighted() bool {
	return h.startLine != -1 || h.endLine != -1
}

// SetHighlighter sets a custom highlighter function.
func (h *Item) SetHighlighter(highlighter Highlighter) {
	h.highlighter = highlighter
}

// RenderHighlighted applies highlighting to rendered content if a range is set.
// Pass the rendered content and its dimensions.
func (h *Item) RenderHighlighted(content string, width, height int) string {
	if !h.IsHighlighted() {
		return content
	}
	area := image.Rect(0, 0, width, height)
	return Highlight(content, area, h.startLine, h.startCol, h.endLine, h.endCol, h.highlighter)
}

// HighlightedContent extracts the plain text from the highlighted region.
func (h *Item) HighlightedContent(content string, width, height int) string {
	if !h.IsHighlighted() {
		return ""
	}
	area := image.Rect(0, 0, width, height)
	return Content(content, area, h.startLine, h.startCol, h.endLine, h.endCol)
}
