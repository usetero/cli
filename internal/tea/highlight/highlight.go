// Package highlight provides text selection and highlighting utilities.
package highlight

import (
	"image"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// Apply applies highlighting to a region of rendered ANSI content.
func Apply(content string, area image.Rectangle, startLine, startCol, endLine, endCol int, h Highlighter) string {
	buf := Buffer(content, area, startLine, startCol, endLine, endCol, h)
	if buf == nil {
		return content
	}
	return buf.Render()
}

// Buffer applies highlighting to a region and returns the screen buffer.
func Buffer(content string, area image.Rectangle, startLine, startCol, endLine, endCol int, h Highlighter) *uv.ScreenBuffer {
	content = normalizeWhitespace(content)

	if startLine < 0 || startCol < 0 {
		return nil
	}

	if h == nil {
		h = Default
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

		colStart := 0
		if y == startLine {
			colStart = min(startCol, len(line))
		}

		colEnd := len(line)
		if y == endLine {
			colEnd = min(endCol, len(line))
		}

		// Find last non-empty position in range
		lastContentX := -1
		for x := colStart; x < colEnd; x++ {
			cell := line.At(x)
			if cell != nil && cell.Content != "" && cell.Content != " " {
				lastContentX = x
			}
		}

		// Only highlight up to last content position
		highlightEnd := colEnd
		if lastContentX >= 0 {
			highlightEnd = lastContentX + 1
		} else if lastContentX == -1 {
			highlightEnd = colStart
		}

		for x := colStart; x < highlightEnd; x++ {
			if !image.Pt(x, y).In(area) {
				continue
			}
			cell := line.At(x)
			if cell != nil {
				line.Set(x, h(x, y, cell))
			}
		}
	}

	return &buf
}

// Extract returns plain text from a region of rendered content.
func Extract(content string, area image.Rectangle, startLine, startCol, endLine, endCol int) string {
	var sb strings.Builder
	pos := image.Pt(-1, -1)

	Buffer(content, area, startLine, startCol, endLine, endCol, func(x, y int, c *uv.Cell) *uv.Cell {
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

func normalizeWhitespace(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}
