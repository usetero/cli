package highlight

import (
	"image/color"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// Highlighter defines how to style a cell during highlighting.
type Highlighter func(x, y int, c *uv.Cell) *uv.Cell

// Default applies inverse style (reverse video) to cells.
var Default Highlighter = func(x, y int, c *uv.Cell) *uv.Cell {
	if c == nil {
		return c
	}
	c.Style.Attrs |= uv.AttrReverse
	return c
}

// WithColors creates a highlighter using specific background and foreground colors.
func WithColors(bg, fg color.Color) Highlighter {
	return func(x, y int, c *uv.Cell) *uv.Cell {
		if c == nil {
			return c
		}
		c.Style.Bg = bg
		c.Style.Fg = fg
		return c
	}
}

// FromStyle converts a lipgloss.Style to a Highlighter.
func FromStyle(s lipgloss.Style) Highlighter {
	return func(_ int, _ int, c *uv.Cell) *uv.Cell {
		if c != nil {
			c.Style = toUVStyle(s)
		}
		return c
	}
}

func toUVStyle(s lipgloss.Style) uv.Style {
	var uvs uv.Style

	uvs.Fg = s.GetForeground()
	uvs.Bg = s.GetBackground()

	var attrs uint8
	if s.GetBold() {
		attrs |= uv.AttrBold
	}
	if s.GetItalic() {
		attrs |= uv.AttrItalic
	}
	if s.GetUnderline() {
		uvs.Underline = uv.UnderlineSingle
	}
	if s.GetStrikethrough() {
		attrs |= uv.AttrStrikethrough
	}
	if s.GetFaint() {
		attrs |= uv.AttrFaint
	}
	if s.GetBlink() {
		attrs |= uv.AttrBlink
	}
	if s.GetReverse() {
		attrs |= uv.AttrReverse
	}
	uvs.Attrs = attrs

	return uvs
}
